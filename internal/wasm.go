package internal

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/ava-labs/avalanchego/utils/logging"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"

	"go.uber.org/zap"
	// bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func RunWASMTests(rpcAddrs []string, log logging.Logger) {

	<-time.After(2 * time.Second)

	c, err := rpchttp.New(rpcAddrs[0], "/websocket")
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}

	// get chain info
	res, err := c.NetInfo(context.Background())
	if err != nil {
		log.Fatal("error NetInfo", zap.Error(err))
		return
	}
	log.Info("NetInfo success", zap.Any("res", res))

	resABCI, err := c.ABCIInfo(context.Background())
	if err != nil {
		log.Fatal("error ABCIInfo", zap.Error(err))
		return
	}
	if resABCI.Response.LastBlockAppHash == nil {
		log.Fatal("ABCIInfo failed")
		return
	}
	log.Info("ABCIInfo success", zap.Any("res", resABCI))

	resBc, err := c.BlockchainInfo(context.Background(), 0, 0)
	if err != nil {
		log.Fatal("error BlockchainInfo", zap.Error(err))
		return
	}
	if len(resBc.BlockMetas) == 0 {
		log.Fatal("BlockchainInfo failed")
		return
	}
	log.Info("BlockchainInfo success", zap.Any("res", resBc))

	// Get balance

	// Define the path for the balance query
	const queryPath = "/cosmos.bank.v1beta1.Query/AllBalances"
	const address = "wasm1vcw0he5l9mu54zawg3h440p83ex70ccmme53ac"
	const encodedQueryAllBalancesRequest = "0a2b7761736d31766377306865356c396d7535347a61776733683434307038336578373063636d6d6535336163"

	// decode hex back to bytes
	reqBytes, err := hex.DecodeString(encodedQueryAllBalancesRequest)
	if err != nil {
		log.Fatal("error decoding hex: %v", zap.Error(err))
		return
	}

	// Perform the query
	resABCIQuery, err := c.ABCIQuery(context.Background(), queryPath, reqBytes)
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

	// The balance is returned in the response value
	balance := resABCIQuery.Response.Value

	log.Info("Balance query success", zap.String("address", address), zap.String("balance", string(balance)))
}
