package internal

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// getProtoCodec - create new proto Codec for keyring.
func getProtoCodec() codec.Codec {
	registry := cdctypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

// NewKeyring - create new keyring.
func NewKeyring() (keyring.Keyring, error) {
	kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, "", nil, getProtoCodec())
	if err != nil {
		return nil, err
	}

	return kr, nil
}
