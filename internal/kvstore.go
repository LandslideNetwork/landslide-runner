package internal

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/ava-labs/avalanchego/utils/logging"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"go.uber.org/zap"
)

// RunKVStoreTests runs the key value store tests
func RunKVStoreTests(rpcAddr string, log logging.Logger) {
	c, err := rpchttp.NewWithTimeout(rpcAddr, "/websocket", 300) // Increased timeout to 300 seconds
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}
	<-time.After(2 * time.Second) // wait for first block to be committed

	CheckTX(c, log)
	Info(c, log)
	Query(c, log)
	Commit(c, log)

	GenerateTXSAsync(c, log, 200)
}

// GenerateTXSAsync generates num transactions asynchronously
// and waits for them to be committed
func GenerateTXSAsync(c *rpchttp.HTTP, log logging.Logger, num int) {
	type KV struct {
		k []byte
		v []byte
	}

	kvs := make([]KV, num)

	for i := 0; i < num; i++ {
		// Create a transaction
		k, v, tx := MakeTxKV()

		res, err := c.BroadcastTxAsync(context.Background(), tx)
		if err != nil {
			log.Fatal("BroadcastTxAsync error", zap.Error(err))
			return
		}

		if res.Code != 0 {
			log.Fatal("BroadcastTxAsync transaction failed", zap.Uint32("code", res.Code))
			return
		}

		// store the key value pair
		kvs[i] = KV{k, v}

		// wait for 100 milliseconds
		<-time.After(100 * time.Millisecond)
	}
	// wait for 15 seconds to let the transactions be committed
	<-time.After(15 * time.Second)

	// 30 attempts to query the key value store with delay of 5 seconds
	for j := 0; j < 30; j++ {
		if len(kvs) == 0 {
			log.Info("All transactions are committed")
			break
		}

		for i := 0; i < len(kvs); i++ {
			err := ABCIQuery(c, log, kvs[i].k, kvs[i].v)
			if err != nil {
				// wait for 5 seconds for block acceptance
				<-time.After(5 * time.Second)
				break
			}
			// remove the key value pair
			kvs = append(kvs[:i], kvs[i+1:]...)
			i--
		}
	}
}

// ABCIQuery queries the key value store
func ABCIQuery(c *rpchttp.HTTP, log logging.Logger, k, v []byte) error {
	abcires, err := c.ABCIQuery(context.Background(), "/key", k)
	if err != nil {
		log.Fatal("ABCIQuery failed", zap.Error(err))
		return err
	}
	if abcires.Response.IsErr() {
		log.Fatal("ABCIQuery failed")
		return errors.New("ABCIQuery failed")
	}
	if !bytes.Equal(abcires.Response.Key, k) {
		log.Fatal("ABCIQuery returned key does not match queried key")
		return errors.New("ABCIQuery returned key does not match queried key")
	}
	if !bytes.Equal(abcires.Response.Value, v) {
		log.Info("ABCIQuery", zap.String("value", string(abcires.Response.Value)), zap.String("expected", string(v)))
		log.Fatal("ABCIQuery returned value does not match sent value")
		return errors.New("ABCIQuery returned value does not match sent value")
	}
	log.Info("ABCIQuery success", zap.String("resp", string(abcires.Response.Key)), zap.String("value", string(abcires.Response.Value)))

	return nil
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
	res, err := c.NetInfo(context.Background())
	if err != nil {
		log.Fatal("error NetInfo", zap.Error(err))
		return
	}
	log.Info("NetInfo success", zap.Any("res", res))

	resABCI, err := c.ABCIInfo(context.Background())
	if err != nil {
		log.Fatal("error ABCIInfo", zap.Error(err))
		return
	}
	if resABCI.Response.LastBlockAppHash == nil {
		log.Fatal("ABCIInfo failed")
		return
	}
	log.Info("ABCIInfo success", zap.Any("res", resABCI))

	resBc, err := c.BlockchainInfo(context.Background(), 0, 0)
	if err != nil {
		log.Fatal("error BlockchainInfo", zap.Error(err))
		return
	}
	if len(resBc.BlockMetas) == 0 {
		log.Fatal("BlockchainInfo failed")
		return
	}
	log.Info("BlockchainInfo success", zap.Any("res", resBc))
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

	height := s.SyncInfo.LatestBlockHeight
	nextHeight := int64(0)
	for i := 0; i < 2; i++ {
		log.Info("BroadcastTxSync")
		start := time.Now() // Capture the start time

		// Create a transaction
		_, _, tx := MakeTxKV()

		res, err := c.BroadcastTxCommit(context.Background(), tx)
		if err != nil {
			log.Error("BroadcastTxCommit error", zap.Error(err))
			elapsed := time.Since(start)
			log.Info("Time spent on iteration", zap.Int("iteration", i), zap.Duration("elapsed", elapsed))
			return
		}

		log.Info("BroadcastTxCommit transaction success", zap.Any("res", res))

		elapsed := time.Since(start) // Calculate the elapsed time
		log.Info("Time spent on iteration", zap.Int("iteration", i), zap.Duration("elapsed", elapsed))

		nextHeight = height + int64(i+1)
		block, err := c.BlockResults(context.Background(), &nextHeight)
		if err != nil {
			log.Fatal("error BlockResults", zap.Error(err))
			return
		}
		log.Info("BlockResults success", zap.Any("block", block))
	}

	commit, err := c.Commit(context.Background(), &nextHeight)
	if err != nil {
		log.Fatal("error Commit", zap.Error(err))
		return
	}
	if commit.Commit == nil {
		log.Fatal("Commit failed")
		return
	}

	// get block info
	block, err := c.Block(context.Background(), &nextHeight)
	if err != nil {
		log.Fatal("error Block", zap.Error(err))
		return
	}
	if !(len(block.Block.Header.AppHash) > 0) {
		log.Fatal("Block failed")
		return
	}
	if !bytes.Equal(block.Block.Header.AppHash.Bytes(), commit.Header.AppHash.Bytes()) {
		log.Fatal("Block failed")
		return
	}
	if nextHeight != block.Block.Header.Height {
		log.Fatal("Block height does not match")
		return
	}

	// get the previous commit
	previousHeight := nextHeight - 1
	commitLast, err := c.Commit(context.Background(), &previousHeight)
	if err != nil {
		log.Fatal("error Commit", zap.Error(err))
		return
	}
	if !bytes.Equal(block.Block.LastCommitHash, commitLast.Commit.Hash()) {
		log.Fatal("Commit failed")
		return
	}

	log.Info("Commit success")
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
