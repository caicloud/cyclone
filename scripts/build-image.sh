#!/bin/bash
#
# The script builds cyclone component container, see usage function for how to run
# the script. After build completes, following container will be built, i.e.
#   caicloud/cyclone:${IMAGE_TAG}
#
# By default, IMAGE_TAG is latest.

set -e
set -u
set -o pipefail

source "$(dirname "${BASH_SOURCE}")/lib/common.sh"

ROOT=$(dirname "${BASH_SOURCE}")/..

function usage {
  echo -e "Usage:"
  echo -e "  ./build-image.sh [tag]"
  echo -e ""
  echo -e "Parameter:"
  echo -e " tag\tDocker image tag, treated as cyclone release version. If provided,"
  echo -e "    \tthe tag must be the form of vA.B.C, where A, B, C are digits, e.g."
  echo -e "    \tv1.0.1. If not provided, it will build images with tag 'latest'"
  echo -e ""
  echo -e "Environment variable:"
  echo -e " PUSH_TO_REGISTRY     \tPush images to caicloud registry or not, options: Y or N. Default value: ${PUSH_TO_REGISTRY}"
}
# -----------------------------------------------------------------------------
# Parameters for building docker image, see usage.
# -----------------------------------------------------------------------------
# Decide if we need to push the new images to caicloud registry.
PUSH_TO_REGISTRY=${PUSH_TO_REGISTRY:-"N"}

# Find image tag version, the tag is considered as release version.
if [[ "$#" == "1" ]]; then
  if [[ "$1" == "help" || "$1" == "--help" || "$1" == "-h" ]]; then
    usage
    exit 0
  elif [[ ! $1 =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "Error: image tag format error, see usage."
    echo -e ""
    usage
    exit 1
  else
    IMAGE_TAG=${1}
  fi
else
  IMAGE_TAG="latest"
fi

# -----------------------------------------------------------------------------
# Start Building containers
# -----------------------------------------------------------------------------
# Setup docker on Mac.
if [[ "$(uname)" == "Darwin" ]]; then
  if [[ "$(which docker-machine)" != "" ]]; then
    eval "$(docker-machine env kube-dev)"
  elif [[ "$(which boot2docker)" != "" ]]; then
    eval "$(boot2docker shellinit)"
  fi
fi

echo "+++++ Start building cyclone server"

cd ${ROOT}

# Build cyclone comtainer.
# We need to disable selinux on selinux supported system to relove https://github.com/caicloud/cyclone/issues/53
readonly platform=$(os::build::host_platform)
if [[ $platform != *"darwin"* ]]; then
  echo "disable the selinux"
  disable-selinux
fi

docker run --rm \
  -v `pwd`:/go/src/github.com/caicloud/cyclone \
  -e GOPATH=/go:/go/src/github.com/caicloud/cyclone/vendor golang:1.6-alpine sh \
  -c "cd /go/src/github.com/caicloud/cyclone && go build -o cyclone-server"

docker build -t caicloud/cyclone-server:${IMAGE_TAG} .
docker tag caicloud/cyclone-server:${IMAGE_TAG} cargo.caicloud.io/caicloud/cyclone-server:${IMAGE_TAG}

cd - > /dev/null

# Decide if we need to push images to docker hub.
if [[ "$PUSH_TO_REGISTRY" == "Y" ]]; then
  echo ""
  echo "+++++ Start pushing cyclone-server"
  docker push cargo.caicloud.io/caicloud/cyclone-server:${IMAGE_TAG}
fi

echo "Successfully built docker image caicloud/cyclone-server:${IMAGE_TAG}"
echo "Successfully built docker image cargo.caicloud.io/caicloud/cyclone-server:${IMAGE_TAG}"

# A reminder for creating Github release.
if [[ "$#" == "1" && $1 =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo -e "Finish building release ; if this is a formal release, please remember"
  echo -e "to create a release tag at Github at: https://github.com/caicloud/cyclone-server/releases"
fi
