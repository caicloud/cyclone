#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

function finish {
  echo "sleep 1 second to collect logs"
  sleep 1
}
trap finish EXIT

if [ ! -z ${LOG_COLLECTOR_URL} ]; then
  touch /tmp/log.txt
  /usr/bin/cyclone-toolbox/fstream -f /tmp/log.txt -s ${LOG_COLLECTOR_URL} &
  stdbuf -o0 -e0 /entrypoint.sh "$@" 2>&1 | stdbuf -o0 -e0 -i0 tee /tmp/log.txt
else
  /entrypoint.sh "$@"
fi
