#!/bin/bash

set -o nounset
set -o errexit

RED_COL="\\033[1;31m"           # red color
GREEN_COL="\\033[32;1m"         # green color
YELLOW_COL="\\033[33;1m"        # yellow color
NORMAL_COL="\\033[0;39m"        # normal color

cd $(dirname "${BASH_SOURCE}")

USAGE=$(cat <<-END
Usage: $ ./generate.sh --registry=<registry>/<project> --auth=<user>:<password> --version=<version> --pvc=<pvc> --execNamespace=<execution namespace>
END
)

for i in "$@"
do
case $i in
    --registry=*)
    REGISTRY="${i#*=}"
    shift
    ;;
    --auth=*)
    AUTH="${i#*=}"
    shift
    ;;
    --version=*)
    VERSION="${i#*=}"
    shift
    ;;
    --pvc=*)
    PVC="${i#*=}"
    shift
    ;;
    --execNamespace=*)
    EXEC_NAMESPACE="${i#*=}"
    shift
    ;;
    *)
    echo -e "$RED_COL Unknown parameter $i $NORMAL_COL"
    echo -e "$GREEN_COL $USAGE $NORMAL_COL"
    exit 1
    ;;
esac
done

# Check whether parameters are set.
if [ -z ${REGISTRY+x} ]; then echo -e "$RED_COL Please provide registry with --registry=<registry> $NORMAL_COL"; exit 1; fi
if [ -z ${AUTH+x} ]; then echo -e "$RED_COL Please provide auth with --auth=<auth> $NORMAL_COL"; exit 1; fi
if [ -z ${VERSION+x} ]; then echo -e "$RED_COL Please provide version with --version=<version> $NORMAL_COL"; exit 1; fi
if [ -z ${PVC+x} ]; then echo -e "$RED_COL Please provide PVC with --pvc=<pvc> $NORMAL_COL"; exit 1; fi
if [ -z ${EXEC_NAMESPACE+x} ]; then echo -e "$RED_COL Please provide EXEC_NAMESPACE with --execNamespace=<execution namespace> $NORMAL_COL"; exit 1; fi

if [[ "$OSTYPE" == "linux-gnu" ]]; then
    BASE64_ARGS="-w0"
fi

DOCKER_CONFIG=$(cat <<-END
{
  "auths": {
    "${REGISTRY%/*}": {
        "auth": "$(echo -n $AUTH | base64 ${BASE64_ARGS:-})"
    }
  }
}
END
)

# Generate manifests
if [ ! -e ./.generated ]; then
    mkdir ./.generated
fi
sed -e "s/__REGISTRY__/${REGISTRY/\//\\/}/g" \
    -e "s/__REGISTRY_AUTH__/$(echo $DOCKER_CONFIG | base64 ${BASE64_ARGS:-})/g" \
    -e "s/__PVC__/${PVC}/g" \
    -e "s/__EXE_NAMESPACE__/${EXEC_NAMESPACE}/g" \
    -e "s/__VERSION__/${VERSION}/g" \
    ./cyclone.yaml.template > ./.generated/cyclone.yaml
