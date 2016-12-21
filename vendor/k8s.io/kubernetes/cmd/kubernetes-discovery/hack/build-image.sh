#!/bin/bash

# Copyright 2014 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


KUBE_ROOT=$(dirname "${BASH_SOURCE}")/../../..
source "${KUBE_ROOT}/hack/lib/util.sh"

# Register function to be called on EXIT to remove generated binary.
function cleanup {
  rm "${KUBE_ROOT}/cmd/kubernetes-discovery/artifacts/simple-image/kubernetes-discovery"
}
trap cleanup EXIT

cp -v ${KUBE_ROOT}/_output/local/bin/linux/amd64/kubernetes-discovery "${KUBE_ROOT}/cmd/kubernetes-discovery/artifacts/simple-image/kubernetes-discovery"
docker build -t kubernetes-discovery:latest ${KUBE_ROOT}/cmd/kubernetes-discovery/artifacts/simple-image
