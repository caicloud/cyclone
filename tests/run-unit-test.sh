#!/bin/bash
#
# Copyright 2016 caicloud authors. All rights reserved.

CYCLONE_ROOT=$(dirname "${BASH_SOURCE}")/..

cd $CYCLONE_ROOT
godep go test -cover $(go list ./... | grep -v '/vendor/' | grep -v '/tests/')
cd - > /dev/null
