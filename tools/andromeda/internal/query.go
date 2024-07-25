package internal

import (
	"context"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

// GetBalances queries the balances of an address
func (s *chainService) GetBalances(address string) {
	// Define the path for the balance query
	queryPath := "/cosmos.bank.v1beta1.Query/AllBalances"

	// get balances
	req := &bank.QueryAllBalancesRequest{
		Address: address,
	}

	queryArgs, err := encoding.GetCodec(proto.Name).Marshal(req)
	if err != nil {
		s.log.Fatal("error marshaling request: %v", zap.Error(err))
		return
	}

	// Perform the query
	resABCIQuery, err := s.c.ABCIQuery(context.Background(), queryPath, queryArgs)
	if err != nil {
		s.log.Fatal("ABCIQuery failed", zap.Error(err))
		return
	}

	if resABCIQuery.Response.IsErr() {
		s.log.Fatal("ABCIQuery failed", zap.String("response", resABCIQuery.Response.Log))
		return
	}

	if resABCIQuery.Response.Code != 0 {
		s.log.Fatal("ABCIQuery failed", zap.String("response", resABCIQuery.Response.Log))
		return
	}

	res := &bank.QueryAllBalancesResponse{}
	err = encoding.GetCodec(proto.Name).Unmarshal(resABCIQuery.Response.Value, res)
	if err != nil {
		s.log.Error("Failed to unmarshal response: %v", zap.Error(err))
	}

	// Now you can use balancesResponse as needed
	s.log.Info("Balance query success",
		zap.String("address", address),
		zap.String("balances", res.Balances.String()),
	)
}
