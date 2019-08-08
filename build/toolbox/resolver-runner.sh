#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail

if [ ! -z ${LOG_COLLECTOR_URL} ]; then
  touch /tmp/log.txt
  /usr/bin/cyclone-toolbox/fstream -f /tmp/log.txt -s ${LOG_COLLECTOR_URL} &
  /entrypoint.sh "$@" 2>&1 | tee /tmp/log.txt
else
  /entrypoint.sh "$@"
fi