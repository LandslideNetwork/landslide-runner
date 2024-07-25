package internal

import (
	"errors"
	"fmt"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
)

// StoreCodeKernel - store kernel code
//
// upload andromeda_kernel.wasm
func StoreCodeKernel(client *ChainClient, signerName string, c *rpchttp.HTTP, log *zap.Logger) (*coretypes.ResultTx, error) {
	WASMByteCode, err := os.ReadFile("./artifacts/andromeda_kernel.wasm")
	if err != nil {
		log.Fatal("error reading wasm file", zap.Error(err))
		return nil, err
	}

	if len(WASMByteCode) == 0 {
		log.Fatal("wasm file is empty")
		return nil, errors.New("wasm file is empty")
	}

	acc, ok := client.GetAccount(signerName)
	if !ok {
		log.Fatal("account not found", zap.String("signerName", signerName))
		return nil, errors.New("account not found")
	}

	msgStore := &wasm.MsgStoreCode{
		Sender:       acc.Address,
		WASMByteCode: WASMByteCode,
		InstantiatePermission: &types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
	}

	txBytes, err := client.GetSignedTxBytes(signerName, msgStore, 4000000)
	if err != nil {
		log.Fatal("error getting signed tx bytes: %v", zap.Error(err))
		return nil, err
	}

	// broadcast transaction async
	log.Info("Deploying wasm contract")
	res, err := BroadCastTxAsync(c, log, txBytes)
	if err != nil {
		log.Fatal("error deploying wasm contract", zap.Error(err))
		return nil, err
	}
	log.Info("Broadcast contract txHash", zap.Any("txHash", res.Hash))

	// wait for the transaction to be committed
	deployResTx, err := WaitTx(c, log, res.Hash)
	if err != nil {
		log.Fatal("error waiting for transaction", zap.Error(err))
		return nil, err
	}
	log.Info("Deploy wasm contract check success", zap.Any("deployResTx", deployResTx.TxResult))

	return deployResTx, nil
}

// InstantiateContract - instantiate wasm contract
func InstantiateContract(client *ChainClient, signerName string, c *rpchttp.HTTP, log *zap.Logger) (*coretypes.ResultTx, error) {
	acc, ok := client.GetAccount(signerName)
	if !ok {
		log.Fatal("account not found", zap.String("signerName", signerName))
		return nil, errors.New("account not found")
	}

	// instantiate wasm contract
	msgInst := &wasm.MsgInstantiateContract{
		Sender: acc.Address,
		CodeID: 1,
		Msg:    []byte(fmt.Sprintf(`{"chain_name": "landslide-test", "owner": "%s"}`, acc.Address)),
		Funds:  sdk.NewCoins(sdk.NewInt64Coin("stake", 10000)),
		Label:  "testing",
	}

	txBytes, err := client.GetSignedTxBytes(signerName, msgInst, 4000000)
	if err != nil {
		log.Fatal("error getting signed tx bytes", zap.Error(err))
		return nil, err
	}
	// broadcast transaction async
	log.Info("MsgInstantiateContract wasm contract")
	res, err := BroadCastTxAsync(c, log, txBytes)
	if err != nil {
		log.Fatal("error MsgInstantiateContract", zap.Error(err))
		return nil, err
	}
	log.Info("Broadcast contract txHash", zap.Any("txHash", res.Hash))

	// wait for the transaction to be committed
	deployResTx, err := WaitTx(c, log, res.Hash)
	if err != nil {
		log.Fatal("error waiting for transaction", zap.Error(err))
		return nil, err
	}
	log.Info("MsgInstantiateContract check success", zap.Any("deployResTx", deployResTx.TxResult))

	log.Info("Success! Instantiating wasm contract committed")

	return deployResTx, nil
}
