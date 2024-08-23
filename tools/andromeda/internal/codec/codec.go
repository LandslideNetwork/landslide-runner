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
		TxConfig client.TxConfig
	}
)

func makeEncodingConfig() encodingConfig {
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	marshaller := sdkcodec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaller, tx.DefaultSignModes)

	return encodingConfig{
		TxConfig: txCfg,
	}
}

func NewCodec() Codec {
	enc := makeEncodingConfig()
	return codec{enc: enc}
}

func (c codec) GetTxConfig() client.TxConfig {
	return c.enc.TxConfig
}
