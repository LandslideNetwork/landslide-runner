package internal

import (
	"bytes"
	"context"

	"github.com/ava-labs/avalanchego/utils/logging"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"go.uber.org/zap"
)

func RunKVStoreTests(rpcAddr string, log logging.Logger) {
	c, err := rpchttp.New(rpcAddr, "/websocket")
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}

	CheckTX(c, log)
	Info(c, log)
	Query(c, log)
	Commit(c, log)
}

func CheckTX(c *rpchttp.HTTP, log logging.Logger) {
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

func Info(c *rpchttp.HTTP, log logging.Logger) {
	_, err := c.NetInfo(context.Background())
	if err != nil {
		log.Fatal("error NetInfo", zap.Error(err))
		return
	}
	log.Info("NetInfo success")

	resABCI, err := c.ABCIInfo(context.Background())
	if err != nil {
		log.Fatal("error ABCIInfo", zap.Error(err))
		return
	}
	if resABCI.Response.LastBlockAppHash == nil {
		log.Fatal("ABCIInfo failed")
		return
	}
	log.Info("ABCIInfo success")

	resBc, err := c.BlockchainInfo(context.Background(), 0, 0)
	if err != nil {
		log.Fatal("error BlockchainInfo", zap.Error(err))
		return
	}
	if len(resBc.BlockMetas) == 0 {
		log.Fatal("BlockchainInfo failed")
		return
	}
	log.Info("BlockchainInfo success")
}

// Commit waits for a new block to be committed
// and then commits the next block
// It also checks the apphash and the last commit hash
// of the new block
func Commit(c *rpchttp.HTTP, log logging.Logger) {
	// get the current status
	s, err := c.Status(context.Background())
	if err != nil {
		log.Fatal("error Status", zap.Error(err))
		return
	}

	log.Info("got status", zap.Any("status", s))
	// TODO: fix c.Commit at vm
	// height := s.SyncInfo.LatestBlockHeight
	//
	// // Create a transaction
	// _, _, tx := MakeTxKV()
	//
	// _, err = c.BroadcastTxCommit(context.Background(), tx)
	// if err != nil {
	// 	log.Fatal("BroadcastTxSync error", zap.Error(err))
	// 	return
	// }
	//
	// nextHeight := height + 1
	// commit, err := c.Commit(context.Background(), &nextHeight)
	// if err != nil {
	// 	log.Fatal("error Commit", zap.Error(err))
	// 	return
	// }
	// if commit.Commit == nil {
	// 	log.Fatal("Commit failed")
	// 	return
	// }
	//
	// // get block info
	// block, err := c.Block(context.Background(), &nextHeight)
	// if err != nil {
	// 	log.Fatal("error Block", zap.Error(err))
	// 	return
	// }
	// if !(len(block.Block.Header.AppHash) > 0) {
	// 	log.Fatal("Block failed")
	// 	return
	// }
	// if !bytes.Equal(block.Block.Header.AppHash.Bytes(), commit.Header.AppHash.Bytes()) {
	// 	log.Fatal("Block failed")
	// 	return
	// }
	// if nextHeight != block.Block.Header.Height {
	// 	log.Fatal("Block height does not match")
	// 	return
	// }
	//
	// // get the previous commit
	// previousHeight := nextHeight - 1
	// commitLast, err := c.Commit(context.Background(), &previousHeight)
	// if err != nil {
	// 	log.Fatal("error Commit", zap.Error(err))
	// 	return
	// }
	// if !bytes.Equal(block.Block.LastCommitHash, commitLast.Commit.Hash()) {
	// 	log.Fatal("Commit failed")
	// 	return
	// }
	//
	// log.Info("Commit success")
}

func Query(c *rpchttp.HTTP, log logging.Logger) {
	log.Info("Querying the key value store")
	// Create a transaction
	k, v, tx := MakeTxKV()

	res, err := c.BroadcastTxCommit(context.Background(), tx)
	if err != nil {
		log.Fatal("BroadcastTxSync error", zap.Error(err))
		return
	}

	if res.CheckTx.IsErr() || res.TxResult.IsErr() {
		log.Fatal("BroadcastTxCommit transaction failed")
	}
	log.Info("BroadcastTxSync transaction success")

	abcires, err := c.ABCIQuery(context.Background(), "/key", k)
	if err != nil {
		log.Fatal("ABCIQuery failed", zap.Error(err))
		return
	}
	if abcires.Response.IsErr() {
		log.Fatal("ABCIQuery failed")
		return
	}
	if !bytes.Equal(abcires.Response.Key, k) {
		log.Fatal("ABCIQuery returned key does not match queried key")
		return
	}
	if !bytes.Equal(abcires.Response.Value, v) {
		log.Info("ABCIQuery", zap.String("value", string(abcires.Response.Value)), zap.String("expected", string(v)))
		log.Fatal("ABCIQuery returned value does not match sent value")
		return
	}
	log.Info("ABCIQuery success", zap.String("key", string(k)), zap.String("value", string(v)))
	log.Info("ABCIQuery success", zap.String("resp", string(abcires.Response.Key)), zap.String("value", string(abcires.Response.Value)))
}
