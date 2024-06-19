package main

import (
	_ "embed"
	"fmt"

	"go.uber.org/zap"

	"payload_gen/internal"
	resp "payload_gen/internal/responses"
)

var (
	//go:embed testdata/nameservice.wasm
	nameserviceWasm []byte
)

func main() {
	// Create the logger
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("can't initialize zap logger: %v", err))
	}
	defer logger.Sync() // flushes buffer, if any
	log := logger.Sugar()

	client := internal.NewChainClient(300000, log)

	resp.PrintDecodedBalances(log)

	client.AddAccount("user1", internal.User1Mnemonic, 0, 1)
	acc1, exist := client.GetAccount("user1")
	if !exist {
		log.Fatalf("account not found")
		return
	}
	log.Infof("user1 address: %s", acc1.Address)

	client.AddAccount("user2", internal.User2Mnemonic, 0, 2)
	acc2, exist := client.GetAccount("user2")
	if !exist {
		log.Fatalf("account not found")
		return
	}
	log.Infof("user2 address: %s", acc2.Address)

	internal.GetQueryAllBalancesRequestHex(acc1.Address, log)
	internal.GetQueryAllBalancesRequestHex(acc2.Address, log)

	internal.BankSendTxHex(client, log, "user1", "user2", 5000000)
	err = client.IncreaseSequence("user1")
	if err != nil {
		log.Fatalf("error increasing sequence: %v", err)
	}

	internal.DeployContractHex(client, log, "user1", nameserviceWasm, "testdata/nameservice.wasm.hex")
}
