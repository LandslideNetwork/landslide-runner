package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ava-labs/avalanchego/api/info"
	validatorstatepb "github.com/ava-labs/avalanchego/proto/pb/validatorstate"
	"github.com/ava-labs/avalanchego/snow/validators/gvalidators"
	"github.com/ava-labs/avalanchego/utils"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp/payload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"

	"github.com/ava-labs/avalanchego/utils/logging"
	bftrand "github.com/cometbft/cometbft/libs/rand"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"go.uber.org/zap"
)

const (
	WarpQuorumNumerator   = 67
	WarpQuorumDenominator = 100
)

var httpClient = http.DefaultClient

// RunKVStoreTests runs the key value store tests
func RunKVStoreTests(rpcAddrList []string, networkID uint32, chainID ids.ID, log logging.Logger) {
	c, err := rpchttp.New(rpcAddrList[0], "/websocket")
	if err != nil {
		log.Fatal("error creating client", zap.Error(err)) //nolint:gocritic
	}
	warpClients := make([]Client, len(rpcAddrList))
	for i, rpcAddr := range rpcAddrList {
		warpClients[i], err = NewClient(rpcAddr)
		if err != nil {
			log.Fatal("error creating rpc client", zap.Error(err)) //nolint:gocritic
		}
	}
	<-time.After(2 * time.Second) // wait for first block to be committed

	re := regexp.MustCompile(`([\w./:]+):([0-9]+)`)

	hostPort := re.FindString(rpcAddrList[0])

	////CheckTX(c, log)
	Info(c, log)
	Query(c, log)
	Commit(c, log)
	WARPGetMessage(warpClients[0], networkID, chainID, log)
	WARPGetMessageSignature(hostPort, warpClients[0], networkID, chainID, log)
	WarpGetMessageAggregateSignature(rpcAddrList[0], c, warpClients, networkID, chainID, log)

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

	// Create a transaction
	_, _, tx := MakeTxKV()

	_, err = c.BroadcastTxCommit(context.Background(), tx)
	if err != nil {
		log.Fatal("BroadcastTxSync error", zap.Error(err))
		return
	}

	nextHeight := height + 1
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

func WARPGetMessage(warpClient Client, networkID uint32, chainID ids.ID, log logging.Logger) {
	addressedCall, err := payload.NewAddressedCall(
		utils.RandomBytes(20),
		[]byte(bftrand.Str(24)),
	)
	if err != nil {
		log.Fatal("failed to initialize addressed call", zap.Error(err))
		return
	}

	msg, err := warp.NewUnsignedMessage(networkID, chainID, addressedCall.Bytes())
	if err != nil {
		log.Fatal("failed to create unsigned message", zap.Error(err))
		return
	}
	err = msg.Initialize()
	if err != nil {
		log.Fatal("failed to initialize unsigned message", zap.Error(err))
		return
	}
	resultAddMsg, err := warpClient.AddMessage(context.Background(), msg.Bytes())
	if err != nil {
		log.Fatal("failed to warp add message", zap.Error(err))
		return
	}
	resultGetMsg, err := warpClient.GetMessage(context.Background(), msg.ID())
	if err != nil {
		log.Fatal("failed to warp get message", zap.Error(err))
		return
	}
	resultMsg, err := warp.ParseUnsignedMessage(resultGetMsg.Message)
	if err != nil {
		log.Fatal("failed to warp get message", zap.Error(err))
		return
	}
	if msg.NetworkID != resultMsg.NetworkID {
		log.Info("warp_get_message", zap.String("value", fmt.Sprintf("%d", msg.NetworkID)), zap.String("expected", fmt.Sprintf("%d", resultMsg.NetworkID)))
		log.Fatal("warp_get_message returned value does not match sent value")
		return
	}
	if msg.SourceChainID != resultMsg.SourceChainID {
		log.Info("warp_get_message", zap.String("value", msg.SourceChainID.String()), zap.String("expected", resultMsg.SourceChainID.String()))
		log.Fatal("warp_get_message returned value does not match sent value")
		return
	}
	if !bytes.Equal(msg.Payload, resultMsg.Payload) {
		log.Info("warp_get_message", zap.String("value", string(msg.Payload)), zap.String("expected", string(resultMsg.Payload)))
		log.Fatal("warp_get_message returned value does not match sent value")
		return
	}

	log.Info("AddMessage result", zap.String("response body", resultAddMsg.MessageID))
}

func WARPGetMessageSignature(rpcURI string, warpClient Client, networkID uint32, chainID ids.ID, log logging.Logger) {
	addressedCall, err := payload.NewAddressedCall(
		utils.RandomBytes(20),
		[]byte(bftrand.Str(24)),
	)
	if err != nil {
		log.Fatal("failed to initialize addressed call", zap.Error(err))
		return
	}

	msg, err := warp.NewUnsignedMessage(networkID, chainID, addressedCall.Bytes())
	if err != nil {
		log.Fatal("failed to create unsigned message", zap.Error(err))
		return
	}
	err = msg.Initialize()
	if err != nil {
		log.Fatal("failed to initialize unsigned message", zap.Error(err))
		return
	}
	resultAddMsg, err := warpClient.AddMessage(context.Background(), msg.Bytes())
	if err != nil {
		log.Fatal("failed to warp add message", zap.Error(err))
		return
	}
	log.Info("AddMessage result", zap.String("response body", resultAddMsg.MessageID))
	resultMsgSignature, err := warpClient.GetMessageSignature(context.Background(), msg.ID())
	if err != nil {
		log.Fatal("failed to warp get message", zap.Error(err))
		return
	}
	sig, err := bls.SignatureFromBytes(resultMsgSignature.Signature)
	if err != nil {
		log.Fatal("failed to parse BLS Signature", zap.Error(err))
		return
	}
	// Note: the public key should not be fetched from the node in practice.
	//       The public key should be fetched from the P-chain directly.
	infoClient := info.NewClient(rpcURI)
	_, nodePOP, err := infoClient.GetNodeID(context.Background())
	if err != nil {
		log.Fatal("failed to fetch BLS public key", zap.String("RPC URI", rpcURI), zap.Error(err))
		return
	}

	pk := nodePOP.Key()
	if !bls.Verify(pk, sig, msg.Bytes()) {
		log.Fatal("failed to verify BLS signature against public key", zap.String("RPC URI", rpcURI))
		return
	}
	log.Info("msg signature", zap.String("signature", string(resultMsgSignature.Signature)))
}

func WarpGetMessageAggregateSignature(serverAddr string, httpClient *rpchttp.HTTP, warpClients []Client, networkID uint32, chainID ids.ID, log logging.Logger) {
	vmServerAddr, err := ioutil.ReadFile("/tmp/vm_server_address")
	if err != nil {
		panic(err)
	}
	clientConn, err := grpc.NewClient(
		"passthrough:///"+string(vmServerAddr),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	validatorStateClient := gvalidators.NewClient(validatorstatepb.NewValidatorStateClient(clientConn))
	ready := false
	for !ready {
		height, err := validatorStateClient.GetCurrentHeight(context.Background())
		if err != nil {
			log.Fatal("failed to get current height", zap.Error(err))
			return
		}
		subnetID, err := validatorStateClient.GetSubnetID(context.Background(), chainID)
		if err != nil {
			log.Fatal("failed to get subnet ID", zap.Error(err))
			return
		}
		vdrSet, err := validatorStateClient.GetValidatorSet(context.Background(), height, subnetID)
		if len(vdrSet) == 5 {
			ready = true
		} else {
			time.Sleep(time.Second * 5)
		}
	}

	addressedCall, err := payload.NewAddressedCall(
		utils.RandomBytes(20),
		[]byte(bftrand.Str(24)),
	)
	if err != nil {
		log.Fatal("failed to initialize addressed call", zap.Error(err))
		return
	}

	msg, err := warp.NewUnsignedMessage(networkID, chainID, addressedCall.Bytes())
	if err != nil {
		log.Fatal("failed to create unsigned message", zap.Error(err))
		return
	}
	err = msg.Initialize()
	if err != nil {
		log.Fatal("failed to initialize unsigned message", zap.Error(err))
		return
	}
	for _, warpClient := range warpClients {
		resultAddMsg, err := warpClient.AddMessage(context.Background(), msg.Bytes())
		if err != nil {
			log.Fatal("failed to warp add message", zap.Error(err))
			return
		}
		log.Info("AddMessage result", zap.String("response body", resultAddMsg.MessageID))
	}

	subnetID := "2c1CbR7FGYdeFPB4WaeWphHZrChLH7TQ92FRW6U4mCWTnaxVsB"
	resultMsgSignature, err := warpClients[0].GetMessageAggregateSignature(context.Background(), msg.ID(), WarpQuorumNumerator, subnetID)
	if err != nil {
		log.Fatal("failed to warp get message", zap.Error(err))
		return
	}
	resultMsg, err := warp.ParseMessage(resultMsgSignature.Message)
	if err != nil {
		log.Fatal("failed to parse message", zap.Error(err))
		return
	}
	// get the current status
	s, err := httpClient.Status(context.Background())
	if err != nil {
		log.Fatal("error Status", zap.Error(err))
		return
	}

	log.Info("got status", zap.Any("status", s))

	log.Info("Server Address initial:", zap.String("addr", serverAddr))
	log.Info("Server Address changed:", zap.String("addr", serverAddr[7:]))
	height, err := validatorStateClient.GetCurrentHeight(context.Background())
	if err != nil {
		log.Fatal("failed to get current height", zap.Error(err))
		return
	}
	err = resultMsg.Signature.Verify(context.Background(), &resultMsg.UnsignedMessage, networkID, validatorStateClient, height, WarpQuorumNumerator, WarpQuorumDenominator)
	if err != nil {
		log.Fatal("failed to verify warp aggregated signature", zap.Error(err))
		return
	}
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
