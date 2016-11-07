#!/bin/bash

REGISTRY_DATA="/tmp/cyclone_registry"
REGISTRY_STORAGE_DATA="/tmp/cyclone-registry_storage"
REGISTRY_AUTH_LOG="/tmp/cyclone_auth_log"

# Return 0 if container exists; return non-zero otherwise.
function container-exist {
  docker ps -a | grep "$1" > /dev/null
}

container-exist cyclone-registry && docker rm -f -v cyclone-registry > /dev/null
container-exist cyclone-docker-auth && docker rm -f -v cyclone-docker-auth > /dev/null
rm -rf ${REGISTRY_DATA} ${REGISTRY_AUTH_LOG}
