package internal

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/ava-labs/avalanchego/utils/logging"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/rpc/core/types"
	"go.uber.org/zap"
	// bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func RunWASMTests(rpcAddrs []string, log logging.Logger, nameserviceDeployHex string) {
	<-time.After(2 * time.Second)

	c, err := rpchttp.New(rpcAddrs[0], "/websocket")
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}

	Info(c, log)

	// Define the path for the balance query
	addressU1 := "wasm1kng6sqkm0mjuh09cwz6u86f75lmeflj9h0fqhr"
	encodedQueryAllBalancesRequestU1 := "0a2b7761736d316b6e673673716b6d306d6a7568303963777a367538366637356c6d65666c6a39683066716872"
	addressU2 := "wasm1c4w4jxdkvj3ygdycdkjy98jve6w0d7257eqfx9"
	encodedQueryAllBalancesRequestU2 := "0a2b7761736d31633477346a78646b766a337967647963646b6a7939386a766536773064373235376571667839"

	GetBalances(c, log, addressU1, encodedQueryAllBalancesRequestU1)
	GetBalances(c, log, addressU2, encodedQueryAllBalancesRequestU2)

	// transfer tokens
	// bank.MsgSend hex encoded: from user1 to user2, amount 5000000
	log.Info("Sending 5000000 tokens from user1 to user2")
	// seq 9
	txSend := "0a8f010a8c010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64126c0a2b7761736d316b6e673673716b6d306d6a7568303963777a367538366637356c6d65666c6a39683066716872122b7761736d31633477346a78646b766a337967647963646b6a7939386a7665367730643732353765716678391a100a057374616b6512073530303030303012670a4e0a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2103ca7cd136cf54b73631e5c40850fc78cca77fdcac94ea4abe2768797c8bc7b71312040a02080112150a0f0a057374616b65120631303030303010c09a0c1a40b1250c76eb38e062e141b5a5dc1badad0c21150850bebb5e3ad9e3ad109dbb5f654daa40f7c46018781f35d0a2ee302d691dc5ebf4c4ea7d0059488e537b3252"
	_, err = BroadCastTxAsync(c, log, txSend)
	if err != nil {
		return
	}

	// wait for the transaction to be committed
	<-time.After(15 * time.Second)

	GetBalances(c, log, addressU1, encodedQueryAllBalancesRequestU1)
	GetBalances(c, log, addressU2, encodedQueryAllBalancesRequestU2)

	<-time.After(1 * time.Second)

	// deploy wasm contract
	log.Info("Deploying wasm contract")
	res, err := BroadCastTxAsync(c, log, nameserviceDeployHex)
	if err != nil {
		log.Fatal("error deploying wasm contract", zap.Error(err))
	}

	// wait for the transaction to be committed
	for i := 0; i < 30; i++ {
		log.Info("waiting for transaction to be committed")

		<-time.After(5 * time.Second)
		deployResultTx, err := c.Tx(context.Background(), res.Hash, false)
		if err != nil {
			log.Fatal("error getting transaction", zap.Error(err))
			continue
		}

		if deployResultTx.TxResult.Code != 0 {
			log.Fatal("error deploying wasm contract", zap.Error(err))
			break
		}

		log.Info("Success! transaction committed")
		break
	}

	// instantiate wasm contract
	log.Info("Instantiating wasm contract")
	txInstantiate := "0ae3010ae0010a282f636f736d7761736d2e7761736d2e76312e4d7367496e7374616e7469617465436f6e747261637412b3010a2b7761736d316b6e673673716b6d306d6a7568303963777a367538366637356c6d65666c6a396830667168721801220774657374696e672a697b2270757263686173655f7072696365223a7b22616d6f756e74223a223130303030222c2264656e6f6d223a227374616b65227d2c227472616e736665725f7072696365223a7b22616d6f756e74223a223130303030222c2264656e6f6d223a227374616b65227d7d320e0a057374616b651205313030303012690a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2103ca7cd136cf54b73631e5c40850fc78cca77fdcac94ea4abe2768797c8bc7b71312040a020801180212150a0f0a057374616b6512063430303030301080ea301a40ea4a22de567ec67d8210b86636c65a87868ccb6438682c7dae2de8e49fe4b1a763e523c3a5607b9e7a772e2a125e1f7eca27d90c6976f3997834a4ccf5c14ba0"
	res, err = BroadCastTxAsync(c, log, txInstantiate)
	if err != nil {
		log.Fatal("error deploying wasm contract", zap.Error(err))
	}
	var (
		rawContractCodeID     string
		rawContractAddress    string
		instantiateEventFound bool
	)

	// wait for the transaction to be committed
	for i := 0; i < 30; i++ {
		log.Info("waiting for transaction to be committed")

		<-time.After(5 * time.Second)
		instantiateResultTx, err := c.Tx(context.Background(), res.Hash, false)
		if err != nil {
			log.Fatal("error getting transaction", zap.Error(err))
			continue
		}

		if instantiateResultTx.TxResult.Code != 0 {
			log.Fatal("error Instantiating wasm contract", zap.Error(err))
			break
		}

		log.Info("Success! Instantiating wasm contract committed")

		for _, event := range instantiateResultTx.TxResult.GetEvents() {
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
			break
		}

		if rawContractAddress == "" || rawContractCodeID == "" {
			log.Fatal("error instantiating wasm contract, rawContractAddress or rawContractCodeID is empty")
			break
		}

		log.Info(
			"Success! Instantiating wasm contract committed",
			zap.String("contract_address", rawContractAddress),
			zap.String("code_id", rawContractCodeID),
		)

		break
	}

	log.Info("executing wasm contract")
	txExecuteContractHex := "0ac9010ac6010a242f636f736d7761736d2e7761736d2e76312e4d736745786563757465436f6e7472616374129d010a2b7761736d316b6e673673716b6d306d6a7568303963777a367538366637356c6d65666c6a39683066716872123f7761736d3134686a32746176713866706573647778786375343472747933686839307668756a7276636d73746c347a723374786d66767739733070686734641a1c7b227265676973746572223a7b226e616d65223a2263696474227d7d2a0f0a057374616b65120631303030303012690a500a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a2103ca7cd136cf54b73631e5c40850fc78cca77fdcac94ea4abe2768797c8bc7b71312040a020801180312150a0f0a057374616b6512063430303030301080ea301a406090971f7ed976cc802eae5942f2c505ded4cf5648f73716eaeb104b69b9153f2b880a76cf270e196fd070fec72f6f97939fe9e12056d3552e9ee5ab32709bd1"
	res, err = BroadCastTxAsync(c, log, txExecuteContractHex)
	if err != nil {
		log.Fatal("error deploying wasm contract", zap.Error(err))
	}

	// wait for the transaction to be committed
	for i := 0; i < 30; i++ {
		log.Info("waiting for transaction to be committed")

		<-time.After(5 * time.Second)
		execResultTx, err := c.Tx(context.Background(), res.Hash, false)
		if err != nil {
			log.Fatal("error getting transaction", zap.Error(err))
			continue
		}

		if execResultTx.TxResult.Code != 0 {
			log.Fatal("error executing wasm contract", zap.Error(err))
			break
		}

		log.Info("Success! executing committed")
		break
	}

	encodedQueryAllBalancesRequestU2 = "0a3f7761736d3134686a32746176713866706573647778786375343472747933686839307668756a7276636d73746c347a723374786d667677397330706867346412247b227265736f6c76655f7265636f7264223a207b226e616d65223a202263696474227d7d"
	QuerySmartContractStateRequest(c, log, rawContractAddress, encodedQueryAllBalancesRequestU2)
}

// BroadCastTxAsync - broadcast transaction async
func BroadCastTxAsync(c *rpchttp.HTTP, log logging.Logger, txHex string) (*coretypes.ResultBroadcastTx, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		log.Fatal("error decoding hex: %v", zap.Error(err))
		return nil, errors.New("error decoding hex")
	}

	res, err := c.BroadcastTxAsync(context.Background(), txBytes)
	if err != nil {
		log.Fatal("BroadcastTxAsync error", zap.Error(err))
		return nil, errors.New("BroadcastTxAsync error")
	}
	if res.Code != 0 {
		log.Fatal("BroadcastTxAsync transaction failed", zap.Uint32("code", res.Code))
		log.Fatal("error", zap.String("Log", res.Log))
		return nil, errors.New("BroadcastTxAsync transaction failed")
	}

	return res, nil
}
