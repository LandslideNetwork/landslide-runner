package internal

import (
	"encoding/json"
	"fmt"
)

const (
	defaultTimeoutBroadcastTxCommit uint16 = 10
	defaultNetworkName                     = "landslide-test"
)

// AppConfig is a Wasm App Config
type AppConfig struct {
	RPCPort  uint16 `json:"rpc_port"`
	GRPCPort uint16 `json:"grpc_port"`
}

type Config struct {
	VMConfig  VMConfig        `json:"vm_config"`
	AppConfig json.RawMessage `json:"app_config"`
}

type VMConfig struct {
	NetworkName              string `json:"network_name"`
	TimeoutBroadcastTxCommit uint16 `json:"timeout_broadcast_tx_commit"`
}

// SetDefaults sets the default values for the config.
func (c *VMConfig) SetDefaults() {
	c.NetworkName = defaultNetworkName
	c.TimeoutBroadcastTxCommit = defaultTimeoutBroadcastTxCommit
}

// Validate returns an error if this is an invalid config.
func (c *VMConfig) Validate() error {
	if len(c.NetworkName) == 0 {
		return fmt.Errorf("network_name can't be empty")
	}

	return nil
}
