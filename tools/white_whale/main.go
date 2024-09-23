package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"whitewhale/internal"
)

const (
	// blockchainID is the ID of the blockchain, which is used in the RPC address
	blockchainID = "mnTRdJ9SnuxH39hv4w1yE5qwsVFrA3pPw7Sa28vsPxR5ZvoPX"

	// user1 and user2 are the names of the accounts
	user1 = "user1"
	user2 = "user2"
)

var isFirstDeploy = true

func main() {
	bcID := blockchainID
	if len(os.Args) > 1 {
		bcID = os.Args[1]
	}

	// rpcAddr is the address of the RPC server
	rpcAddr := "http://127.0.0.1:9750/ext/bc/" + bcID + "/rpc"

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

	// Create client and set up accounts
	client, acc1, acc2 := createClient(log)
	// create chain service
	// it is responsible for interacting with the blockchain
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

}

func createClient(log *zap.Logger) (*internal.ChainClient, internal.AccountInfo, internal.AccountInfo) {
	client := internal.NewChainClient(300000, log)

	client.AddAccount(user1, internal.User1Mnemonic, 0, 1)
	acc1, exist := client.GetAccount(user1)
	if !exist {
		log.Fatal("account not found")
	}

	log.Info("account address", zap.String(user1, acc1.Address))

	client.AddAccount(user2, internal.User2Mnemonic, 0, 2)
	acc2, exist := client.GetAccount(user2)
	if !exist {
		log.Fatal("account not found")
	}

	log.Info("account address", zap.String(user2, acc2.Address))

	return client, acc1, acc2
}

func uploadAndInstantiate(
	chainService *internal.ChainService,
	client *internal.ChainClient,
	log *zap.Logger,
	signer internal.AccountInfo,
	msg []byte,
	filepath string,
	gasPrice uint64,
) (string, string, error) {
	if !isFirstDeploy {
		err := client.IncreaseSequence(signer.Name)
		if err != nil {
			log.Fatal("error increasing sequence", zap.Error(err))
			return "", "", err
		}
	}
	isFirstDeploy = false

	txRes, err := chainService.DeployContract(signer.Name, filepath, gasPrice)
	if err != nil {
		log.Fatal("error storing code", zap.Error(err))
		return "", "", err
	}

	rawCodeID, _, err := extractResultTxDetails(txRes)
	CodeID, err := strconv.ParseUint(rawCodeID, 10, 64)
	if err != nil {
		log.Fatal("error parsing code id", zap.Error(err))
		return "", "", err
	}

	err = client.IncreaseSequence(signer.Name)
	if err != nil {
		log.Fatal("error increasing sequence", zap.Error(err))
		return "", "", err
	}

	txRes, err = chainService.InstantiateContract(signer.Name, CodeID, msg, 2000000)
	if err != nil {
		log.Fatal("error instantiating wasm contract", zap.Error(err))
		return "", "", err
	}

	rawContractCodeID, rawContractAddress, err := extractResultTxDetails(txRes)
	if err != nil {
		log.Fatal("error extracting contract details", zap.Error(err))
		return "", "", err
	}
	log.Info(
		fmt.Sprintf("committed %s contract info: ", filepath),
		zap.String(fmt.Sprintf("%s contract_address", filepath), rawContractAddress),
		zap.String(fmt.Sprintf("%s code_id", filepath), rawContractCodeID),
	)

	return rawContractCodeID, rawContractAddress, nil
}

// extractResultTxDetails extracts contract details from the transaction
func extractResultTxDetails(deployResTx *coretypes.ResultTx) (string, string, error) {
	var (
		rawContractCodeID  string
		rawContractAddress string
		eventFound         bool
	)

	for _, event := range deployResTx.TxResult.GetEvents() {
		if event.Type == "instantiate" || event.Type == "store_code" {
			for _, attr := range event.Attributes {
				switch attr.Key {
				case "_contract_address":
					rawContractAddress = attr.Value
				case "code_id":
					rawContractCodeID = attr.Value
				}
			}
			eventFound = true
			break
		}
	}

	if !eventFound {
		return "", "", errors.New("event not found, can`t extract result tx details")
	}

	if rawContractAddress == "" && rawContractCodeID == "" {
		return "", "", errors.New("rawContractAddress and rawContractCodeID is empty")
	}

	return rawContractCodeID, rawContractAddress, nil
}
