#!/bin/sh
set -e
sed -i '/2375/d' /usr/local/bin/dockerd-entrypoint.sh
dockerd-entrypoint.sh > /dev/null &
sleep 10 
/cyclone-worker
