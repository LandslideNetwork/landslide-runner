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

	cd "payload_gen/internal/codec"
)

// SetPrefixes - set prefixes for the chain.
const prefix = "wasm"

type (
	AccountInfo struct {
		Name     string
		Mnemonic string
		Sequence uint64
		Number   uint64
		Address  string
	}

	ChainClient struct {
		chainID        string
		gasLimit       uint64
		signerAccounts map[string]AccountInfo
		Codec          cd.Codec
		keyring        keyring.Keyring
		log            *zap.SugaredLogger
	}
)

// NewChainClient - create a new chain client.
func NewChainClient(
	gasLimit uint64,
	log *zap.SugaredLogger,
) *ChainClient {
	SetPrefixes(prefix)
	kr, err := NewKeyring()
	if err != nil {
		log.Fatalf("error creating keyring: %v", err)
		return nil
	}

	return &ChainClient{
		chainID:        chainID,
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
		c.log.Fatalf("error creating account: %v", err)
		return
	}

	addr, err := newAcc.GetAddress()
	if err != nil {
		c.log.Fatalf("error creating account: %v", err)
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

	acc, exists := c.GetAccount(signerAccountName)
	if !exists {
		return nil, fmt.Errorf("account not found")
	}

	var gasPrice sdk.Coins
	if gasPriceOverride == 0 {
		gasPrice = defaultGasPrice
	} else {
		gasPrice = sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(int64(gasPriceOverride))))
	}

	txBuilder := c.Codec.GetTxConfig().NewTxBuilder()
	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, fmt.Errorf("set msg error: %s", err)
	}

	factory := tx.Factory{}.
		WithKeybase(c.keyring).
		WithChainID(c.chainID).
		WithAccountNumber(acc.Number).
		WithSequence(acc.Sequence).
		WithTxConfig(c.Codec.GetTxConfig())

	txBuilder.SetFeeAmount(gasPrice)
	txBuilder.SetGasLimit(gasPrice.AmountOf(denom).Mul(math.NewInt(2)).Uint64())

	if err := tx.Sign(context.Background(), factory, signer.Name, txBuilder, true); err != nil {
		return nil, fmt.Errorf("sign tx error: %s", err)
	}

	txBytes, err := c.Codec.GetTxConfig().TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, fmt.Errorf("tx encode error: %s", err)
	}

	return txBytes, nil
}
