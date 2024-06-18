package codec_internal

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
)

// encodingConfig specifies the concrete encoding types to use for a given app.
// This is provided for compatibility between protobuf and amino implementations.
type (
	encodingConfig struct {
		InterfaceRegistry cdctypes.InterfaceRegistry
		Marshaller        sdkcodec.Codec
		TxConfig          client.TxConfig
		Amino             *sdkcodec.LegacyAmino
	}

	codec struct {
		enc encodingConfig
	}
)

func makeEncodingConfig() encodingConfig {
	amino := sdkcodec.NewLegacyAmino()
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	marshaller := sdkcodec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaller, tx.DefaultSignModes)

	return encodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaller:        marshaller,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

func NewCodec() Codec {
	enc := makeEncodingConfig()
	auth.RegisterInterfaces(enc.InterfaceRegistry)
	std.RegisterLegacyAminoCodec(enc.Amino)
	std.RegisterInterfaces(enc.InterfaceRegistry)
	cryptocodec.RegisterInterfaces(enc.InterfaceRegistry)
	bank.RegisterInterfaces(enc.InterfaceRegistry)
	return codec{enc: enc}
}

func (c codec) GetTxConfig() client.TxConfig {
	return c.enc.TxConfig
}

func (c codec) MarshalMessage(dest proto.Message) ([]byte, error) {
	return c.enc.Marshaller.MarshalJSON(dest)
}

func (c codec) GetRegistry() cdctypes.InterfaceRegistry {
	return c.enc.InterfaceRegistry
}
