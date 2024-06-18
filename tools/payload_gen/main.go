package main

import (
	"encoding/hex"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func main() {
	// Create the logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("can't initialize zap logger: %v", err))
	}
	defer logger.Sync() // flushes buffer, if any
	log := logger.Sugar()

	address := "wasm1vcw0he5l9mu54zawg3h440p83ex70ccmme53ac"
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
	log.Infof("QueryAllBalancesRequest hex encoded: %s", queryHex)

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

	// Print the decoded query
	log.Infof("QueryAllBalancesRequest decoded: %v", req)
}
