package internal

import (
	"context"

	"github.com/ava-labs/avalanchego/utils/logging"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"go.uber.org/zap"
)

func CheckTX(rpcAddr string, log logging.Logger) {
	c, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}

	// Create a transaction
	k := []byte("name")
	v := []byte("satoshi")
	tx := append(k, append([]byte("="), v...)...)

	checkTx, err := c.CheckTx(context.Background(), tx)
	if err != nil {
		log.Fatal("error CheckTx", zap.Error(err))
		return
	}

	if checkTx.IsErr() || checkTx.ResponseCheckTx.IsErr() {
		log.Fatal("CheckTx transaction failed")
		return
	}

	log.Info("CheckTx transaction success")
}
