package internal

import (
	"errors"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
)

type ChainService struct {
	client *ChainClient
	c      *rpchttp.HTTP
	log    *zap.Logger
}

func NewChainService(client *ChainClient, c *rpchttp.HTTP, log *zap.Logger) *ChainService {
	return &ChainService{
		client: client,
		c:      c,
		log:    log,
	}
}

// DeployContract - deploy wasm contract
//
// upload "./artifacts/andromeda_kernel.wasm" 4000000
func (s *ChainService) DeployContract(signerName string, fileName string, gasPrice uint64) (*coretypes.ResultTx, error) {
	WASMByteCode, err := os.ReadFile(fileName)
	if err != nil {
		s.log.Fatal("error reading wasm file", zap.Error(err))
		return nil, err
	}

	if len(WASMByteCode) == 0 {
		s.log.Fatal("wasm file is empty")
		return nil, errors.New("wasm file is empty")
	}

	acc, ok := s.client.GetAccount(signerName)
	if !ok {
		s.log.Fatal("account not found", zap.String("signerName", signerName))
		return nil, errors.New("account not found")
	}

	msgStore := &wasm.MsgStoreCode{
		Sender:       acc.Address,
		WASMByteCode: WASMByteCode,
		InstantiatePermission: &types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
	}

	txBytes, err := s.client.GetSignedTxBytes(signerName, msgStore, gasPrice)
	if err != nil {
		s.log.Fatal("error getting signed tx bytes: %v", zap.Error(err))
		return nil, err
	}

	// broadcast transaction async
	s.log.Info("Deploying wasm contract", zap.String("fileName", fileName))
	res, err := s.BroadCastTxAsync(txBytes)
	if err != nil {
		s.log.Fatal("error deploying wasm contract", zap.Error(err))
		return nil, err
	}
	s.log.Info("Broadcast contract txHash", zap.Any("txHash", res.Hash))

	// wait for the transaction to be committed
	deployResTx, err := s.WaitTx(res.Hash)
	if err != nil {
		s.log.Fatal("error waiting for transaction", zap.Error(err))
		return nil, err
	}
	s.log.Info("Deploy wasm contract check success", zap.Any("deployResTx", deployResTx.TxResult))

	return deployResTx, nil
}

// InstantiateContract - instantiate wasm contract
func (s *ChainService) InstantiateContract(signerName string, codeID uint64, msg []byte, gasPrice uint64) (*coretypes.ResultTx, error) {
	acc, ok := s.client.GetAccount(signerName)
	if !ok {
		s.log.Fatal("account not found", zap.String("signerName", signerName))
		return nil, errors.New("account not found")
	}

	// instantiate wasm contract
	msgInst := &wasm.MsgInstantiateContract{
		Sender: acc.Address,
		CodeID: codeID,
		Label:  "testing",
		Msg:    msg,
		Funds:  sdk.NewCoins(sdk.NewInt64Coin("stake", 10000)),
	}

	txBytes, err := s.client.GetSignedTxBytes(signerName, msgInst, gasPrice)
	if err != nil {
		s.log.Fatal("error getting signed tx bytes", zap.Error(err))
		return nil, err
	}
	// broadcast transaction async
	s.log.Info("MsgInstantiateContract wasm contract")
	res, err := s.BroadCastTxAsync(txBytes)
	if err != nil {
		s.log.Fatal("error MsgInstantiateContract", zap.Error(err))
		return nil, err
	}
	s.log.Info("Broadcast contract txHash", zap.Any("txHash", res.Hash))

	// wait for the transaction to be committed
	resTx, err := s.WaitTx(res.Hash)
	if err != nil {
		s.log.Fatal("error waiting for transaction", zap.Error(err))
		return nil, err
	}
	s.log.Info("MsgInstantiateContract check success", zap.Any("deployResTx", resTx.TxResult))
	s.log.Info("Success! Instantiating wasm contract committed")

	return resTx, nil
}
