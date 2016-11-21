#!/bin/bash
#
# Copyright 2016 caicloud authors. All rights reserved.

# The script starts a local cyclone instance and run e2e tests. It
# assumes that dependencies are met following guidance from README.

CYCLONE_ROOT=$(dirname "${BASH_SOURCE}")/..
TMPDIR=`mktemp -d /tmp/cyclone.XXXXXXXXXX`
REGISTRY_DATA=${TMPDIR}/registry
REGISTRY_AUTH_LOG=${TMPDIR}/auth-log

# Timestamped log, e.g. log "cluster created".
#
# Input:
#   $1 Log string.
function log {
  echo -e "[`TZ=Asia/Shanghai date`] ${1}"
}

# Note we must pass the same environments to `go test`.
# Dynamic configs.
export DOCKER_HOST=${DOCKER_HOST:-"unix:///var/run/docker.sock"}
export WORK_DOCKER_HOST=${WORK_DOCKER_HOST:-"unix:///var/run/docker.sock"}
export CYCLONE_SERVER_HOST=${CYCLONE_SERVER_HOST:-"http://172.17.0.1:7099"}
export WORK_REGISTRY_LOCATION=${WORK_REGISTRY_LOCATION:-"localhost:5000"}
# Static configs.
export ENABLE_CAICLOUD_AUTH="false"
export REGISTRY_LOCATION="localhost:5000"
export REGISTRY_USERNAME="admin"
export REGISTRY_PASSWORD="admin_password"
export JENKINS="yes"
export SUCCESSTEMPLATE="./notify/provider/success.html"
export ERRORTEMPLATE="./notify/provider/error.html"
export REGISTRY_AUTH_LOG=${REGISTRY_AUTH_LOG}
export REGISTRY_DATA=${REGISTRY_DATA}
export MONGO_DB_IP="127.0.0.1:28017"
export KAFKA_SERVER_IP="127.0.0.1:9092"
export CLAIR_SERVER_IP="127.0.0.1:6060"
export ETCD_SERVER_IP="http://127.0.0.1:2379"
export WORKER_IMAGE="cargo.caicloud.io/caicloud/cyclone-worker"
export CLIENTID=""
export CLIENTIDSECRET=""
export CLIENTID_GITLAB=0
export CLIENTIDSECRET_GITLAB=0
export SERVER_GITLAB=https://gitlab.com
export UI_SERVER_PATH="http://127.0.0.1:7099"

function run-local-up {
  if [[ "$(which docker)" == "" ]]; then
    echo "Unable to find docker"
    exit
  fi
  docker pull cargo.caicloud.io/circle/e2e-test-long-running-task
  docker tag cargo.caicloud.io/circle/e2e-test-long-running-task localhost:5000/minimal-long-running-task
  mkdir ${REGISTRY_DATA} ${REGISTRY_AUTH_LOG}

  log "Mongo, Kafka, and the registry are all running in a docker container, cyclone running in local."
  cd ${CYCLONE_ROOT}
  docker-compose -f docker-compose-dev.yml up -d --force-recreate
  cd - > /dev/null

  # Run local registry.
  ${CYCLONE_ROOT}/scripts/registry/start.sh

  # Prepare for running yaml test
  # docker pull mongo:3.0.5
  # docker tag mongo:3.0.5 localhost:5000/mongo:3.0.5 > /dev/null

  echo "-> registry location: ${REGISTRY_LOCATION}"
  echo "-> registry data for auth log: ${REGISTRY_AUTH_LOG}"
  echo "-> registry data for local registry: ${REGISTRY_DATA}"
  echo "-> registry username: ${REGISTRY_USERNAME}"
  echo "-> registry password: ${REGISTRY_PASSWORD}"
  echo ""

  # Build and run cyclone.
  cd ${CYCLONE_ROOT}
  godep go build -race .
  ./scripts/build-worker-image.sh
  docker start clair_clair
  ./cyclone "$@" &
  CYCLONE_PID=$!
  cd - > /dev/null
}

# Clean up local run of cyclone.
function local-cleanup {
  [ -d "${TMPDIR}" ] && rm -rf ${TMPDIR}
  docker-compose -f docker-compose-dev.yml stop
  docker-compose -f docker-compose-dev.yml rm -vf
  [ -n "${CYCLONE_PID-}" ] && ps -p ${CYCLONE_PID} > /dev/null && kill ${CYCLONE_PID}
  ${CYCLONE_ROOT}/scripts/registry/stop.sh
  log "local-up cleanup now."
}
trap local-cleanup INT EXIT

run-local-up

godep go test -v ${CYCLONE_ROOT}/tests/service &&
godep go test -v ${CYCLONE_ROOT}/tests/version &&
godep go test -v ${CYCLONE_ROOT}/tests/yaml
