package internal

import (
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	denom   = "stake"
	chainID = "landslide-test"
)

var defaultGasPrice = sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(100000)))

const (
	User1Mnemonic = "tip yard art tape orchard universe angle flame wave gadget raven coyote crater ethics able evoke luxury predict leopard delay peanut embody blast soap"
	User2Mnemonic = "thumb scorpion sting term fiscal dream nephew fitness session nation happy attitude canyon raise omit town garment enroll term trial math neglect truck feature"
)

// SetPrefixes - set prefix for accounts and validators.
func SetPrefixes(accountAddressPrefix string) {
	// Set prefixes
	accountPubKeyPrefix := accountAddressPrefix + "pub"
	validatorAddressPrefix := accountAddressPrefix + "valoper"
	validatorPubKeyPrefix := accountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := accountAddressPrefix + "valcons"
	consNodePubKeyPrefix := accountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(accountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.Seal()
}

// getProtoCodec - create new proto codec for keyring.
func getProtoCodec() codec.Codec {
	registry := cdctypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

func NewKeyring() (keyring.Keyring, error) {
	kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, "", nil, getProtoCodec())
	if err != nil {
		return nil, err
	}

	return kr, nil
}
