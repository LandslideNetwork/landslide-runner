package main

import (
	"fmt"
	"os"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"whitewhale/internal"
)

const (
	// blockchainID is the ID of the blockchain, which is used in the local RPC address
	blockchainID = "mnTRdJ9SnuxH39hv4w1yE5qwsVFrA3pPw7Sa28vsPxR5ZvoPX"
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

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bcID := blockchainID
	if len(os.Args) > 1 {
		bcID = os.Args[1]
	}

	// Get RPC address from environment variable
	rpcAddr := os.Getenv("RPC_ADDR")
	if rpcAddr == "" {
		rpcAddr = "http://127.0.0.1:9750/ext/bc/" + bcID + "/rpc"
	}

	c, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}

	// Create chain client and set up accounts
	client, acc1, acc2 := initChainClient(log)
	// create chain service it is responsible for interacting with the blockchain
	chainService := internal.NewChainService(client, c, log)

	// get chain info
	chainService.Info()
	chainService.GetBalances(acc1.Address)
	chainService.GetBalances(acc2.Address)
	// update account sequence
	err = chainService.UpdateAccountSequence(acc1.Name)
	if err != nil {
		log.Fatal("error getting account info", zap.Error(err))
		return
	}

	deployer := internal.NewContractDeployer(acc1, chainService, client, log)

	// whale_lair.wasm
	msgInst := map[string]string{
		"kernel_address": "sdf",
		"owner":          acc1.Address,
	}
	_, _, err = deployer.UploadAndInstantiate(
		msgInst,
		"./artifacts/whale_lair.wasm",
		5000000,
	)

	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}

	log.Info("All contracts deployed successfully")
}

func initChainClient(log *zap.Logger) (*internal.ChainClient, internal.AccountInfo, internal.AccountInfo) {
	prefix := os.Getenv("PUB_ADDRESS_PREFIX")
	chainID := os.Getenv("CHAIN_ID")
	denom := os.Getenv("GAS_DENOM")
	user1Mnemonic := os.Getenv("USER1_MNEMONIC")
	user2Mnemonic := os.Getenv("USER2_MNEMONIC")
	user1 := "user1"
	user2 := "user2"

	client := internal.NewChainClient(300000, prefix, chainID, denom, log)

	client.AddAccount(user1, user1Mnemonic, 0, 1)
	acc1, exist := client.GetAccount(user1)
	if !exist {
		log.Fatal("account not found")
	}

	log.Info("account address", zap.String(user1, acc1.Address))

	client.AddAccount(user2, user2Mnemonic, 0, 2)
	acc2, exist := client.GetAccount(user2)
	if !exist {
		log.Fatal("account not found")
	}

	log.Info("account address", zap.String(user2, acc2.Address))

	return client, acc1, acc2
}
