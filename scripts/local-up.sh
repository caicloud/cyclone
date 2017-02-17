#!/bin/bash
#
# Copyright 2015 caicloud authors. All rights reserved.

# The script starts a local cyclone, mongodb, registry.
#
# Usage:
#   ./local-up.sh [docker endpint]

set -e
set -o nounset
set -o pipefail

# Timestamped log, e.g. log "cluster created".
#
# Input:
#   $1 Log string.
function log {
  echo -e "[`TZ=Asia/Shanghai date`] ${1}"
}

CYCLONE_ROOT=$(dirname "${BASH_SOURCE}")/..
TMPDIR=`mktemp -d /tmp/cyclone.XXXXXXXXXX`
REGISTRY_DATA=${TMPDIR}/registry
REGISTRY_AUTH_LOG=${TMPDIR}/auth-log

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

if [[ "$(which docker)" == "" ]]; then
  echo "Unable to find docker"
fi
mkdir ${REGISTRY_DATA} ${REGISTRY_AUTH_LOG}

log "Mongo, Kafka, and the registry are all running in a docker container, cyclone running in local."
cd ${CYCLONE_ROOT}
docker-compose -f docker-compose-dev.yml up -d --force-recreate
cd - > /dev/null

# Run local registry.
${CYCLONE_ROOT}/scripts/registry/start.sh

# Export env variables for cyclone.
export DOCKER_HOST=${DOCKER_HOST:-"unix:///var/run/docker.sock"}
export ENABLE_CAICLOUD_AUTH=${ENABLE_CAICLOUD_AUTH:-"false"}
export REGISTRY_LOCATION=${REGISTRY_LOCATION:-"cargo.caicloud.io"}
export REGISTRY_USERNAME=${REGISTRY_USERNAME:-""}
export REGISTRY_PASSWORD=${REGISTRY_PASSWORD:-""}
export WORK_REGISTRY_LOCATION=${WORK_REGISTRY_LOCATION:-"cargo.caicloud.io"}
export SUCCESSTEMPLATE=${SUCCESSTEMPLATE:-"./notify/provider/success.html"}
export ERRORTEMPLATE=${ERRORTEMPLATE:-"./notify/provider/error.html"}
export WORKER_NODE_DOCKER_VERSION=${WORKER_NODE_DOCKER_VERSION:-"1.10.1"}

# Static configs.
export REGISTRY_AUTH_LOG=${REGISTRY_AUTH_LOG}
export REGISTRY_DATA=${REGISTRY_DATA}
export MONGO_DB_IP="127.0.0.1:28017"
export KAFKA_SERVER_IP="127.0.0.1:9092"
export CLAIR_SERVER_IP="127.0.0.1:6060"
export ETCD_SERVER_IP="http://127.0.0.1:2379"
export CYCLONE_SERVER_HOST="http://127.0.0.1:7099"
export LOG_SERVER="ws://127.0.0.1:8000/ws"
export WORKER_IMAGE="cargo.caicloud.io/caicloud/cyclone-worker"
export CLIENTID=""
export CLIENTIDSECRET=""
export CLIENTID_GITLAB=0
export CLIENTIDSECRET_GITLAB=0
export SERVER_GITLAB=https://gitlab.com
export CONSOLE_WEB_ENDPOINT="http://127.0.0.1:7099"

echo "-> registry location: ${REGISTRY_LOCATION}"
echo "-> registry data for auth log: ${REGISTRY_AUTH_LOG}"
echo "-> registry data for local registry: ${REGISTRY_DATA}"
echo "-> registry username: ${REGISTRY_USERNAME}"
echo "-> registry password: ${REGISTRY_PASSWORD}"
echo ""

# Build and run cyclone.
cd ${CYCLONE_ROOT}
godep go build -race .
docker start clair_clair
./cyclone "$@" &
CYCLONE_PID=$!
cd - > /dev/null

while true; do sleep 1; done
