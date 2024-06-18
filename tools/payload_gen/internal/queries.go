package internal

import (
	"encoding/hex"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

// GetQueryAllBalancesRequestHex creates a QueryAllBalancesRequest protobuf message and prints the hex encoded query
func GetQueryAllBalancesRequestHex(address string, log *zap.SugaredLogger) {
	// // Create a QueryAllBalancesRequest protobuf message
	req := &bank.QueryAllBalancesRequest{
		Address: address,
	}

	codec := encoding.GetCodec(proto.Name)

	queryArgs, err := codec.Marshal(req)
	if err != nil {
		log.Fatalf("error marshaling request: %v", err)
		return
	}

	// convert to hex
	queryHex := hex.EncodeToString(queryArgs)
	// Print the hex encoded query
	log.Infof("QueryAllBalancesRequest hex encoded for address %s: %s", address, queryHex)

	// decode hex back to bytes
	decodedBytes, err := hex.DecodeString(queryHex)
	if err != nil {
		log.Fatalf("error decoding hex: %v", err)
		return
	}
	// decode back to bank.QueryAllBalancesRequest
	err = encoding.GetCodec(proto.Name).Unmarshal(decodedBytes, req)
	if err != nil {
		log.Fatalf("error unmarshaling request: %v", err)
		return
	}
}
