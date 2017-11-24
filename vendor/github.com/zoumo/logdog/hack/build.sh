#!/bin/bash

set -e

command_exists() {
    command -v "$@" > /dev/null 2>&1
}

if ! command_exists glide; then
    go get -u github.com/Masterminds/glide
fi
if ! command_exists glide-vc; then
    go get -u github.com/sgotti/glide-vc
fi

glide install --strip-vendor

glide-vc --use-lock-file --only-code --no-tests

go test -cover $(go list ./... | grep -v '/vendor/' | grep -v '/tests/')

