#!/usr/bin/env bash

EXEC_PATH="../../bin/admin"
CLUSTER_NAME="web-78"
KUBE_HOST_ADDR=""
KUBE_CONFIG_PATH="/home/zxq/.kube/config-web-78.yaml"
CLUSTER_ADMIN_HOST="http://192.168.16.78:32233"

${EXEC_PATH} \
    -ENV_CTRL_CLUSTER_NAME="${CLUSTER_NAME}" \
    -ENV_KUBE_HOST="${CLUSTER_NAME}" \
    -ENV_KUBE_CONFIG="${KUBE_CONFIG_PATH}" \
    -ENV_CLUSTER_ADMIN_HOST="${CLUSTER_ADMIN_HOST}" \
    -alsologtostderr=true \
    -ENV_LISTEN_PORT=2333
