#!/usr/bin/env bash

EXEC_PATH="../../bin/controller-sub"
KUBE_HOST="https://localhost:6443"
KUBE_CONFIG="/var/run/kubernetes/admin.kubeconfig"
CONTROLLER_MAX_RETRY_TIMES="0"

${EXEC_PATH} \
    -ENV_KUBE_HOST="${KUBE_HOST}" \
    -ENV_KUBE_CONFIG="${KUBE_CONFIG}" \
    -ENV_CONTROLLER_MAX_RETRY_TIMES="${CONTROLLER_MAX_RETRY_TIMES}" \
    -alsologtostderr=true
