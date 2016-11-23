#!/bin/bash
#
# Copyright 2016 caicloud authors. All rights reserved.

set -e
set -u
set -o pipefail

CYCLONE_ROOT=$(dirname "${BASH_SOURCE}")/..

# Build and run cyclone.
cd ${CYCLONE_ROOT}
docker run --rm \
       -v `pwd`:/go/src/github.com/caicloud/cyclone \
       -e GOPATH=/go:/go/src/github.com/caicloud/cyclone/vendor cargo.caicloud.io/caicloud/golang-gcc:1.6-alpine sh \
       -c "cd /go/src/github.com/caicloud/cyclone/worker && go build cyclone-worker.go"

docker build -t cargo.caicloud.io/caicloud/cyclone-worker ./worker

cd - > /dev/null
