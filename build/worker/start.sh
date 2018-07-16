#!/bin/sh
set -e
dockerd-entrypoint.sh dockerd > /dev/null &
sleep 10 
/cyclone-worker
