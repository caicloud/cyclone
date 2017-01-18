#!/bin/bash
#
# Copyright 2016 caicloud authors. All rights reserved.

set -e
set -u
set -o pipefail

CYCLONE_ROOT=$(dirname "${BASH_SOURCE}")/..

IMAGE="cargo.caicloud.io/caicloud/cyclone-worker"
IMAGE_TEG=${1:-"latest"}
BUILD_IN="cargo.caicloud.io/caicloud/golang-docker:1.7-1.11"

# Build and run cyclone.
cd ${CYCLONE_ROOT}
docker run --rm \
       -v $(pwd):/go/src/github.com/caicloud/cyclone \
       -e GOPATH=/go \
       -w "/go/src/github.com/caicloud/cyclone/worker" \
       ${BUILD_IN} go build cyclone-worker.go


docker build -t ${IMAGE}:${IMAGE_TEG} ./worker

cd - > /dev/null
