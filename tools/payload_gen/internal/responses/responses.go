package responses

import (
	"encoding/hex"
	"fmt"

	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"go.uber.org/zap"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
)

// QueryAllBalancesResponses
// [06-19|19:51:05.798] INFO internal/wasm.go:96 Balance query success {"address": "wasm1kng6sqkm0mjuh09cwz6u86f75lmeflj9h0fqhr", "balance": "0a130a057374616b65120a3130303030303030303012021001"}
// [06-19|19:51:05.799] INFO internal/wasm.go:96 Balance query success {"address": "wasm1c4w4jxdkvj3ygdycdkjy98jve6w0d7257eqfx9", "balance": "1200"}
// [06-19|19:51:05.799] INFO internal/wasm.go:36 Sending 5000000 tokens from user1 to user2
// [06-19|19:51:20.803] INFO internal/wasm.go:96 Balance query success {"address": "wasm1kng6sqkm0mjuh09cwz6u86f75lmeflj9h0fqhr", "balance": "0a120a057374616b65120939393439303030303012021001"}
// [06-19|19:51:20.804] INFO internal/wasm.go:96 Balance query success {"address": "wasm1c4w4jxdkvj3ygdycdkjy98jve6w0d7257eqfx9", "balance": "0a100a057374616b6512073530303030303012021001"}
//
// Decoded balances:
// balances:<denom:"stake" amount:"1000000000" > pagination:<total:1 >
// pagination:<>
// balances:<denom:"stake" amount:"994900000" > pagination:<total:1 >
// balances:<denom:"stake" amount:"5000000" > pagination:<total:1 >

var (
	user1Balance1 = "0a130a057374616b65120a3130303030303030303012021001"
	user2Balance1 = "1200"
	user1Balance2 = "0a120a057374616b65120939393439303030303012021001"
	user2Balance2 = "0a100a057374616b6512073530303030303012021001"
)

func PrintDecodedBalances(log *zap.SugaredLogger) {
	printDecoded := func(hexData string) {
		// decode hexdata into bytes
		data, err := hex.DecodeString(hexData)
		if err != nil {
			log.Fatalf("Failed to decode hexdata: %v", err)
		}

		res := &bank.QueryAllBalancesResponse{}
		err = encoding.GetCodec(proto.Name).Unmarshal(data, res)
		if err != nil {
			log.Fatalf("Failed to unmarshal response: %v", err)
		}

		fmt.Printf("%+v", res)
		fmt.Println()
	}

	printDecoded(user1Balance1)
	printDecoded(user2Balance1)
	printDecoded(user1Balance2)
	printDecoded(user2Balance2)

}
