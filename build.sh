#!/usr/bin/env bash
set -eo pipefail

docker build -t ipprovider:`git rev-parse HEAD` .
