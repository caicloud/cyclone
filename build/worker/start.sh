#!/bin/sh
set -e
echo "dind parameter is:" $DIND_PARAMETER
dockerd-entrypoint.sh dockerd $DIND_PARAMETER > /dev/null &
sleep 10 
/cyclone-worker
