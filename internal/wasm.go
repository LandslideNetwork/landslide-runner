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
	for i := 0; i < 10; i++ {
		log.Info("waiting for transaction to be committed")

		<-time.After(15 * time.Second)
		deployResultTx, err := c.Tx(context.Background(), res.Hash, false)
		if err != nil {
			log.Fatal("error getting transaction", zap.Error(err))
			continue
		}

		if deployResultTx.TxResult.Code != 0 {
			log.Fatal("error deploying wasm contract", zap.Error(err))
		}
	}

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
