#!/bin/bash
#
# Copyright 2016 caicloud authors. All rights reserved.

set -e
set -u
set -o pipefail

source "$(dirname "${BASH_SOURCE}")/lib/common.sh"
cd ${CYCLONE_ROOT}


IMAGE="cargo.caicloud.io/caicloud/cyclone-worker"
IMAGE_TEG=${1:-"latest"}
BUILD_IN="cargo.caicloud.io/caicloud/golang-docker:1.8-17.03"
cyclone_src="/go/src/github.com/caicloud/cyclone"

# Build and run cyclone.
docker run --rm \
       -v ${CYCLONE_ROOT}:${cyclone_src} \
       -e GOPATH=/go \
       -w ${cyclone_src} \
       ${BUILD_IN} bash -c "go build -o cyclone-worker github.com/caicloud/cyclone/cmd/worker"

docker build -t ${IMAGE}:${IMAGE_TEG} -f build/worker/Dockerfile .

cd - > /dev/null
