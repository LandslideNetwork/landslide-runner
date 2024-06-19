package internal

import (
	"encoding/hex"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

func DeployContractHex(
	client *ChainClient,
	log *zap.SugaredLogger,
	from string,
	nameserviceWasm []byte,
	filePath string,
) {
	signer, exist := client.GetAccount(from)
	if !exist {
		log.Fatalf("account not found")
		return
	}
	req := &wasm.MsgStoreCode{
		Sender:       signer.Address,
		WASMByteCode: nameserviceWasm,
		InstantiatePermission: &types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
	}

	codec := encoding.GetCodec(proto.Name)

	queryArgs, err := codec.Marshal(req)
	if err != nil {
		log.Fatalf("error marshaling request: %v", err)
		return
	}

	queryHex := hex.EncodeToString(queryArgs)
	// Write queryHex to file
	err = os.WriteFile(filePath, []byte(queryHex), 0644)
	if err != nil {
		log.Fatalf("error writing to file: %v", err)
		return
	}

	decodedBytes, err := hex.DecodeString(queryHex)
	if err != nil {
		log.Fatalf("error decoding hex: %v", err)
		return
	}

	err = encoding.GetCodec(proto.Name).Unmarshal(decodedBytes, req)
	if err != nil {
		log.Fatalf("error unmarshaling request: %v", err)
		return
	}

	log.Infof("wasm.MsgStoreCode hex stored to %s", filePath)
}
