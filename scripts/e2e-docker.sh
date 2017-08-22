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
    echo "=> cleanup start."

    unset DEBUG
    unset CYCLONE_SERVER
    unset MONGODB_HOST
    unset KAFKA_HOST
    unset ETCD_HOST
    unset LOG_SERVER
    unset REGISTRY_LOCATION
    unset REGISTRY_USERNAME
    unset REGISTRY_PASSWORD
    unset WORKER_IMAGE
    unset CLAIR_DISABLE
    unset DOCKER_HOST
    unset HOST_IP

    [ -n "${CYCLONE_PID-}" ] && ps -p ${CYCLONE_PID} > /dev/null && kill ${CYCLONE_PID}

    if [[ -n $(docker ps -a | grep kafka) ]];then
        docker rm -f kafka
    fi
    if [[ -n $(docker ps -a | grep zookeeper) ]];then
        docker rm -f zookeeper
    fi
    if [[ -n $(docker ps -a | grep etcd) ]];then
        docker rm -f etcd
    fi
    if [[ -n $(docker ps -a | grep mongo) ]];then
        docker rm -f mongo
    fi
    if [[ -n $(docker ps -a | grep cyclone_server) ]];then
        docker rm -f cyclone_server
    fi

    echo "=> cleanup finished."
}

function run_e2e {
    cleanup

    # for worker to connect server
    export HOST_IP="$(ifconfig | grep "inet " | grep -v 127.0.0.1 | tail -1 | cut -d " " -f 2 )"

    export DEBUG=true
    export CLAIR_DISABLE=true
    export MONGODB_HOST=127.0.0.1:27017
    export KAFKA_HOST=127.0.0.1:9092
    export ETCD_HOST=http://127.0.0.1:2379
    export CYCLONE_SERVER=http://${HOST_IP}:7099
    export LOG_SERVER=ws://${HOST_IP}:8000/ws
    export REGISTRY_LOCATION=cargo.caicloud.io
    export REGISTRY_USERNAME=caicloudadmin
    export REGISTRY_PASSWORD=caicloudadmin
    export WORKER_IMAGE=cargo.caicloud.io/caicloud/cyclone-worker:latest

    if [[ $OS == "darwin" ]] 
    then
        # for mac docker
        export DOCKER_HOST=unix:///${HOME}/Library/Containers/com.docker.docker/Data/s60
    else
        # for linux
        export DOCKER_HOST=unix:///var/run/docker.sock
    fi

    echo "setup zookeeper"
    docker run -d --name zookeeper wurstmeister/zookeeper:3.4.6
    sleep 2
    echo "setup kafka"
    docker run -d --name kafka --hostname kafka \
                -p 9092:9092 \
                -e KAFKA_ADVERTISED_HOST_NAME=0.0.0.0 \
                -e KAFKA_ADVERTISED_PORT=9092 \
                -e KAFKA_LOG_DIRS=/data/kafka_log \
                --link zookeeper:zk \
                wurstmeister/kafka:0.10.1.0

    echo "setup etcd"
    docker run -d --name etcd \
               -p 2379:2379 -p 2380:2380 -p 4001:4001 \
                quay.io/coreos/etcd:v3.1.3 \
                etcd -name=etcd0 \
                -advertise-client-urls http://0.0.0.0:2379 \
                -listen-client-urls http://0.0.0.0:2379 \
                -initial-advertise-peer-urls http://${HOST_IP}:2380 \
                -listen-peer-urls http://0.0.0.0:2380 \
                -initial-cluster-token etcd-cluster-1 \
                -initial-cluster etcd0=http://${HOST_IP}:2380 \
                -initial-cluster-state new

    echo "setup mongo"
    docker run -d --name mongo -p 27017:27017 mongo:3.0.5 mongod --smallfiles

    echo "buiding server"
    go build -i -v -o cyclone-server github.com/caicloud/cyclone/cmd/server

    echo "buiding worker"
    # worker run in linux, so need cross compiling
    GOOS=linux GOARCH=amd64 go build -i -v -o cyclone-worker github.com/caicloud/cyclone/cmd/worker 
    docker -H ${DOCKER_HOST} build -t ${WORKER_IMAGE} -f Dockerfile.worker .

    echo "start server"
    ./cyclone-server &
    CYCLONE_PID=$!

    echo "start testing ..."
    # go test compile
    go test -i ./tests/...

    go test -v ./tests/service 
    go test -v ./tests/version 
    go test -v ./tests/yaml

}

trap cleanup SIGINT EXIT SIGQUIT

run_e2e

cleanup
