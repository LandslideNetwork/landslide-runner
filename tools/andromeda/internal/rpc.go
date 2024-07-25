package internal

import (
	"context"
	"errors"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
)

// BroadCastTxAsync - broadcast transaction async
func BroadCastTxAsync(c *rpchttp.HTTP, log *zap.Logger, txBytes []byte) (*coretypes.ResultBroadcastTx, error) {
	res, err := c.BroadcastTxAsync(context.Background(), txBytes)
	if err != nil {
		log.Fatal("BroadcastTxAsync error", zap.Error(err))
		return nil, errors.New("BroadcastTxAsync error")
	}
	if res.Code != 0 {
		log.Fatal("BroadcastTxAsync transaction failed", zap.Uint32("code", res.Code))
		log.Fatal("error", zap.String("Log", res.Log))
		return nil, errors.New("BroadcastTxAsync transaction failed")
	}

	return res, nil
}

// WaitTx - wait for transaction to be committed
func WaitTx(c *rpchttp.HTTP, log *zap.Logger, txHash []byte) (*coretypes.ResultTx, error) {
	for i := 0; i < 30; i++ {
		log.Info("Waiting for transaction to be committed")

		<-time.After(5 * time.Second)
		execResultTx, err := c.Tx(context.Background(), txHash, false)
		if err != nil {
			log.Error("Error getting transaction", zap.Error(err))
			continue
		}

		if execResultTx.TxResult.Code != 0 {
			log.Error("Error executing wasm contract", zap.Error(err))
			return nil, errors.New("error executing wasm contract")
		}

		log.Info("Success! Executing committed")
		return execResultTx, nil
	}

	return nil, errors.New("WaitTx failed")
}

func Info(c *rpchttp.HTTP, log *zap.Logger) {
	res, err := c.NetInfo(context.Background())
	if err != nil {
		log.Fatal("error NetInfo", zap.Error(err))
		return
	}
	log.Info("NetInfo success: ", zap.Any("res", res))

	resABCI, err := c.ABCIInfo(context.Background())
	if err != nil {
		log.Fatal("error ABCIInfo", zap.Error(err))
		return
	}
	if resABCI.Response.LastBlockAppHash == nil {
		log.Fatal("ABCIInfo failed")
		return
	}
	log.Info("ABCIInfo success: ", zap.Any("resABCI", resABCI))

	resBc, err := c.BlockchainInfo(context.Background(), 0, 0)
	if err != nil {
		log.Fatal("error BlockchainInfo", zap.Error(err))
		return
	}
	if len(resBc.BlockMetas) == 0 {
		log.Fatal("BlockchainInfo failed")
		return
	}
	log.Info("BlockchainInfo success: ", zap.Any("resBc", resBc))
}
