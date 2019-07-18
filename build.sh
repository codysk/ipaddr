#!/usr/bin/env bash
set -eo pipefail

CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -mod=vendor ipprovider.go

docker build -t ipprovider:`git rev-parse HEAD` .
