#!/bin/bash

# Script that runs tests, code coverage, and benchmarks all at once.

GVAL_PATH=$HOME/gopath/src/github.com/PaesslerAG/gval

# run the actual tests.
cd "${GVAL_PATH}"
go test -bench=. -benchmem -timeout 10m -coverprofile coverage.out
status=$?

if [ "${status}" != 0 ];
then
	exit $status
fi

# run random test for a longer period.
go test -bench=Random -benchtime 5m -timeout 30m -benchmem -coverprofile coverage.out
status=$?

if [ "${status}" != 0 ];
then
	exit $status
fi

