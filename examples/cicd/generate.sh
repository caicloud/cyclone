#!/bin/bash

set -o nounset
set -o errexit

RED_COL="\\033[1;31m"           # red color
GREEN_COL="\\033[32;1m"         # green color
YELLOW_COL="\\033[33;1m"        # yellow color
NORMAL_COL="\\033[0;39m"        # normal color

cd $(dirname "${BASH_SOURCE}")

USAGE=$(cat <<-END
Usage: $ ./generate.sh --registry=<registry>/<project>
END
)

for i in "$@"
do
case $i in
    --registry=*)
    REGISTRY="${i#*=}"
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

# Generate manifests
if [ ! -e ./.generated ]; then
    mkdir ./.generated
fi
sed -e "s/__REGISTRY__/${REGISTRY/\//\\/}/g" ./golang/manifests.yaml.template > ./.generated/golang.yaml
sed -e "s/__REGISTRY__/${REGISTRY/\//\\/}/g" ./java/manifests.yaml.template > ./.generated/java.yaml
sed -e "s/__REGISTRY__/${REGISTRY/\//\\/}/g" ./nodejs/manifests.yaml.template > ./.generated/nodejs.yaml
