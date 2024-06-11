# LandslideVM Runner

## Install AvalancheGo

Run following command from [landslidevm](https://github.com/ConsiderItDone/landslidevm) repo to download AvalancheGo

```shell
BASEDIR=/tmp/e2e-test-landslide AVALANCHEGO_BUILD_PATH=/tmp/e2e-test-landslide/avalanchego ./scripts/install_avalanchego_release.sh
```

It downloads compatible version of AvalanceGo.

## Run and test KVStore Application

### Build subnet

Run following command from [landslidevm](https://github.com/ConsiderItDone/landslidevm) repo to download AvalancheGo

```shell
./scripts/build.sh /tmp/e2e-test-landslide/avalanchego/plugins/pjSL9ksard4YE96omaiTkGL5H6XX2W5VEo3ZgWC9S2P6gzs9A
```

### Run subnet

To spin up avalanche node with Landslide Subnet deployed run:

```shell
make run-kvstore
```

You will see something like that:

```
[06-12|01:25:45.448] INFO internal/helpers.go:63 all nodes healthy...
[06-12|01:25:45.451] INFO cmd/main.go:214 subnet rpc url {"node": "node1", "url": "http://127.0.0.1:9750/ext/bc/2od6FnMX7i3jEDspaKbrsSE3VU9qNynhQNgDBNxCvt1zS5ko8x/rpc"}
[06-12|01:25:45.451] INFO cmd/main.go:214 subnet rpc url {"node": "node2", "url": "http://127.0.0.1:9752/ext/bc/2od6FnMX7i3jEDspaKbrsSE3VU9qNynhQNgDBNxCvt1zS5ko8x/rpc"}
[06-12|01:25:45.451] INFO cmd/main.go:214 subnet rpc url {"node": "node3", "url": "http://127.0.0.1:9754/ext/bc/2od6FnMX7i3jEDspaKbrsSE3VU9qNynhQNgDBNxCvt1zS5ko8x/rpc"}
[06-12|01:25:45.451] INFO cmd/main.go:214 subnet rpc url {"node": "node4", "url": "http://127.0.0.1:9756/ext/bc/2od6FnMX7i3jEDspaKbrsSE3VU9qNynhQNgDBNxCvt1zS5ko8x/rpc"}
[06-12|01:25:45.451] INFO cmd/main.go:214 subnet rpc url {"node": "node5", "url": "http://127.0.0.1:9758/ext/bc/2od6FnMX7i3jEDspaKbrsSE3VU9qNynhQNgDBNxCvt1zS5ko8x/rpc"}
[06-12|01:25:45.452] INFO internal/helpers.go:40 Network will run until you CTRL + C to exit...
```

Use one of the RPC endpoint to broadcast transaction:

```shell
curl -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"id\":20,\"method\":\"broadcast_tx_async\",\"params\":{\"tx\":\"WFQ9Yjg=\"}}" http://127.0.0.1:9758/ext/bc/2od6FnMX7i3jEDspaKbrsSE3VU9qNynhQNgDBNxCvt1zS5ko8x/rpc
```

and you get

```shell
{
  "jsonrpc": "2.0",
  "id": 20,
  "result": {
    "code": 0,
    "data": "",
    "log": "",
    "codespace": "",
    "hash": "024C240894026AC6139CFBAB2A3A8D234BAA9F4CA7B05A7241B7F457EAC831E8"
  }
}
```

Query data by key:

```shell
curl -s -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"id\":20,\"method\":\"abci_query\",\"params\":{\"data\":\"5854\"}}" http://127.0.0.1:9758/ext/bc/2od6FnMX7i3jEDspaKbrsSE3VU9qNynhQNgDBNxCvt1zS5ko8x/rpc | jq .
```

Output:

```json
{
  "jsonrpc": "2.0",
  "id": 20,
  "result": {
    "response": {
      "code": 0,
      "log": "exists",
      "info": "",
      "index": "0",
      "key": "WFQ=",
      "value": "Yjg=",
      "proofOps": null,
      "height": "2",
      "codespace": ""
    }
  }
}
```

Get block by number:

```shell
curl -s -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"id\":20,\"method\":\"block\",\"params\":{\"height\":\"2\"}}" http://127.0.0.1:9758/ext/bc/2od6FnMX7i3jEDspaKbrsSE3VU9qNynhQNgDBNxCvt1zS5ko8x/rpc | jq .
```

<details>
  <summary>Output</summary>


```json
{
  "jsonrpc": "2.0",
  "id": 20,
  "result": {
    "block_id": {
      "hash": "72F582C0DF43C48353C4C9E61CE1C8254B7BBB58851F4D9DBD3BB39327332C21",
      "parts": {
        "total": 1,
        "hash": "E5927AA721470C69019AF883DD130BE6E0A917CFCF9BDCA8FBA55B1D05115A6F"
      }
    },
    "block": {
      "header": {
        "version": {
          "block": "11",
          "app": "1"
        },
        "chain_id": "test-chain-U8te75",
        "height": "2",
        "time": "2024-06-11T22:28:59.800566065Z",
        "last_block_id": {
          "hash": "0980836557E1FAAB5A98B823B8C2BF69CFE3EE61874BDDE3E9B4313BA097EDF9",
          "parts": {
            "total": 1,
            "hash": "D3AF12230F844A2D1C0034F487352B149D7E4C27E1868FB22B0CADE51CA19161"
          }
        },
        "last_commit_hash": "25E1513A41B2C67B3E1E8C5D9AB622B2C64509FAA253111C9D00ED031B474224",
        "data_hash": "2C5AC34DDEC2A0A55FDD1D1942FFAA44B90D95605F8AA93DB4F820BB44066D83",
        "validators_hash": "C26E437040C490CF122BA48E10F4B47E1B65E55945764D52CAE438CDE902F666",
        "next_validators_hash": "C26E437040C490CF122BA48E10F4B47E1B65E55945764D52CAE438CDE902F666",
        "consensus_hash": "048091BC7DDC283F77BFBF91D73C44DA58C3DF8A9CBC867405D8B7F3DAADA22F",
        "app_hash": "0000000000000000",
        "last_results_hash": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
        "evidence_hash": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
        "proposer_address": "0000000000000000000000000000000000000000"
      },
      "data": {
        "txs": [
          "WFQ9Yjg="
        ]
      },
      "evidence": {
        "evidence": []
      },
      "last_commit": {
        "height": "1",
        "round": 0,
        "block_id": {
          "hash": "0980836557E1FAAB5A98B823B8C2BF69CFE3EE61874BDDE3E9B4313BA097EDF9",
          "parts": {
            "total": 1,
            "hash": "D3AF12230F844A2D1C0034F487352B149D7E4C27E1868FB22B0CADE51CA19161"
          }
        },
        "signatures": [
          {
            "block_id_flag": 3,
            "validator_address": "D4CAE735FFC8559F79A26DB8B75E39395F97C2AE",
            "timestamp": "2024-06-11T22:28:59.800566065Z",
            "signature": "AA=="
          }
        ]
      }
    }
  }
}
```

</details>


### End-to-end tests

To run e2e tests:

```shell
make e2e-kvstore 
```

## Run and CosmWasm Application

Run following command from [landslidevm](https://github.com/ConsiderItDone/landslidevm) repo to download AvalancheGo

```shell
./scripts/build_wasm.sh /tmp/e2e-test-landslide/avalanchego/plugins/pjSL9ksard4YE96omaiTkGL5H6XX2W5VEo3ZgWC9S2P6gzs9A
```

### Run subnet

To spin up avalanche node with Landslide Subnet deployed run:

```shell
make run-wasm
```