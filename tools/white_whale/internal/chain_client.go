package internal

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	cd "whitewhale/internal/codec"
)

var _ ChainClientInterface = &ChainClient{}

// ChainClientInterface - chain client interface.
// It is used to interact with the blockchain,
// perform transactions, and manage accounts.
type ChainClientInterface interface {
	AddAccount(name, mnemonic string, sequence, number uint64)
	GetAccount(name string) (AccountInfo, bool)
	IncreaseSequence(name string) error
	GetSignedTxBytes(signerAccountName string, msg types.Msg, gasPriceOverride uint64) ([]byte, error)
}

type (
	// AccountInfo - account information.
	AccountInfo struct {
		Name     string
		Mnemonic string
		Sequence uint64
		Number   uint64
		Address  string
	}

	// ChainClient - chain client.
	ChainClient struct {
		chainID        string
		gasLimit       uint64
		signerAccounts map[string]AccountInfo
		Codec          cd.Codec
		keyring        keyring.Keyring
		log            *zap.Logger
		denom          string
	}
)

// NewChainClient - create a new chain client.
func NewChainClient(gasLimit uint64, prefix string, chainID string, denom string, log *zap.Logger) *ChainClient {
	SetPrefixes(prefix)
	kr, err := NewKeyring()
	if err != nil {
		log.Fatal("error creating keyring: %v", zap.Error(err))
		return nil
	}

	return &ChainClient{
		chainID:        chainID,
		denom:          denom,
		gasLimit:       gasLimit,
		signerAccounts: make(map[string]AccountInfo),
		Codec:          cd.NewCodec(),
		keyring:        kr,
		log:            log,
	}
}

// AddAccount - add account to the chain client.
func (c *ChainClient) AddAccount(name, mnemonic string, sequence, number uint64) {
	newAcc, err := c.keyring.NewAccount(name, mnemonic, "", sdk.FullFundraiserPath, hd.Secp256k1)
	if err != nil {
		c.log.Fatal("error creating account: %v", zap.Error(err))
		return
	}

	addr, err := newAcc.GetAddress()
	if err != nil {
		c.log.Fatal("error creating account: %v", zap.Error(err))
		return
	}

	c.signerAccounts[name] = AccountInfo{
		Name:     name,
		Mnemonic: mnemonic,
		Sequence: sequence,
		Number:   number,
		Address:  addr.String(),
	}
}

// GetDenom - get the denomination used by the chain client.
func (c *ChainClient) GetDenom() string {
	return c.denom
}

// GetAccount - get account by name.
func (c *ChainClient) GetAccount(name string) (AccountInfo, bool) {
	acc, ok := c.signerAccounts[name]
	return acc, ok
}

// IncreaseSequence - increase sequence number for the account.
func (c *ChainClient) IncreaseSequence(name string) error {
	acc, ok := c.signerAccounts[name]
	if !ok {
		return fmt.Errorf("account not found")
	}

	acc.Sequence++
	c.signerAccounts[name] = acc
	return nil
}

// SetSequence - sets a sequence number for the account.
func (c *ChainClient) SetSequence(name string, seq uint64) error {
	acc, ok := c.GetAccount(name)
	if !ok {
		return fmt.Errorf("account not found")
	}

	acc.Sequence = seq
	c.signerAccounts[name] = acc
	return nil
}

// GetSignedTxBytes - get signed tx bytes.
func (c *ChainClient) GetSignedTxBytes(
	signerAccountName string,
	msg types.Msg,
	gasPriceOverride uint64,
) ([]byte, error) {
	signer, err := c.keyring.Key(signerAccountName)
	if err != nil {
		return nil, fmt.Errorf("not found signer: %s", err)
	}

	// Get the account from the keyring
	acc, exists := c.GetAccount(signerAccountName)
	if !exists {
		return nil, fmt.Errorf("account not found")
	}

	// Set the gas price
	var gasPrice sdk.Coins
	if gasPriceOverride == 0 {
		gasPrice = sdk.NewCoins(sdk.NewInt64Coin(c.denom, 100000))
	} else {
		gasPrice = sdk.NewCoins(sdk.NewInt64Coin(c.denom, int64(gasPriceOverride)))
	}

	// Create a new tx builder and set the message
	txBuilder := c.Codec.GetTxConfig().NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("set msg error: %s", err)
	}
	// Set the fee amount and gas limit
	txBuilder.SetFeeAmount(gasPrice)
	txBuilder.SetGasLimit(gasPrice.AmountOf(c.denom).Mul(math.NewInt(2)).Uint64())

	// Sign the transaction
	factory := tx.Factory{}.
		WithKeybase(c.keyring).
		WithChainID(c.chainID).
		WithAccountNumber(acc.Number).
		WithSequence(acc.Sequence).
		WithTxConfig(c.Codec.GetTxConfig())

	if err := tx.Sign(context.Background(), factory, signer.Name, txBuilder, true); err != nil {
		return nil, fmt.Errorf("sign tx error: %s", err)
	}

	txBytes, err := c.Codec.GetTxConfig().TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("tx encode error: %s", err)
	}

	return txBytes, nil
}
