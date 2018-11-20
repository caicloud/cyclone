#!/usr/bin/env bash

# kubernetes 版本 1.7 的 deepcopy 格式不同，此处脚本暂停使用

# EXEC_CMD="go run ${GOPATH}/src/k8s.io/code-generator/cmd/deepcopy-gen/main.go"
# EXEC_CMD="go run ${GOPATH}/src/k8s.io/kubernetes/staging/src/k8s.io/code-generator/cmd/deepcopy-gen/main.go"
# EXEC_CMD="${GOPATH}/src/k8s.io/kubernetes/_output/bin/deepcopy-gen"

# ${EXEC_CMD} \
#     --input-dirs="github.com/caicloud/storage-admin/pkg/kubernetes/apis/resource/v1alpha1" \
#     --bounding-dirs="github.com/caicloud/storage-admin/pkg/kubernetes/apis/resource/v1alpha1" \
#     --output-file-base="zz_generated.deepcopy" \
#     --logtostderr=true

# ${EXEC_CMD} \
#     --input-dirs="github.com/caicloud/storage-admin/pkg/kubernetes/apis/resource/v1beta1" \
#     --bounding-dirs="github.com/caicloud/storage-admin/pkg/kubernetes/apis/resource/v1beta1" \
#     --output-file-base="zz_generated.deepcopy" \
#     --logtostderr=true

# ${EXEC_CMD} \
#     --input-dirs="github.com/caicloud/storage-admin/pkg/kubernetes/apis/tenant/v1alpha1" \
#     --bounding-dirs="github.com/caicloud/storage-admin/pkg/kubernetes/apis/tenant/v1alpha1" \
#     --output-file-base="zz_generated.deepcopy" \
#     --logtostderr=true
