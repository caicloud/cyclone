#!/usr/bin/env bash

# crd client generate 移动到公共区域，这里的脚本不再生成线上用客户端
# 因公共区域暂时没有生成 fake 客户端，有在本 repo 保留生成脚本的必要
# 因公共客户端结构问题，此处的 fake 客户端需要额外修改以适应，故不再直接使用生成代码，此处脚本留作参考

# EXEC_CMD="go run ${GOPATH}/src/k8s.io/code-generator/cmd/client-gen/main.go"
# EXEC_CMD="go run ${GOPATH}/src/k8s.io/kubernetes/staging/src/k8s.io/code-generator/cmd/client-gen/main.go"
# EXEC_CMD="go run ${GOPATH}/src/github.com/caicloud/clientset/cmd/client-gen/main.go"

# rm -rf ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/clients/*

# ${EXEC_CMD} \
#     --input-base="github.com/caicloud/storage-admin/pkg/kubernetes/apis" \
#     --input="resource/v1beta1" \
#     --clientset-path="github.com/caicloud/storage-admin/pkg/kubernetes/clients" \
#     --clientset-name="resource" \
#     --fake-clientset=true

# wait for bug fix
# generator 现在有 bug，生成的 fake 部分的 group 变成简单的从 --input 中抓取，但这个值不正确，正常应该从代码中读取
# https://github.com/kubernetes/kubernetes/issues/53498
# 后续会从 tag 中读取，但这个 pr 貌似还没合并
# https://github.com/kubernetes/kubernetes/pull/53579
# 暂时先用 sed 处理下

#rm -rf ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/*.*
#rm -rf ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/scheme
#rm -r -f ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/typed/resource/v1alpha1/*.*
#
#sed -i "s/{Group:\ \"resource\",\ Version:\ \"v1alpha1\",\ /{Group:\ v1alpha1.SchemeGroupVersion.Group,\ Version:\ v1alpha1.SchemeGroupVersion.Version,\ /g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/typed/resource/v1alpha1/fake/fake_storage*.go
#
#sed -i "s/scheme\ \"k8s.io/kubescheme\ \"k8s.io/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/fake/register.go
#sed -i "s/\ scheme\./\ kubescheme\./g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/fake/register.go
#
#sed -i "s/github.com\/caicloud\/storage-admin\/pkg\/kubernetes\/fake\/resource\/typed\/resource\/v1alpha1\"/github.com\/caicloud\/clientset\/kubernetes\/typed\/resource\/v1alpha1\"/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/fake/clientset_generated.go
#sed -i "s/var\ _\ clientset/\/\/var\ _\ clientset/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/fake/clientset_generated.go
#sed -i "s/clientset\ \"/\/\/clientset\ \"/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/fake/clientset_generated.go
#
#sed -i "s/github.com\/caicloud\/storage-admin\/pkg\/kubernetes\/fake\/resource\/typed\/resource\/v1alpha1\"/github.com\/caicloud\/clientset\/kubernetes\/typed\/resource\/v1alpha1\"/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/resource/typed/resource/v1alpha1/fake/fake_resource_client.go


# ${EXEC_CMD} \
#     --input-base="github.com/caicloud/storage-admin/pkg/kubernetes/apis" \
#     --input="tenant/v1alpha1" \
#     --clientset-path="github.com/caicloud/storage-admin/pkg/kubernetes/clients" \
#     --clientset-name="tenant" \
#     --fake-clientset=true

#rm -rf ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/*.*
#rm -rf ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/scheme
#rm -r -f ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/typed/tenant/v1alpha1/*.*
#
#sed -i "s/{Group:\ \"tenant\",\ Version:\ \"v1alpha1\",\ /{Group:\ v1alpha1.SchemeGroupVersion.Group,\ Version:\ v1alpha1.SchemeGroupVersion.Version,\ /g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/typed/tenant/v1alpha1/fake/fake_*.go
#
#sed -i "s/scheme\ \"k8s.io/kubescheme\ \"k8s.io/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/fake/register.go
#sed -i "s/\ scheme\./\ kubescheme\./g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/fake/register.go
#
#sed -i "s/github.com\/caicloud\/storage-admin\/pkg\/kubernetes\/fake\/tenant\/typed\/tenant\/v1alpha1\"/github.com\/caicloud\/tenant-admin\/kubernetes\/typed\/tenant\/v1alpha1\"/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/fake/clientset_generated.go
#sed -i "s/var\ _\ clientset/\/\/var\ _\ clientset/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/fake/clientset_generated.go
#sed -i "s/clientset\ \"/\/\/clientset\ \"/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/fake/clientset_generated.go
#
#sed -i "s/github.com\/caicloud\/storage-admin\/pkg\/kubernetes\/fake\/tenant\/typed\/tenant\/v1alpha1\"/github.com\/caicloud\/tenant-admin\/kubernetes\/typed\/tenant\/v1alpha1\"/g" \
#    ${GOPATH}/src/github.com/caicloud/storage-admin/pkg/kubernetes/fake/tenant/typed/tenant/v1alpha1/fake/fake_tenant_client.go


# ${EXEC_CMD} \
#     --input-base="k8s.io/apiextensions-apiserver/pkg/apis" \
#     --input="apiextensions/v1beta1" \
#     --clientset-path="github.com/caicloud/storage-admin/pkg/kubernetes/clients" \
#     --clientset-name="apiextensions" \
#     --fake-clientset=true
