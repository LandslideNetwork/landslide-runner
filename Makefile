.SILENT:
.EXPORT_ALL_VARIABLES:

.PHONY: run
run:
	cd cmd; go run main.go run

.PHONY: e2e-kvstore
e2e-kvstore:
	cd cmd; go run main.go e2e kvstore