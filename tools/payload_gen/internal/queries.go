package internal

import (
	"encoding/hex"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

// GetQueryAllBalancesRequestHex creates a QueryAllBalancesRequest protobuf message and prints the hex encoded query
func GetQueryAllBalancesRequestHex(log *zap.SugaredLogger, address string) {
	req := &bank.QueryAllBalancesRequest{
		Address: address,
	}

	queryArgs, err := encoding.GetCodec(proto.Name).Marshal(req)
	if err != nil {
		log.Fatalf("error marshaling request: %v", err)
		return
	}

	log.Infof(
		"QueryAllBalancesRequest hex encoded for address %s: %s",
		address,
		hex.EncodeToString(queryArgs),
	)
}

// GetQuerySmartContractStateRequestHex creates a GetQuerySmartContractStateRequestHex protobuf message
// and prints the hex encoded query
func GetQuerySmartContractStateRequestHex(log *zap.SugaredLogger, rawContractAddress string) {
	req := &types.QuerySmartContractStateRequest{
		Address:   rawContractAddress,
		QueryData: []byte(`{"resolve_record": {"name": "cidt"}}`),
	}

	queryArgs, err := encoding.GetCodec(proto.Name).Marshal(req)
	if err != nil {
		log.Fatalf("error marshaling request: %v", err)
		return
	}

	log.Infof(
		"QuerySmartContractStateRequest hex encoded for address %s: %s",
		rawContractAddress,
		hex.EncodeToString(queryArgs),
	)
}
