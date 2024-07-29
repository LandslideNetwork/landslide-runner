package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"andromeda/internal"
)

const (
	// blockchainID is the ID of the blockchain, which is used in the RPC address
	blockchainID = "2RtUeU61RwkYFKULcWjLWR36FKvtUdkDpFmMd24mXuKXr48Hxz"

	// rpcAddr is the address of the RPC server
	rpcAddr = "http://127.0.0.1:9750/ext/bc/" + blockchainID + "/rpc"

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
	txRes, err := chainService.DeployContract(user1, "./artifacts/andromeda_kernel.wasm", 4000000)
	if err != nil {
		log.Fatal("error storing kernel code", zap.Error(err))
		return
	}

	rawKernelCodeID, _, err := extractResultTxDetails(txRes)
	kernelCodeId, err := strconv.ParseUint(rawKernelCodeID, 10, 64)
	if err != nil {
		log.Fatal("error parsing kernel code id", zap.Error(err))
		return
	}

	err = client.IncreaseSequence(user1)
	if err != nil {
		log.Fatal("error increasing sequence", zap.Error(err))
		return
	}

	msg := []byte(fmt.Sprintf(`{"chain_name": "%s", "owner": "%s"}`, internal.ChainID, acc1.Address))
	txRes, err = chainService.InstantiateContract(user1, kernelCodeId, msg, 2000000)
	if err != nil {
		log.Fatal("error instantiating wasm contract", zap.Error(err))
		return
	}

	rawContractCodeID, rawKernelContractAddress, err := extractResultTxDetails(txRes)
	if err != nil {
		log.Fatal("error extracting contract details", zap.Error(err))
		return
	}
	log.Info(
		"committed contract info: ",
		zap.String("contract_address", rawKernelContractAddress),
		zap.String("code_id", rawContractCodeID),
	)

	// store vfs code
	// upload andromeda_vfs.wasm
	err = client.IncreaseSequence(user1)
	if err != nil {
		log.Fatal("error increasing sequence", zap.Error(err))
		return
	}

	txRes, err = chainService.DeployContract(user1, "./artifacts/andromeda_vfs.wasm", 5000000)
	if err != nil {
		log.Fatal("error storing andromeda_vfs code", zap.Error(err))
		return
	}

	rawVFSCodeID, _, err := extractResultTxDetails(txRes)
	VFSCodeID, err := strconv.ParseUint(rawVFSCodeID, 10, 64)
	if err != nil {
		log.Fatal("error parsing kernel code id", zap.Error(err))
		return
	}

	type instantiateVFS struct {
		KernelAddress string `json:"kernel_address"`
		Owner         string `json:"owner"`
	}

	err = client.IncreaseSequence(user1)
	if err != nil {
		log.Fatal("error increasing sequence", zap.Error(err))
		return
	}

	var msgInst = instantiateVFS{
		KernelAddress: rawKernelContractAddress,
		Owner:         acc1.Address,
	}

	msgInstBytes, err := json.Marshal(msgInst)
	if err != nil {
		log.Fatal("error marshaling msgVFS", zap.Error(err))
		return
	}

	txRes, err = chainService.InstantiateContract(user1, VFSCodeID, msgInstBytes, 2000000)
	if err != nil {
		log.Fatal("error instantiating wasm contract", zap.Error(err))
		return
	}

	rawVFSContractCodeID, rawVFSContractAddress, err := extractResultTxDetails(txRes)
	if err != nil {
		log.Fatal("error extracting contract details", zap.Error(err))
		return
	}
	log.Info(
		"committed contract info: ",
		zap.String("contract_address", rawVFSContractAddress),
		zap.String("code_id", rawVFSContractCodeID),
	)

	// andromeda_adodb.wasm
	adodbCodeId, adodbAddr, err := uploadAndInstantiate(
		chainService,
		client,
		log,
		acc1,
		msgInstBytes,
		"./artifacts/andromeda_adodb.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error uploading andromeda_adodb.wasm", zap.Error(err))
		return
	}
	log.Info(
		"committed andromeda_adodb contract info: ",
		zap.String("contract_address", adodbAddr),
		zap.String("code_id", adodbCodeId),
	)

	// andromeda_economics.wasm
	adoEcoCodeId, adoEcoAddr, err := uploadAndInstantiate(
		chainService,
		client,
		log,
		acc1,
		msgInstBytes,
		"./artifacts/andromeda_economics.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error uploading andromeda_adodb.wasm", zap.Error(err))
		return
	}
	log.Info(
		"committed andromeda_economics contract info: ",
		zap.String("contract_address", adoEcoCodeId),
		zap.String("code_id", adoEcoAddr),
	)
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
	err := client.IncreaseSequence(signer.Name)
	if err != nil {
		log.Fatal("error increasing sequence", zap.Error(err))
		return "", "", err
	}

	txRes, err := chainService.DeployContract(signer.Name, filepath, gasPrice)
	if err != nil {
		log.Fatal("error storing andromeda_vfs code", zap.Error(err))
		return "", "", err
	}

	rawVFSCodeID, _, err := extractResultTxDetails(txRes)
	CodeID, err := strconv.ParseUint(rawVFSCodeID, 10, 64)
	if err != nil {
		log.Fatal("error parsing kernel code id", zap.Error(err))
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
		"committed contract info: ",
		zap.String("contract_address", rawContractAddress),
		zap.String("code_id", rawContractCodeID),
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
