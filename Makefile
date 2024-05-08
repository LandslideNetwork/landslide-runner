.SILENT:
.EXPORT_ALL_VARIABLES:

.PHONY: run
run:
	cd cmd; go run main.go run

e2e-kv:
	cd cmd; go run main.go e2e kvstore http://127.0.0.1:9750/ext/bc/dVhHMjftS7WZe6n23DwssWoazoQ9nkdX1NC71xfBThbnkkbSX/rpc