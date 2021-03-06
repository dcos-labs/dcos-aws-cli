#!/bin/bash

# exit immediately on failure
set -e

BASEDIR=$(pwd)/$(dirname "$0")
cd "$BASEDIR"

CLI_EXE_NAME=dcos-aws-cli

# ---

# go
cd ../src

go get

CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags="-s -w" -o $CLI_EXE_NAME".exe"
echo  $(shasum -a 256 $CLI_EXE_NAME.exe)
CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -ldflags="-s -w" -o $CLI_EXE_NAME"-darwin"
echo $(shasum -a 256 $CLI_EXE_NAME-darwin)
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="-s -w" -o $CLI_EXE_NAME"-linux"
echo $(shasum -a 256 $CLI_EXE_NAME-linux)
