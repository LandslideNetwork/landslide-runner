package internal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ava-labs/avalanche-network-runner/network"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/cometbft/cometbft/libs/rand"
	"go.uber.org/zap"
)

// When we get a SIGINT or SIGTERM, stop the network and close [closedOnShutdownCh]
// Blocks until a signal is received on [signalChan], upon which
// [n.Stop()] is called. If [signalChan] is closed, does nothing.
// Closes [closedOnShutdownChan] amd [signalChan] when done shutting down network.
// This function should only be called once.
func GracefulShutdown(nw network.Network, log logging.Logger) {
	signalsChan := make(chan os.Signal, 1)
	closedOnShutdownCh := make(chan struct{})
	defer func() {
		close(closedOnShutdownCh)
		close(signalsChan)
		signal.Reset()
	}()

	signal.Notify(signalsChan, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-signalsChan
		log.Info("got OS signal", zap.Stringer("signal", sig))

		closedOnShutdownCh <- struct{}{}
	}()

	log.Info("Network will run until you CTRL + C to exit...")
	<-closedOnShutdownCh

	log.Info("Shutting down network...")
	if err := nw.Stop(context.Background()); err != nil {
		log.Error("error while shutting down network", zap.Error(err))
	}
}

// MakeTxKV returns a text transaction, allong with expected key, value pair
func MakeTxKV() ([]byte, []byte, []byte) {
	k := []byte(rand.Str(2))
	v := []byte(rand.Str(2))
	return k, v, append(k, append([]byte("="), v...)...)
}

// Wait until the nodes in the network are ready
func Await(nw network.Network, log logging.Logger, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.Info("waiting for all nodes to report healthy...")
	err := nw.Healthy(ctx)
	if err == nil {
		log.Info("all nodes healthy...")
	}
	return err
}

// Copy a file from src to dst
func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, source)
	if err := os.Chmod(dst, 0777); err != nil {
		return 0, err
	}
	return nBytes, err
}
