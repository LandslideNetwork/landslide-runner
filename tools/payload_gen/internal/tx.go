package internal

import (
	"encoding/hex"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
)

// BankSendTxHex creates a MsgSend protobuf message and prints the hex encoded tx
func BankSendTxHex(
	client *ChainClient,
	log *zap.SugaredLogger,

	from, to string,
	amount uint64,

) {
	fromAcc, ok := client.signerAccounts[from]
	if !ok {
		log.Fatalf("account %s not found", from)
		return
	}
	toAcc, ok := client.signerAccounts[to]
	if !ok {
		log.Fatalf("account %s not found", to)
		return

	}
	msg := &bank.MsgSend{
		FromAddress: fromAcc.Address,
		ToAddress:   toAcc.Address,
		Amount:      types.NewCoins(types.NewCoin(denom, math.NewInt(int64(amount)))),
	}

	txBytes, err := client.GetSignedTxBytes(from, msg, 0)
	if err != nil {
		log.Fatalf("error getting signed tx bytes: %v", err)
		return
	}

	// Print the hex encoded txBytes
	log.Infof(
		"bank.MsgSend hex encoded for from %s to %s, amount %d: %s",
		from,
		to,
		amount,
		hex.EncodeToString(txBytes),
	)
}
