package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	"andromeda/internal"
)

const (
	rpcAddr = "http://127.0.0.1:9756/ext/bc/7iBaHcp9SAwHLTXXoY8RxqWm5CwXPLviJEaBuSk3ZLHQJ81bP/rpc"
	user1   = "user1"
	user2   = "user2"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("can't initialize zap logger: %v", err))
	}
	defer log.Sync() // flushes buffer, if any

	c, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}

	// get chain info
	Info(c, log)

	// create chain client
	client := internal.NewChainClient(300000, log)

	// set up accounts
	client.AddAccount(user1, internal.User1Mnemonic, 0, 1)
	acc1, exist := client.GetAccount(user1)
	if !exist {
		log.Fatal("account not found")
		return
	}
	log.Info("user1 address", zap.String("acc1.Address", acc1.Address))

	client.AddAccount(user2, internal.User2Mnemonic, 0, 2)
	acc2, exist := client.GetAccount("user2")
	if !exist {
		log.Fatal("account not found")
		return
	}
	log.Info("user2 address", zap.String("acc2.Address", acc2.Address))

	// get balances
	GetBalances(c, log, acc1.Address)
	GetBalances(c, log, acc2.Address)

	//
	// upload andromeda_kernel.wasm
	//
	WASMByteCode, err := os.ReadFile("./artifacts/andromeda_kernel.wasm")
	if err != nil {
		log.Fatal("error reading wasm file", zap.Error(err))
		return
	}

	if len(WASMByteCode) == 0 {
		log.Fatal("wasm file is empty")
		return
	}

	msgStore := &wasm.MsgStoreCode{
		Sender:       acc1.Address,
		WASMByteCode: WASMByteCode,
		InstantiatePermission: &types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
	}

	txBytes, err := client.GetSignedTxBytes(user1, msgStore, 4000000)
	if err != nil {
		log.Fatal("error getting signed tx bytes: %v", zap.Error(err))
		return
	}

	// broadcast transaction async
	log.Info("Deploying wasm contract")
	res, err := BroadCastTxAsync(c, log, txBytes)
	if err != nil {
		log.Fatal("error deploying wasm contract", zap.Error(err))
		return
	}
	log.Info("Broadcast contract txHash", zap.Any("txHash", res.Hash))

	// wait for the transaction to be committed
	deployResTx, err := WaitTx(c, log, res.Hash)
	if err != nil {
		log.Fatal("error waiting for transaction", zap.Error(err))
		return
	}
	log.Info("Deploy wasm contract check success", zap.Any("deployResTx", deployResTx.TxResult))

	//
	// instantiate andromeda_kernel.wasm
	//
	err = client.IncreaseSequence(user1)
	if err != nil {
		log.Fatal("error increasing sequence", zap.Error(err))
		return
	}

	var (
		rawContractCodeID     string
		rawContractAddress    string
		instantiateEventFound bool
	)

	// instantiate wasm contract
	msgInst := &wasm.MsgInstantiateContract{
		Sender: acc1.Address,
		CodeID: 1,
		Msg:    []byte(fmt.Sprintf(`{"chain_name": "landslide-test", "owner": "%s"}`, acc1.Address)),
		Funds:  sdk.NewCoins(sdk.NewInt64Coin("stake", 10000)),
		Label:  "testing",
	}

	txBytes, err = client.GetSignedTxBytes(user1, msgInst, 4000000)
	if err != nil {
		log.Fatal("error getting signed tx bytes", zap.Error(err))
		return
	}
	// broadcast transaction async
	log.Info("MsgInstantiateContract wasm contract")
	res, err = BroadCastTxAsync(c, log, txBytes)
	if err != nil {
		log.Fatal("error MsgInstantiateContract", zap.Error(err))
		return
	}
	log.Info("Broadcast contract txHash", zap.Any("txHash", res.Hash))

	// wait for the transaction to be committed
	deployResTx, err = WaitTx(c, log, res.Hash)
	if err != nil {
		log.Fatal("error waiting for transaction", zap.Error(err))
		return
	}
	log.Info("MsgInstantiateContract check success", zap.Any("deployResTx", deployResTx.TxResult))

	log.Info("Success! Instantiating wasm contract committed")

	for _, event := range deployResTx.TxResult.GetEvents() {
		if event.Type == "instantiate" {
			for _, attr := range event.Attributes {
				switch attr.Key {
				case "_contract_address":
					rawContractAddress = attr.Value
				case "code_id":
					rawContractCodeID = attr.Value
				}
			}
			instantiateEventFound = true
			break
		}
	}

	if !instantiateEventFound {
		log.Fatal("error instantiating wasm contract")
		return
	}

	if rawContractAddress == "" || rawContractCodeID == "" {
		log.Fatal("error instantiating wasm contract, rawContractAddress or rawContractCodeID is empty")
		return
	}

	log.Info(
		"Success! Instantiating wasm contract committed",
		zap.String("contract_address", rawContractAddress),
		zap.String("code_id", rawContractCodeID),
	)
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

// GetBalances queries the balances of an address
func GetBalances(c *rpchttp.HTTP, log *zap.Logger, address string) {
	// Define the path for the balance query
	queryPath := "/cosmos.bank.v1beta1.Query/AllBalances"

	// get balances
	req := &bank.QueryAllBalancesRequest{
		Address: address,
	}

	queryArgs, err := encoding.GetCodec(proto.Name).Marshal(req)
	if err != nil {
		log.Fatal("error marshaling request: %v", zap.Error(err))
		return
	}

	// Perform the query
	resABCIQuery, err := c.ABCIQuery(context.Background(), queryPath, queryArgs)
	if err != nil {
		log.Fatal("ABCIQuery failed", zap.Error(err))
		return
	}

	if resABCIQuery.Response.IsErr() {
		log.Fatal("ABCIQuery failed", zap.String("response", resABCIQuery.Response.Log))
		return
	}

	if resABCIQuery.Response.Code != 0 {
		log.Fatal("ABCIQuery failed", zap.String("response", resABCIQuery.Response.Log))
		return
	}

	res := &bank.QueryAllBalancesResponse{}
	err = encoding.GetCodec(proto.Name).Unmarshal(resABCIQuery.Response.Value, res)
	if err != nil {
		log.Error("Failed to unmarshal response: %v", zap.Error(err))
	}

	// Now you can use balancesResponse as needed
	log.Info("Balance query success",
		zap.String("address", address),
		zap.String("balances", res.Balances.String()),
	)
}

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
