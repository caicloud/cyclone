#!/bin/bash

# Notice: change registry to which you can access to
# assume you running the script on mac

set -e
set -u
set -o pipefail


source "$(dirname "${BASH_SOURCE}")/lib/common.sh"
cd ${CYCLONE_ROOT}


# Clean up local run of cyclone.
function cleanup {

    unset DEBUG
    unset CYCLONE_SERVER
    unset MONGODB_HOST
    unset REGISTRY_LOCATION
    unset REGISTRY_USERNAME
    unset REGISTRY_PASSWORD
    unset WORKER_IMAGE
    unset DOCKER_HOST
    unset HOST_IP

    [ -n "${CYCLONE_PID-}" ] && ps -p ${CYCLONE_PID} > /dev/null && kill ${CYCLONE_PID}

    if [[ -n $(docker ps -a | grep mongo) ]];then
        docker rm -f mongo
    fi
    if [[ -n $(docker ps -a | grep cyclone_server) ]];then
        docker rm -f cyclone_server
    fi

    echo "=> cleanup now."

}

function run_e2e {
    cleanup

    # for worker to connect server
    export HOST_IP="$(ifconfig | grep "inet " | grep -v 127.0.0.1 | tail -1 | cut -d " " -f 2 )"

    export DEBUG=true
    export MONGODB_HOST=127.0.0.1:27017
    export CYCLONE_SERVER=http://${HOST_IP}:7099
    export REGISTRY_LOCATION=cargo.caicloud.io
    export REGISTRY_USERNAME=caicloudadmin
    export REGISTRY_PASSWORD=caicloudadmin
    export WORKER_IMAGE=cargo.caicloudprivatetest.com/caicloud/cyclone-worker:latest
    export CYCLONE_CLOUD_KIND=kubernetes


    echo "setup mongo"
    docker run -d --name mongo -p 27017:27017 mongo:3.0.5 mongod --smallfiles

    echo "buiding server"
    go build -v -o bin/server github.com/caicloud/cyclone/cmd/server

    echo "buiding worker"
    # worker run in linux, so need cross compiling
    GOOS=linux GOARCH=amd64 go build -v -o bin/worker github.com/caicloud/cyclone/cmd/worker
    docker build -t ${WORKER_IMAGE} -f build/worker/Dockerfile .

    docker push ${WORKER_IMAGE}

    echo "start server"
    ./bin/server &
    CYCLONE_PID=$!

    echo "start testing ..."
    # go test compile
    go test -i ./tests/...

    go test -v ./tests/project

}

trap cleanup SIGINT EXIT SIGQUIT

run_e2e

cleanup
