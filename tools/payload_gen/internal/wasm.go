package internal

import (
	"encoding/hex"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"
)

// DeployContractHex - deploy contract message in hexadecimal format
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

// MsgInstantiateContractHex - instantiate contract message in hexadecimal format
func MsgInstantiateContractHex(
	client *ChainClient,
	log *zap.SugaredLogger,
	from string,
) {
	signer, exist := client.GetAccount(from)
	if !exist {
		log.Fatalf("account not found")
		return
	}

	msg := &wasm.MsgInstantiateContract{
		Sender: signer.Address,
		CodeID: 1,
		Msg:    []byte(`{"purchase_price":{"amount":"10000","denom":"stake"},"transfer_price":{"amount":"10000","denom":"stake"}}`),
		Funds:  sdk.NewCoins(sdk.NewInt64Coin(denom, 10000)),
		Label:  "testing",
	}

	txBytes, err := client.GetSignedTxBytes(from, msg, 400000)
	if err != nil {
		log.Fatalf("error getting signed tx bytes: %v", err)
		return
	}

	log.Infof("wasm.MsgInstantiateContract hex: %s", hex.EncodeToString(txBytes))
}

// MsgExecuteContractHex - execute contract message in hexadecimal format
func MsgExecuteContractHex(
	client *ChainClient,
	log *zap.SugaredLogger,
	from string,
	rawContractAddress string,
) {
	signer, exist := client.GetAccount(from)
	if !exist {
		log.Fatalf("account not found")
		return
	}

	msg := &wasm.MsgExecuteContract{
		Sender:   signer.Address,
		Contract: rawContractAddress,
		Msg:      []byte(`{"register":{"name":"cidt"}}`),
		Funds:    sdk.NewCoins(sdk.NewInt64Coin(denom, 100000)),
	}

	txBytes, err := client.GetSignedTxBytes(from, msg, 400000)
	if err != nil {
		log.Fatalf("error getting signed tx bytes: %v", err)
		return
	}

	log.Infof("wasm.MsgExecuteContractHex hex: %s", hex.EncodeToString(txBytes))
}
