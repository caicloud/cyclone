#!/bin/bash

output=$(gofmt -d -s pkg)

if [ -n "$output" ]; then
    echo "${output}"
    echo "Go code not formatted, please run 'gofmt -w -s pkg' to fix."
    exit 1
fi