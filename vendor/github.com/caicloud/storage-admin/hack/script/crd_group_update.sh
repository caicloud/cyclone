#!/usr/bin/env bash

kubectl get storagetypes.storage.resource.caicloud.io -o yaml > _update_tp.yaml
kubectl get storageservices.storage.resource.caicloud.io -o yaml > _update_ss.yaml


sed -i "s/requiredParameters:/serviceParameters:/g" _update_tp.yaml
sed -i "s/optionalParameters:/classParameters:/g" _update_tp.yaml

for f in `ls _update_*.yaml`; do
	sed -i "s/storage.resource.caicloud.io\/v1alpha1/resource.caicloud.io\/v1beta1/g" _update_*.yaml
	kubectl create -f _update_tp.yaml
done
