#!/bin/bash
#
# Copyright 2015 caicloud authors. All rights reserved.

set -o errexit
set -o nounset
set -o pipefail

CYCLONE_ROOT=$(dirname "${BASH_SOURCE}")/..

# Build and run cyclone.
cd ${CYCLONE_ROOT}
#docker login -u caicloud -p caicloud2015ABC -e caicloud index.caicloud.io
docker run --rm \
       -v `pwd`:/go/src/github.com/caicloud/cyclone \
       -e GOPATH=/go:/go/src/github.com/caicloud/cyclone/vendor gaojq007/golang-gcc:1.6-alpine sh \
       -c "cd /go/src/github.com/caicloud/cyclone/worker && go build cyclone-worker.go"

docker build -t index.caicloud.io/caicloud/cyclone-worker ./worker

cd - > /dev/null
