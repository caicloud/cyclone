#!/bin/bash

# Copyright 2017 caicloud authors. All rights reserved.

root_path="$(cd "${BASH_SOURCE[0]%/*}"/.. && pwd -P)"

cd ${root_path}

function join() {
	local IFS="$1"
	shift
	echo "$*"
}

rm -rf listerfactory

pkgs=($(find ./vendor/k8s.io/api -type d -mindepth 2 -maxdepth 2 2>/dev/null | sed 's|^\./vendor/||g'))
# pkgs+=($(find ./pkg/apis/ -type d -mindepth 2 -maxdepth 2 2>/dev/null | sed 's|^\./pkg/apis/*|github.com/caicloud/clientset/pkg/apis/|g'))
full_pkgs=$(join "," "${pkgs[@]}")

TEMPBIN=$(mktemp -d)

go build -o ${TEMPBIN}/listerfactory-gen ./cmd/listerfactory-gen

${TEMPBIN}/listerfactory-gen --alsologtostderr --single-directory -p github.com/caicloud/clientset/listerfactory --versioned-clientset-package k8s.io/client-go/kubernetes -i $full_pkgs
