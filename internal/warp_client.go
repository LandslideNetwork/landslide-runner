// (c) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package internal

import (
	"context"
	"fmt"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"

	"github.com/ava-labs/avalanchego/ids"
	jsonrpc "github.com/cometbft/cometbft/rpc/jsonrpc/client"
)

var _ Client = (*client)(nil)

type ResultAddMessage struct {
	MessageID string `json:"messageID"`
}

type ResultGetMessage struct {
	Message []byte `json:"message"`
}

type ResultGetMessageSignature struct {
	Signature []byte `json:"signature"`
}

type ResultGetBlockSignature struct {
	Signature []byte `json:"signature"`
}

type Client interface {
	AddMessage(ctx context.Context, message []byte) (*ResultAddMessage, error)
	GetMessage(ctx context.Context, messageID ids.ID) (*ResultGetMessage, error)
	GetMessageSignature(ctx context.Context, messageID ids.ID) (*ResultGetMessageSignature, error)
	GetMessageAggregateSignature(ctx context.Context, messageID ids.ID, quorumNum uint64, subnetIDStr string) ([]byte, error)
	GetBlockSignature(ctx context.Context, blockID ids.ID) (*ResultGetBlockSignature, error)
	GetBlockAggregateSignature(ctx context.Context, blockID ids.ID, quorumNum uint64, subnetIDStr string) ([]byte, error)
}

// client implementation for interacting with EVM [chain]
type client struct {
	*jsonrpc.Client
}

// NewClient returns a Client for interacting with EVM [chain]
func NewClient(rpcAddr string) (Client, error) {
	rpcClient, err := jsonrpc.New(rpcAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial client. err: %w", err)
	}
	return &client{
		Client: rpcClient,
	}, nil
}

func (c *client) AddMessage(ctx context.Context, message []byte) (*ResultAddMessage, error) {
	res := new(ResultAddMessage)
	if _, err := c.Call(ctx, "warp_add_message", map[string]interface{}{"message": message}, res); err != nil {
		return nil, fmt.Errorf("call to warp_add_message failed. err: %w", err)
	}
	return res, nil
}

func (c *client) GetMessage(ctx context.Context, messageID ids.ID) (*ResultGetMessage, error) {
	res := new(ResultGetMessage)
	if _, err := c.Call(ctx, "warp_get_message", map[string]interface{}{"messageID": messageID}, res); err != nil {
		return nil, fmt.Errorf("call to warp_get_message failed. err: %w", err)
	}
	return res, nil
}

func (c *client) GetMessageSignature(ctx context.Context, messageID ids.ID) (*ResultGetMessageSignature, error) {
	res := new(ResultGetMessageSignature)
	if _, err := c.Call(ctx, "warp_get_message_signature", map[string]interface{}{"messageID": messageID}, res); err != nil {
		return nil, fmt.Errorf("call to warp_get_message_signature failed. err: %w", err)
	}
	return res, nil
}

func (c *client) GetMessageAggregateSignature(ctx context.Context, messageID ids.ID, quorumNum uint64, subnetIDStr string) ([]byte, error) {
	var res tmbytes.HexBytes
	if _, err := c.Call(ctx, "warp_get_message_aggregate_signature",
		map[string]interface{}{
			"messageID":   messageID,
			"quorumNum":   quorumNum,
			"subnetIDStr": subnetIDStr,
		}, &res); err != nil {
		return nil, fmt.Errorf("call to warp_get_message_aggregate_signature failed. err: %w", err)
	}
	return res, nil
}

func (c *client) GetBlockSignature(ctx context.Context, blockID ids.ID) (*ResultGetBlockSignature, error) {
	var res ResultGetBlockSignature
	if _, err := c.Call(ctx, "warp_get_block_signature", map[string]interface{}{"blockID": blockID}, &res); err != nil {
		return nil, fmt.Errorf("call to warp_get_block_signature failed. err: %w", err)
	}
	return &res, nil
}

func (c *client) GetBlockAggregateSignature(ctx context.Context, blockID ids.ID, quorumNum uint64, subnetIDStr string) ([]byte, error) {
	var res tmbytes.HexBytes
	if _, err := c.Call(ctx, "warp_get_block_aggregate_signature",
		map[string]interface{}{
			"blockID":     blockID,
			"quorumNum":   quorumNum,
			"subnetIDStr": subnetIDStr,
		}, &res); err != nil {
		return nil, fmt.Errorf("call to warp_get_block_aggregate_signature failed. err: %w", err)
	}
	return res, nil
}
