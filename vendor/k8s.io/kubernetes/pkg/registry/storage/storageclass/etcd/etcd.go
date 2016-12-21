/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package etcd

import (
	storageapi "k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/pkg/registry/generic"
	genericregistry "k8s.io/kubernetes/pkg/registry/generic/registry"
	"k8s.io/kubernetes/pkg/registry/storage/storageclass"
	"k8s.io/kubernetes/pkg/runtime"
)

type REST struct {
	*genericregistry.Store
}

// NewREST returns a RESTStorage object that will work against persistent volumes.
func NewREST(optsGetter generic.RESTOptionsGetter) *REST {
	store := &genericregistry.Store{
		NewFunc:     func() runtime.Object { return &storageapi.StorageClass{} },
		NewListFunc: func() runtime.Object { return &storageapi.StorageClassList{} },
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*storageapi.StorageClass).Name, nil
		},
		PredicateFunc:     storageclass.MatchStorageClasses,
		QualifiedResource: storageapi.Resource("storageclasses"),

		CreateStrategy:      storageclass.Strategy,
		UpdateStrategy:      storageclass.Strategy,
		DeleteStrategy:      storageclass.Strategy,
		ReturnDeletedObject: true,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: storageclass.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	return &REST{store}
}
