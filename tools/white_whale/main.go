package main

import (
	"fmt"
	"math"
	"os"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"whitewhale/internal"
)

const (
	// blockchainID is the ID of the blockchain, which is used in the local RPC address
	blockchainID = "2CLtTpjAwFXGF96KC2XNJ5p7vH6noz4Jntujv4YV61r5ZH9xLc"
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
	// chainService.Info()
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
	whaleLairCodeID, whaleLairAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"unbonding_period": fmt.Sprintf("%d", int64(time.Hour*24*7/time.Nanosecond)), // default value is 14 days, in nanoseconds
			"growth_rate":      fmt.Sprintf("%f", math.Pow(2, 1.0/365)-1),                // this is the value when you interpolate the growth rate to 2X with 365 days of bonding
			"bonding_assets": []map[string]interface{}{
				{"native_token": map[string]string{"denom": client.GetDenom()}},
				{"native_token": map[string]string{"denom": "factory/" + acc1.Address + "/test"}},
			},
		},
		"./artifacts/whale_lair.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Whale Lair contract deployed", zap.String("address", whaleLairAddr))
	log.Info("Whale Lair contract deployed", zap.Uint64("codeID", whaleLairCodeID))

	// fee_collector.wasm
	feeCollCodeID, feeCollAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{},
		"./artifacts/fee_collector.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Fee collector contract deployed", zap.String("address", feeCollAddr))
	log.Info("Fee collector contract deployed", zap.Uint64("codeID", feeCollCodeID))

	// start time for epoch manager
	startTime := fmt.Sprintf("%d", time.Now().Add(time.Hour*24).UnixNano())

	// fee_distributor.wasm
	feeDistCodeID, feeDistAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"bonding_contract_addr": whaleLairAddr,
			"fee_collector_addr":    feeCollAddr,
			"grace_period":          "21", // default value is 21 epochs
			"distribution_asset": map[string]interface{}{
				"native_token": map[string]string{"denom": client.GetDenom()},
			},
			"epoch_config": map[string]interface{}{
				"duration":      fmt.Sprintf("%d", int64(time.Hour*24*7/time.Nanosecond)), // default value is 7 days, in nanoseconds
				"genesis_epoch": startTime,
			},
		},
		"./artifacts/fee_distributor.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Fee distributor contract deployed", zap.String("address", feeDistAddr))
	log.Info("Fee distributor contract deployed", zap.Uint64("codeID", feeDistCodeID))

	// incentive.wasm
	incentiveCodeID, err := deployer.Upload("./artifacts/incentive.wasm", 5000000)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Incentive contract deployed", zap.Uint64("codeID", incentiveCodeID))

	// incentive_factory.wasm
	incentiveFactoryCodeID, incentiveFactoryAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"fee_collector_addr":   feeCollAddr,
			"fee_distributor_addr": feeDistAddr,
			"create_flow_fee": map[string]interface{}{
				"amount": "1000000000",
				"info": map[string]interface{}{
					"native_token": map[string]string{"denom": client.GetDenom()},
				},
			},
			"max_concurrent_flows":   5,
			"incentive_code_id":      incentiveCodeID,
			"max_flow_epoch_buffer":  14,
			"min_unbonding_duration": 86400,
			"max_unbonding_duration": 31556926,
		},
		"./artifacts/incentive_factory.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error instantiating contract", zap.Error(err))
		return
	}
	log.Info("Incentive factory contract deployed", zap.String("address", incentiveFactoryAddr))
	log.Info("Incentive factory contract deployed", zap.Uint64("codeID", incentiveFactoryCodeID))

	// frontend_helper.wasm
	frontendHelperCodeID, frontendHelperAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"incentive_factory": incentiveFactoryAddr,
		},
		"./artifacts/frontend_helper.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Frontend Helper contract deployed", zap.String("address", frontendHelperAddr))
	log.Info("Frontend Helper contract deployed", zap.Uint64("codeID", frontendHelperCodeID))

	// terra swap_token.wasm
	tokenCodeID, tokenAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"name":     "TerraSwap Token",
			"symbol":   "LND",
			"decimals": 6,
			"initial_balances": []map[string]string{
				{
					"address": acc1.Address,
					"amount":  "10000000000000000",
				},
			},
			"mint": map[string]string{
				"minter": acc1.Address,
				"cap":    "10000000000000000000000",
			},
		},
		"./artifacts/terraswap_token.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Token contract deployed", zap.String("address", tokenAddr))
	log.Info("Token contract deployed", zap.Uint64("codeID", tokenCodeID))

	// vault.wasm
	vaultCodeID, err := deployer.Upload("./artifacts/vault.wasm", 5000000)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Vault contract deployed", zap.Uint64("codeID", vaultCodeID))

	// vault_factory.wasm
	vaultFactoryCodeID, vaultFactoryAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"owner":              acc1.Address,
			"vault_id":           vaultCodeID,
			"token_id":           tokenCodeID,
			"fee_collector_addr": feeCollAddr,
		},
		"./artifacts/vault_factory.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Vault factory contract deployed", zap.String("address", vaultFactoryAddr))
	log.Info("Vault factory contract deployed", zap.Uint64("codeID", vaultFactoryCodeID))

	vaultRouterCodeID, vaultRouterAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"owner":              acc1.Address,
			"vault_factory_addr": vaultFactoryAddr,
		},
		"./artifacts/vault_router.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Vault router contract deployed", zap.String("address", vaultRouterAddr))
	log.Info("Vault router contract deployed", zap.Uint64("codeID", vaultRouterCodeID))

	// terraswap_pair.wasm
	pairCodeID, err := deployer.Upload(
		"./artifacts/terraswap_pair.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Pair contract deployed", zap.Uint64("codeID", pairCodeID))

	// stableswap_3pool.wasm
	stableSwapCodeID, err := deployer.Upload(
		"./artifacts/stableswap_3pool.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("StableSwap contract deployed", zap.Uint64("codeID", stableSwapCodeID))

	// terraswap_factory.wasm
	factoryCodeID, factoryAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"pair_code_id":       pairCodeID,
			"trio_code_id":       stableSwapCodeID,
			"token_code_id":      tokenCodeID,
			"fee_collector_addr": feeCollAddr,
		},
		"./artifacts/terraswap_factory.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Factory contract deployed", zap.String("address", factoryAddr))
	log.Info("Factory contract deployed", zap.Uint64("codeID", factoryCodeID))

	// terraswap_router.wasm
	routerCodeID, routerAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"terraswap_factory": factoryAddr,
		},
		"./artifacts/terraswap_router.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Router contract deployed", zap.String("address", routerAddr))
	log.Info("Router contract deployed", zap.Uint64("codeID", routerCodeID))

	// epoch_manager.wasm
	epochManagerCodeID, epochManagerAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{
			"start_epoch": map[string]interface{}{
				"id":         1,
				"start_time": startTime,
			},
			"epoch_config": map[string]interface{}{
				"duration":      fmt.Sprintf("%d", int64(time.Hour*24*7/time.Nanosecond)), // default value is 7 days, in nanoseconds
				"genesis_epoch": startTime,
			},
		},
		"./artifacts/epoch_manager.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Epoch Manager contract deployed", zap.String("address", epochManagerAddr))
	log.Info("Epoch Manager contract deployed", zap.Uint64("codeID", epochManagerCodeID))

	// fee_distributor_mock.wasm
	feeDistMockCodeID, feeDistMockAddr, err := deployer.UploadAndInstantiate(
		map[string]interface{}{},
		"./artifacts/fee_distributor_mock.wasm",
		5000000,
	)
	if err != nil {
		log.Fatal("error deploying contract", zap.Error(err))
		return
	}
	log.Info("Fee distributor mock contract deployed", zap.String("address", feeDistMockAddr))
	log.Info("Fee distributor mock contract deployed", zap.Uint64("codeID", feeDistMockCodeID))

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
