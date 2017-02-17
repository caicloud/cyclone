#!/bin/sh
#
# Copyright 2016 caicloud authors. All rights reserved.

CYCLONE_ROOT=$(dirname "${BASH_SOURCE}")/..

set -e
cd $CYCLONE_ROOT
go test -cover $(go list ./... | grep -v '/vendor/' | grep -v '/tests/')
cd - > /dev/null
