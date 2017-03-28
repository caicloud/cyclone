#!/bin/bash
#
# Copyright 2016 caicloud authors. All rights reserved.

set -e
set -u
set -o pipefail

#获得该文件的位置
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

#去掉路径中的相对路径，如a/..b/c
CYCLONE_ROOT=$(dirname ${CUR_DIR})

IMAGE="cargo.caicloud.io/caicloud/cyclone-worker"
IMAGE_TEG=${1:-"latest"}
BUILD_IN="cargo.caicloud.io/caicloud/golang-docker:1.8-17.03"
cyclone_src="/go/src/github.com/caicloud/cyclone"

# Build and run cyclone.
cd ${CYCLONE_ROOT}
docker run --rm \
       -v ${CYCLONE_ROOT}:${cyclone_src} \
       -e GOPATH=/go \
       -w ${cyclone_src} \
       ${BUILD_IN} bash -c "go build -o cyclone-worker github.com/caicloud/cyclone/cmd/worker"

docker build -t ${IMAGE}:${IMAGE_TEG} -f Dockerfile.worker .

cd - > /dev/null
