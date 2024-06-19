package codec_internal

import (
	"github.com/cosmos/cosmos-sdk/client"
)

type (
	Codec interface {
		GetTxConfig() client.TxConfig
	}
)
