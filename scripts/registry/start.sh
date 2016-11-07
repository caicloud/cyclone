#!/bin/bash

CYCLONE_ROOT=$(dirname "${BASH_SOURCE}")/../..
REGISTRY_DATA=${REGISTRY_DATA:-"/tmp/cyclone-registry"}
REGISTRY_AUTH_LOG=${REGISTRY_AUTH_LOG:-"/tmp/cyclone-auth-log"}

# Check if image exists:
#
# Input:
#   $1 image name
#   $2 image tag
function docker-image-exist {
  docker images | grep $1 | grep $2 > /dev/null
}

# Make sure we have required repositories.
function pull-repositories {
  if ! docker-image-exist "caicloud/registry" "v0.3.0"; then
    docker pull caicloud/registry:v0.3.0
  fi
  if ! docker-image-exist "caicloud/docker_auth" "v2.1.0"; then
    docker pull caicloud/docker_auth:v2.1.0
  fi
}

# Start registry listening on 5000. Images are stored under /tmp/registry
# and will be deleted via 'stop.sh' script.
function start-registry {
  echo "Start registry"
  cd ${CYCLONE_ROOT}
  rm -rf ${REGISTRY_DATA} && mkdir -p ${REGISTRY_DATA}
  docker run \
         --name cyclone-registry \
         --restart=always \
         -v ${REGISTRY_DATA}:/var/lib/registry \
         -v `pwd`/scripts/registry/config-registry.yml:/etc/docker/registry/config.yml \
         -v `pwd`/scripts/registry/ssl:/etc/docker/registry/ssl \
         -p 5000:5000 \
         -d caicloud/registry:v0.3.0 > /dev/null
  cd - > /dev/null
}

# Start docker-auth, configured to be listening on port 3000. Docker auth
# logs are stored under /tmp/docker_auth_log and will be deleted via 'stop.sh'
# script.
function start-docker-auth {
  echo "Start docker-auth"
  cd ${CYCLONE_ROOT}
  rm -rf ${REGISTRY_AUTH_LOG} && mkdir -p ${REGISTRY_AUTH_LOG}
  docker run \
         --name cyclone-docker-auth \
         -p 3000:3000 \
         -v ${REGISTRY_AUTH_LOG}:/logs \
         -v `pwd`/scripts/registry:/config:ro \
         -v `pwd`/scripts/registry/ssl:/etc/docker-auth/ssl:ro \
         -d caicloud/docker_auth:v2.1.0 \
         --v=2 --alsologtostderr /config/config-auth.yml > /dev/null
  cd - > /dev/null
}

pull-repositories
start-registry
start-docker-auth
