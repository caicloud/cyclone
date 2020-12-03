#!/usr/bin/env bash

# The script auto-generate kubernetes clients, listers, informers

set -e

ORIGIN=$(pwd)
cd $(dirname ${BASH_SOURCE[0]})

# Add your packages here
PKGS=(cyclone/v1alpha1)

CLIENT_PATH=github.com/caicloud/cyclone/pkg
CLIENT_APIS=${CLIENT_PATH}/apis

# use the relative path as output base path, we need the starting path of the module,
# from xx/github.com/caicloud/module-template/hack to xx (../../../../)
# we are not using ../ because if the ORIGIN path is soft link (this will happen in CI), use ../ will get the realpath dirname
OUTPUT_BASE=$(dirname $(dirname $(dirname $(dirname $(pwd)))))

for path in $PKGS
do
	ALL_PKGS="$CLIENT_APIS/$path "$ALL_PKGS
done

function join {
	local IFS="$1"
   	shift
   	result="$*"
}

join "," ${PKGS[@]}
PKGS=$result

join "," $ALL_PKGS
FULL_PKGS=$result

BINS=(
  client-gen
  conversion-gen
  deepcopy-gen
  defaulter-gen
  informer-gen
  lister-gen
)

TEMPBIN=./tmpbin

unset GOOS GOARCH

mkdir -p $TEMPBIN
for bin in "${BINS[@]}"
do
	go build -mod=vendor -o $TEMPBIN/$bin ../vendor/k8s.io/code-generator/cmd/$bin
done

echo "Generating conversions"
${TEMPBIN}/conversion-gen \
  --output-base ${OUTPUT_BASE} --input-dirs ${FULL_PKGS} -O zz_generated.conversion --go-header-file boilerplate/boilerplate.go.txt

echo "Generating defaulters"
${TEMPBIN}/defaulter-gen \
  --output-base ${OUTPUT_BASE} --input-dirs ${FULL_PKGS} -O zz_generated.defaults --go-header-file boilerplate/boilerplate.go.txt

echo "Generating deepcopy funcs"
${TEMPBIN}/deepcopy-gen \
  --output-base ${OUTPUT_BASE} --input-dirs ${FULL_PKGS} -O zz_generated.deepcopy --bounding-dirs "${CLIENT_APIS}" --go-header-file boilerplate/boilerplate.go.txt

echo "Generating clientset for ${PKGS} at ${CLIENT_PATH}/k8s/clientset"
${TEMPBIN}/client-gen \
  --output-base ${OUTPUT_BASE} --clientset-name "clientset" --input-base "" --input ${FULL_PKGS} --output-package "${CLIENT_PATH}/k8s" --go-header-file boilerplate/boilerplate.go.txt

echo "Generating listers for ${PKGS} at ${CLIENT_PATH}/k8s/listers"
${TEMPBIN}/lister-gen \
  --output-base ${OUTPUT_BASE} --input-dirs ${FULL_PKGS} --output-package "${CLIENT_PATH}/k8s/listers" --go-header-file boilerplate/boilerplate.go.txt

echo "Generating informers for ${PKGS} at ${CLIENT_PATH}/k8s/informers"
${TEMPBIN}/informer-gen \
  --output-base ${OUTPUT_BASE} --single-directory \
  --input-dirs ${FULL_PKGS} \
  --versioned-clientset-package "${CLIENT_PATH}/k8s/clientset" \
  --listers-package "${CLIENT_PATH}/k8s/listers" \
  --output-package "${CLIENT_PATH}/k8s/informers" --go-header-file boilerplate/boilerplate.go.txt

rm -rf $TEMPBIN

cd $ORIGIN
