package main

import (
	"fmt"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"andromeda/internal"
)

const (
	// rpcAddr is the address of the RPC server
	rpcAddr = "http://127.0.0.1:9750/ext/bc/2B3qDLRgCPR8fN8BCdEDKZ3VLt4E3WE7YjmPRn2XGimPGVwi4P/rpc"

	// user1 and user2 are the names of the accounts
	user1 = "user1"
	user2 = "user2"
)

func main() {
	// Configure zap logger
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("")
	config.EncoderConfig.CallerKey = "" // Remove caller information

	log, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("can't initialize zap logger: %v", err))
	}
	defer log.Sync() // flushes buffer, if any

	c, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}

	// create blockchain client,
	// it is responsible for managing accounts and signing transactions
	client := internal.NewChainClient(300000, log)

	// set up accounts witch will be used in this example
	client.AddAccount(user1, internal.User1Mnemonic, 0, 1)
	acc1, exist := client.GetAccount(user1)
	if !exist {
		log.Fatal("account not found")
		return
	}
	log.Info("account address", zap.String(user1, acc1.Address))

	client.AddAccount(user2, internal.User2Mnemonic, 0, 2)
	acc2, exist := client.GetAccount("user2")
	if !exist {
		log.Fatal("account not found")
		return
	}
	log.Info("account address", zap.String(user2, acc2.Address))

	// create chain service
	// it is responsible for interacting with the blockchain
	chainService := internal.NewChainService(client, c, log)

	// get chain info
	chainService.Info()

	// get balances
	chainService.GetBalances(acc1.Address)
	chainService.GetBalances(acc2.Address)

	// store kernel code
	// upload andromeda_kernel.wasm
	_, err = chainService.StoreCodeKernel(user1, "./artifacts/andromeda_kernel.wasm", 4000000)
	if err != nil {
		log.Fatal("error storing kernel code", zap.Error(err))
		return
	}

	err = client.IncreaseSequence(user1)
	if err != nil {
		log.Fatal("error increasing sequence", zap.Error(err))
		return
	}

	msg := []byte(fmt.Sprintf(`{"chain_name": "%s", "owner": "%s"}`, internal.ChainID, acc1.Address))
	deployResTx, err := chainService.InstantiateContract(user1, msg, 4000000)
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
		"committed contract: ",
		zap.String("contract_address", rawContractAddress),
		zap.String("code_id", rawContractCodeID),
	)
}
