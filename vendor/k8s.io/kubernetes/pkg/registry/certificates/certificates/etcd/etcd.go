/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/rest"
	"k8s.io/kubernetes/pkg/apis/certificates"
	csrregistry "k8s.io/kubernetes/pkg/registry/certificates/certificates"
	"k8s.io/kubernetes/pkg/registry/generic"
	genericregistry "k8s.io/kubernetes/pkg/registry/generic/registry"
	"k8s.io/kubernetes/pkg/runtime"
)

// REST implements a RESTStorage for CertificateSigningRequest against etcd
type REST struct {
	*genericregistry.Store
}

// NewREST returns a registry which will store CertificateSigningRequest in the given helper
func NewREST(optsGetter generic.RESTOptionsGetter) (*REST, *StatusREST, *ApprovalREST) {
	store := &genericregistry.Store{
		NewFunc:     func() runtime.Object { return &certificates.CertificateSigningRequest{} },
		NewListFunc: func() runtime.Object { return &certificates.CertificateSigningRequestList{} },
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return obj.(*certificates.CertificateSigningRequest).Name, nil
		},
		PredicateFunc:     csrregistry.Matcher,
		QualifiedResource: certificates.Resource("certificatesigningrequests"),

		CreateStrategy: csrregistry.Strategy,
		UpdateStrategy: csrregistry.Strategy,
		DeleteStrategy: csrregistry.Strategy,
	}
	options := &generic.StoreOptions{RESTOptions: optsGetter, AttrFunc: csrregistry.GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	// Subresources use the same store and creation strategy, which only
	// allows empty subs. Updates to an existing subresource are handled by
	// dedicated strategies.
	statusStore := *store
	statusStore.UpdateStrategy = csrregistry.StatusStrategy

	approvalStore := *store
	approvalStore.UpdateStrategy = csrregistry.ApprovalStrategy

	return &REST{store}, &StatusREST{store: &statusStore}, &ApprovalREST{store: &approvalStore}
}

// StatusREST implements the REST endpoint for changing the status of a CSR.
type StatusREST struct {
	store *genericregistry.Store
}

func (r *StatusREST) New() runtime.Object {
	return &certificates.CertificateSigningRequest{}
}

// Update alters the status subset of an object.
func (r *StatusREST) Update(ctx api.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}

// ApprovalREST implements the REST endpoint for changing the approval state of a CSR.
type ApprovalREST struct {
	store *genericregistry.Store
}

func (r *ApprovalREST) New() runtime.Object {
	return &certificates.CertificateSigningRequest{}
}

// Update alters the approval subset of an object.
func (r *ApprovalREST) Update(ctx api.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}
