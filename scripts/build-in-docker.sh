#!/bin/bash
#
# Copyright 2015 caicloud authors. All rights reserved.

# The script build cyclone master in docker.
#
# Usage:
#   ./osx-build-in-docker.sh [GOOS]

ROOT=$(dirname "${BASH_SOURCE}")/..
GOOS=${1:-"darwin"}

cd $ROOT

docker run --rm \
           -v $(pwd):/go/src/github.com/caicloud/cyclone \
           -w /go/src/github.com/caicloud/cyclone \
           -e GOOS=$1 \
           -e GOARCH=amd64 \
           index.caicloud.io/caicloud/golang:1.6 go build -v .

docker run --rm \
       -v $(pwd):/go/src/github.com/caicloud/cyclone \
       -w /go/src/github.com/caicloud/cyclone/worker \
       -e GOOS=$1 \
       -e GOARCH=amd64 \
       index.caicloud.io/caicloud/golang:1.6 \
       go build -v cyclone-worker.go

cd - > /dev/null
