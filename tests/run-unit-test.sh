#!/bin/sh
#
# Copyright 2016 caicloud authors. All rights reserved.

set -e
set -u
set -o pipefail

# get this file path
echo "$0" | grep -q "$0"
if [ $? -eq 0 ];
then
    cd "$(dirname ${BASH_SOURCE})"
    CUR_FILE=$(pwd)/$(basename ${BASH_SOURCE})
    CUR_DIR=$(dirname ${CUR_FILE})
    cd - > /dev/null
else
    if [ ${0:0:1} = "/" ]; then
        CUR_FILE=$0
    else
        CUR_FILE=$(pwd)/$0
    fi
    CUR_DIR=$(dirname ${CUR_FILE})
fi

# eliminate relative path ，like a/..b/c
CYCLONE_ROOT=$(dirname ${CUR_DIR})

cyclone_src="/go/src/github.com/caicloud/cyclone"

BUILD_IN="cargo.caicloud.io/caicloud/golang-docker:1.8-17.03"
echo "build in ${BUILD_IN}"
docker run --rm \
  -v ${CYCLONE_ROOT}:${cyclone_src} \
  -e GOPATH=/go \
  -w ${cyclone_src} \
  ${BUILD_IN} go test -cover $(go list ./... | grep -v '/vendor/' | grep -v '/tests/')
