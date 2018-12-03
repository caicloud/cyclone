#!/bin/sh
set -e
dockerd-entrypoint.sh dockerd > /dev/null &

# set CERT_DATA
if [ ! -z "$CERT_DATA" ]; then
  echo $CERT_DATA | base64 -d >> /etc/ssl/certs/ca-certificates.crt
fi

sleep 10
/cyclone-worker
