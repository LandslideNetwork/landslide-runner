package codec_internal

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

var _ Codec = codec{}

type (
	codec struct {
		enc encodingConfig
	}

	encodingConfig struct {
		// InterfaceRegistry cdctypes.InterfaceRegistry
		// Marshaller        sdkcodec.Codec
		TxConfig client.TxConfig
		// Amino             *sdkcodec.LegacyAmino
	}
)

func makeEncodingConfig() encodingConfig {
	// amino := sdkcodec.NewLegacyAmino()
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	marshaller := sdkcodec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaller, tx.DefaultSignModes)

	return encodingConfig{
		// InterfaceRegistry: interfaceRegistry,
		// Marshaller:        marshaller,
		TxConfig: txCfg,
		// Amino:             amino,
	}
}

func NewCodec() Codec {
	enc := makeEncodingConfig()
	// auth.RegisterInterfaces(enc.InterfaceRegistry)
	// std.RegisterLegacyAminoCodec(enc.Amino)
	// std.RegisterInterfaces(enc.InterfaceRegistry)
	// cryptocodec.RegisterInterfaces(enc.InterfaceRegistry)
	// bank.RegisterInterfaces(enc.InterfaceRegistry)
	return codec{enc: enc}
}

func (c codec) GetTxConfig() client.TxConfig {
	return c.enc.TxConfig
}
