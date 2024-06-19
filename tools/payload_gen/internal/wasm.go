package internal

import (
	"encoding/hex"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	"go.uber.org/zap"
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
	msg := &wasm.MsgStoreCode{
		Sender:       signer.Address,
		WASMByteCode: nameserviceWasm,
		InstantiatePermission: &types.AccessConfig{
			Permission: types.AccessTypeEverybody,
		},
	}

	txBytes, err := client.GetSignedTxBytes(from, msg, 2000000)
	if err != nil {
		log.Fatalf("error getting signed tx bytes: %v", err)
		return
	}

	// Write hex to file
	err = os.WriteFile(filePath, []byte(hex.EncodeToString(txBytes)), 0644)
	if err != nil {
		log.Fatalf("error writing to file: %v", err)
		return
	}

	log.Infof("wasm.MsgStoreCode hex stored to %s", filePath)
}
