.SILENT:
.EXPORT_ALL_VARIABLES:

.PHONY: run-kvstore
run-kvstore:
	cd cmd; go run main.go run kvstore

.PHONY: run-wasm
run-wasm:
	cd cmd; go run main.go run wasm

.PHONY: e2e-kvstore
e2e-kvstore:
	cd cmd; go run main.go e2e kvstore