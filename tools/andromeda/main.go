package main

import (
	"fmt"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"go.uber.org/zap"

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
	internal.Info(c, log)

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
	internal.GetBalances(c, log, acc1.Address)
	internal.GetBalances(c, log, acc2.Address)

	//
	// store kernel code
	// upload andromeda_kernel.wasm
	_, err = internal.StoreCodeKernel(client, user1, c, log)
	if err != nil {
		log.Fatal("error storing kernel code", zap.Error(err))
		return
	}

	//
	// instantiate andromeda_kernel.wasm
	//
	err = client.IncreaseSequence(user1)
	if err != nil {
		log.Fatal("error increasing sequence", zap.Error(err))
		return
	}

	deployResTx, err := internal.InstantiateContract(client, user1, c, log)
	if err != nil {
		log.Fatal("error instantiating wasm contract", zap.Error(err))
		return
	}
	var (
		rawContractCodeID     string
		rawContractAddress    string
		instantiateEventFound bool
	)
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
