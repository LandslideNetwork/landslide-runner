package codec_internal

import (
	"github.com/cosmos/cosmos-sdk/client"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/gogo/protobuf/proto"
)

type (
	Codec interface {
		Marshaler
		TxConfig

		GetRegistry() cdctypes.InterfaceRegistry
	}

	Marshaler interface {
		MarshalMessage(dest proto.Message) ([]byte, error)
	}

	TxConfig interface {
		GetTxConfig() client.TxConfig
	}
)
