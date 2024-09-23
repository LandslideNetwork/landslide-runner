package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
)

type ContractDeployer struct {
	chainService  *ChainService
	client        *ChainClient
	log           *zap.Logger
	signer        AccountInfo
	isFirstDeploy bool
}

// NewContractDeployer creates a new instance of ContractDeployer
func NewContractDeployer(signer AccountInfo, chainService *ChainService, client *ChainClient, log *zap.Logger) *ContractDeployer {
	return &ContractDeployer{
		chainService:  chainService,
		client:        client,
		signer:        signer,
		log:           log,
		isFirstDeploy: true,
	}
}

// SetAccount sets the account information for the deployer
func (cd *ContractDeployer) SetAccount(signer AccountInfo) {
	cd.signer = signer
}

// SetFirstDeploy sets the flag indicating if this is the first deployment
func (cd *ContractDeployer) SetFirstDeploy(isFirst bool) {
	cd.isFirstDeploy = isFirst
}

// IncreaseSequence increases the sequence number for the account if not the first deployment
func (cd *ContractDeployer) IncreaseSequence() error {
	if cd.isFirstDeploy {
		cd.isFirstDeploy = false
		return nil
	}

	err := cd.client.IncreaseSequence(cd.signer.Name)
	if err != nil {
		cd.log.Fatal("error increasing sequence", zap.Error(err))
		return err
	}

	return nil
}

// UploadAndInstantiate uploads and instantiates a smart contract
func (cd *ContractDeployer) UploadAndInstantiate(msg interface{}, filepath string, gasPrice uint64) (string, string, error) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		cd.log.Fatal("error marshaling instantiate message", zap.Error(err))
		return "", "", err
	}

	txRes, err := cd.chainService.DeployContract(cd.signer.Name, filepath, gasPrice)
	if err != nil {
		cd.log.Fatal("error storing code", zap.Error(err))
		return "", "", err
	}

	rawCodeID, _, err := extractResultTxDetails(txRes)
	CodeID, err := strconv.ParseUint(rawCodeID, 10, 64)
	if err != nil {
		cd.log.Fatal("error parsing code id", zap.Error(err))
		return "", "", err
	}

	err = cd.client.IncreaseSequence(cd.signer.Name)
	if err != nil {
		cd.log.Fatal("error increasing sequence", zap.Error(err))
		return "", "", err
	}

	txRes, err = cd.chainService.InstantiateContract(cd.signer.Name, CodeID, msgBytes, 2000000)
	if err != nil {
		cd.log.Fatal("error instantiating wasm contract", zap.Error(err))
		return "", "", err
	}

	rawContractCodeID, rawContractAddress, err := extractResultTxDetails(txRes)
	if err != nil {
		cd.log.Fatal("error extracting contract details", zap.Error(err))
		return "", "", err
	}

	cd.log.Info(
		fmt.Sprintf("Contract deployed: %s\n", filepath),
		zap.String(fmt.Sprintf("\ncontract_address:"), rawContractAddress),
		zap.String(fmt.Sprintf("\ncode_id:"), rawContractCodeID),
	)

	return rawContractCodeID, rawContractAddress, nil
}

// extractResultTxDetails extracts contract details from the transaction result
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
