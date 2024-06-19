package internal

import (
	"context"
	"encoding/hex"

	"github.com/ava-labs/avalanchego/utils/logging"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"go.uber.org/zap"
)

var decodedBalances = map[string]string{
	"0a130a057374616b65120a3130303030303030303012021001": "balances:<denom:\"stake\" amount:\"1000000000\" > pagination:<total:1 >",
	"1200": "pagination:<>",
	"0a120a057374616b65120939393439303030303012021001": "balances:<denom:\"stake\" amount:\"994900000\" > pagination:<total:1 >",
	"0a100a057374616b6512073530303030303012021001":     "balances:<denom:\"stake\" amount:\"5000000\" > pagination:<total:1 >",
}

func GetBalances(c *rpchttp.HTTP, log logging.Logger, address, querystring string) {
	// Define the path for the balance query
	queryPath := "/cosmos.bank.v1beta1.Query/AllBalances"

	reqBytes, err := hex.DecodeString(querystring)
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
	// encode the balance to hex
	balanceHex := hex.EncodeToString(resABCIQuery.Response.Value)
	// Print the balance from decodedBalances if exists
	if decoded, ok := decodedBalances[balanceHex]; ok {
		log.Info("Balance query success", zap.String("address", address), zap.String("balance", decoded))
		return
	} else {
		log.Error("Balance query failed", zap.String("address", address), zap.String("balance", balanceHex))
	}
}
