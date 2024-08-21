#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 blockchainID"
    exit 1
fi

cd ./tools/andromeda; go run main.go $1

