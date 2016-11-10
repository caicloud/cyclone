#!/bin/sh

curl -X GET --header 'Accept: application/json' "http://localhost:7099/api/v0.1/fake-user-id/versions/$1"
