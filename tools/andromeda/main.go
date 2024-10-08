package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"andromeda/internal"
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
	fmt.Println("BcID:", bcID)

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

	err = chainService.UpdateAccountSequence(acc1.Name)
	if err != nil {
		log.Fatal("error getting account info", zap.Error(err))
		return
	}

	// upload andromeda_kernel.wasm
	msgInstKernelBytes, err := json.Marshal(map[string]interface{}{
		"chain_name": internal.ChainID,
		"owner":      acc1.Address,
	})
	if err != nil {
		log.Fatal("error marshaling instantiate message", zap.Error(err))
		return
	}

	_, rawKernelContractAddress, err := uploadAndInstantiate(
		chainService,
		client,
		log,
		acc1,
		msgInstKernelBytes,
		"./artifacts/andromeda_kernel.wasm",
		4000000,
	)
	if err != nil {
		log.Fatal("error uploading andromeda_vfs.wasm", zap.Error(err))
		return
	}

	// andromeda_adodb.wasm
	msgInstBytes, err := json.Marshal(map[string]string{
		"kernel_address": rawKernelContractAddress,
		"owner":          acc1.Address,
	})
	if err != nil {
		log.Fatal("error marshaling instantiate message", zap.Error(err))
		return
	}

	_, _, err = uploadAndInstantiate(
		chainService,
		client,
		log,
		acc1,
		msgInstBytes,
		"./artifacts/andromeda_vfs.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error uploading andromeda_vfs.wasm", zap.Error(err))
		return
	}

	// andromeda_adodb.wasm
	_, _, err = uploadAndInstantiate(
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
	//
	// // andromeda_economics.wasm
	// _, _, err = uploadAndInstantiate(
	// 	chainService,
	// 	client,
	// 	log,
	// 	acc1,
	// 	msgInstBytes,
	// 	"./artifacts/andromeda_economics.wasm",
	// 	5000000,
	// )
	// if err != nil {
	// 	log.Fatal("error uploading andromeda_economics.wasm", zap.Error(err))
	// 	return
	// }
	//
	// // andromeda_cw721.wasm
	// msgInstCW721Bytes, err := json.Marshal(map[string]string{
	// 	"name":           "Example Token",
	// 	"symbol":         "ET",
	// 	"minter":         acc1.Address,
	// 	"kernel_address": rawKernelContractAddress,
	// })
	// if err != nil {
	// 	log.Fatal("error marshaling instantiate message", zap.Error(err))
	// 	return
	// }
	//
	// _, cw721Addr, err := uploadAndInstantiate(
	// 	chainService,
	// 	client,
	// 	log,
	// 	acc1,
	// 	msgInstCW721Bytes,
	// 	"./artifacts/andromeda_cw721.wasm",
	// 	5000000,
	// )
	// if err != nil {
	// 	log.Fatal("error uploading andromeda_cw721.wasm", zap.Error(err))
	// 	return
	// }
	//
	// // andromeda_auction.wasm
	// _, _, err = uploadAndInstantiate(
	// 	chainService,
	// 	client,
	// 	log,
	// 	acc1,
	// 	msgInstBytes,
	// 	"./artifacts/andromeda_auction.wasm",
	// 	5000000,
	// )
	// if err != nil {
	// 	log.Fatal("error uploading andromeda_auction.wasm", zap.Error(err))
	// 	return
	// }
	//
	// // andromeda_crowdfund.wasm
	// msgInstCrowdfundBytes, err := json.Marshal(map[string]interface{}{
	// 	"token_address":       cw721Addr,
	// 	"can_mint_after_sale": true,
	// 	"owner":               acc1.Address,
	// 	"kernel_address":      rawKernelContractAddress,
	// })
	// if err != nil {
	// 	log.Fatal("error marshaling instantiate message", zap.Error(err))
	// 	return
	// }
	// _, _, err = uploadAndInstantiate(
	// 	chainService,
	// 	client,
	// 	log,
	// 	acc1,
	// 	msgInstCrowdfundBytes,
	// 	"./artifacts/andromeda_crowdfund.wasm",
	// 	5000000,
	// )
	// if err != nil {
	// 	log.Fatal("error uploading andromeda_crowdfund.wasm", zap.Error(err))
	// 	return
	// }
	//
	// // andromeda_marketplace.wasm
	// _, _, err = uploadAndInstantiate(
	// 	chainService,
	// 	client,
	// 	log,
	// 	acc1,
	// 	msgInstBytes,
	// 	"./artifacts/andromeda_marketplace.wasm",
	// 	5000000,
	// )
	// if err != nil {
	// 	log.Fatal("error uploading andromeda_marketplace.wasm", zap.Error(err))
	// 	return
	// }
	//
	// andromeda_cw20.wasm
	msgInstCw20Bytes, err := json.Marshal(map[string]interface{}{
		"name":     "CW20 Token",
		"symbol":   "LND", // only letters A-z, "-", 3-20 characters
		"decimals": 6,
		"initial_balances": []map[string]interface{}{
			{
				"address": acc1.Address,
				"amount":  "10000000000",
			},
		},
		"mint": map[string]interface{}{
			"minter": acc1.Address,
			"cap":    "1000000000000000000",
		},
		"owner":          acc1.Address,
		"kernel_address": rawKernelContractAddress,
	})
	if err != nil {
		log.Fatal("error marshaling instantiate message", zap.Error(err))
		return
	}
	_, cw20Addr, err := uploadAndInstantiate(
		chainService,
		client,
		log,
		acc1,
		msgInstCw20Bytes,
		"./artifacts/andromeda_cw20.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error uploading andromeda_cw20.wasm", zap.Error(err))
		return
	}

	// andromeda_cw20_exchange.wasm
	msgInstCw20ExchBytes, err := json.Marshal(map[string]interface{}{
		"token_address":  cw20Addr,
		"owner":          acc1.Address,
		"kernel_address": rawKernelContractAddress,
	})
	if err != nil {
		log.Fatal("error marshaling instantiate message", zap.Error(err))
		return
	}
	_, _, err = uploadAndInstantiate(
		chainService,
		client,
		log,
		acc1,
		msgInstCw20ExchBytes,
		"./artifacts/andromeda_cw20_exchange.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error uploading andromeda_cw20_exchange.wasm", zap.Error(err))
		return
	}

	// // andromeda_cw20_staking.wasm
	// msgInstCw20StakingBytes, err := json.Marshal(map[string]interface{}{
	// 	"staking_token":  cw20Addr,
	// 	"owner":          acc1.Address,
	// 	"kernel_address": rawKernelContractAddress,
	// })
	// if err != nil {
	// 	log.Fatal("error marshaling instantiate message", zap.Error(err))
	// 	return
	// }
	// _, _, err = uploadAndInstantiate(
	// 	chainService,
	// 	client,
	// 	log,
	// 	acc1,
	// 	msgInstCw20StakingBytes,
	// 	"./artifacts/andromeda_cw20_staking.wasm",
	// 	5000000,
	// )
	// if err != nil {
	// 	log.Fatal("error uploading andromeda_cw20_staking.wasm", zap.Error(err))
	// 	return
	// }
	//
	// // andromeda_merkle_airdrop.wasm
	// msgInstAirdropBytes, err := json.Marshal(map[string]interface{}{
	// 	"asset_info": map[string]interface{}{
	// 		"native": "stake", // "cw20": "cw20_address"
	// 	},
	// 	"owner":          acc1.Address,
	// 	"kernel_address": rawKernelContractAddress,
	// })
	// if err != nil {
	// 	log.Fatal("error marshaling instantiate message", zap.Error(err))
	// 	return
	// }
	// _, _, err = uploadAndInstantiate(
	// 	chainService,
	// 	client,
	// 	log,
	// 	acc1,
	// 	msgInstAirdropBytes,
	// 	"./artifacts/andromeda_merkle_airdrop.wasm",
	// 	5000000,
	// )
	// if err != nil {
	// 	log.Fatal("error uploading andromeda_merkle_airdrop.wasm", zap.Error(err))
	// 	return
	// }
	//
	// // andromeda_lockdrop.wasm
	// msgInstLockdropBytes, err := json.Marshal(map[string]interface{}{
	// 	"init_timestamp":    time.Now().Add(time.Hour * 24 * 7).UnixMilli(),
	// 	"deposit_window":    (time.Hour * 24 * 7).Milliseconds(),
	// 	"withdrawal_window": (time.Hour * 24 * 5).Milliseconds(), // should be less than deposit_window
	// 	"incentive_token":   cw20Addr,
	// 	"native_denom":      "stake",
	// 	"owner":             acc1.Address,
	// 	"kernel_address":    rawKernelContractAddress,
	// })
	// if err != nil {
	// 	log.Fatal("error marshaling instantiate message", zap.Error(err))
	// 	return
	// }
	// _, _, err = uploadAndInstantiate(
	// 	chainService,
	// 	client,
	// 	log,
	// 	acc1,
	// 	msgInstLockdropBytes,
	// 	"./artifacts/andromeda_lockdrop.wasm",
	// 	5000000,
	// )
	// if err != nil {
	// 	log.Fatal("error uploading andromeda_lockdrop.wasm", zap.Error(err))
	// 	return
	// }

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
