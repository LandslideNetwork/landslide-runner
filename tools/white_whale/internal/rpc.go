package internal

import (
	"context"
	"errors"
	"strings"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
)

// BroadCastTxAsync - broadcast transaction async
func (s *ChainService) BroadCastTxAsync(txBytes []byte) (*coretypes.ResultBroadcastTx, error) {
	res, err := s.c.BroadcastTxAsync(context.Background(), txBytes)
	if err != nil {
		s.log.Fatal("BroadcastTxAsync error", zap.Error(err))
		return nil, errors.New("BroadcastTxAsync error")
	}
	if res.Code != 0 {
		s.log.Fatal("BroadcastTxAsync transaction failed", zap.Uint32("code", res.Code))
		s.log.Fatal("error", zap.String("Log", res.Log))
		return nil, errors.New("BroadcastTxAsync transaction failed")
	}

	return res, nil
}

// WaitTx - wait for transaction to be committed
func (s *ChainService) WaitTx(txHash []byte) (*coretypes.ResultTx, error) {
	s.log.Info("Waiting for transaction to be committed")
	for i := 0; i < 30; i++ {
		<-time.After(5 * time.Second)
		execResultTx, err := s.c.Tx(context.Background(), txHash, false)
		if err != nil {
			s.log.Info("waiting for transaction another 5 sek...")
			continue
		}

		if execResultTx.TxResult.Code != 0 {
			s.log.Warn("execResultTx.TxResult.Code != 0", zap.String("Log", execResultTx.TxResult.Log))
			return nil, errors.New("error executing wasm contract")
		}

		return execResultTx, nil
	}

	return nil, errors.New("WaitTx failed")
}

// Info - get chain info
func (s *ChainService) Info() {
	res, err := s.c.NetInfo(context.Background())
	if err != nil {
		if strings.Contains(err.Error(), "Status: 404 Not Found") {
			s.log.Fatal("Invalid RPC address, no connection.", zap.Error(err))
		} else {
			s.log.Fatal("error NetInfo", zap.Error(err))
		}
		return
	}
	s.log.Info("NetInfo success: ", zap.Any("res", res))

	resABCI, err := s.c.ABCIInfo(context.Background())
	if err != nil {
		s.log.Fatal("error ABCIInfo", zap.Error(err))
		return
	}
	if resABCI.Response.LastBlockAppHash == nil {
		s.log.Fatal("ABCIInfo failed")
		return
	}
	s.log.Info("ABCIInfo success: ", zap.Any("resABCI", resABCI))

	resBc, err := s.c.BlockchainInfo(context.Background(), 0, 0)
	if err != nil {
		s.log.Fatal("error BlockchainInfo", zap.Error(err))
		return
	}
	if len(resBc.BlockMetas) == 0 {
		s.log.Fatal("BlockchainInfo failed")
		return
	}

	s.log.Info("BlockchainInfo success: ", zap.Any("resBc", resBc))

	resStatus, err := s.c.Status(context.Background())
	if err != nil {
		s.log.Fatal("get status error", zap.Error(err))
		return
	}
	s.log.Info("Status success: ", zap.Any("resStatus", resStatus))
}
