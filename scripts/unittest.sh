#!/bin/sh
#
# Copyright 2016 caicloud authors. All rights reserved.

set -e
set -u
set -o pipefail

source "$(dirname "${BASH_SOURCE}")/lib/common.sh"
cd ${CYCLONE_ROOT}

cyclone_src="/go/src/github.com/caicloud/cyclone"

BUILD_IN="cargo.caicloud.io/caicloud/golang-docker:1.7-17.03"
echo "build in ${BUILD_IN}"
docker run --rm \
  -v ${CYCLONE_ROOT}:${cyclone_src} \
  -e GOPATH=/go \
  -w ${cyclone_src} \
  ${BUILD_IN} go test -cover $(go list ./... | grep -v '/vendor/' | grep -v '/tests/')

BUILD_IN="cargo.caicloud.io/caicloud/golang-docker:1.8-17.03"
echo "build in ${BUILD_IN}"
docker run --rm \
  -v ${CYCLONE_ROOT}:${cyclone_src} \
  -e GOPATH=/go \
  -w ${cyclone_src} \
  ${BUILD_IN} go test -cover $(go list ./... | grep -v '/vendor/' | grep -v '/tests/')
