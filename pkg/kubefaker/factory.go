/*
Copyright 2016 caicloud authors. All rights reserved.

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

package kubefaker

import (
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/typed/discovery"
	"k8s.io/kubernetes/pkg/client/typed/dynamic"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/runtime"
)

type Factory struct {
	Namespace        string
	EnforceNamespace bool
	clientConfig     *restclient.Config
}

func NewFactory(namespace string, enforce bool, clientConfig restclient.Config) *Factory {

	f := &Factory{
		Namespace:        namespace,
		EnforceNamespace: enforce,
		clientConfig:     &clientConfig,
	}
	return f
}

func (f *Factory) DefaultNamespace() (string, bool) {
	return f.Namespace, f.EnforceNamespace
}

func (f *Factory) UnstructuredClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	cfg := f.clientConfig
	if err := restclient.SetKubernetesDefaults(cfg); err != nil {
		return nil, err
	}
	cfg.APIPath = "/apis"
	if mapping.GroupVersionKind.Group == api.GroupName {
		cfg.APIPath = "/api"
	}
	gv := mapping.GroupVersionKind.GroupVersion()
	cfg.ContentConfig = dynamic.ContentConfig()
	cfg.GroupVersion = &gv
	return restclient.RESTClientFor(cfg)
}

func (f *Factory) DiscoveryClient() (discovery.DiscoveryInterface, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(f.clientConfig)
	if err != nil {
		return nil, err
	}
	return discoveryClient, nil
}

func (f *Factory) UnstructuredObject() (meta.RESTMapper, runtime.ObjectTyper, error) {
	discoveryClient, err := f.DiscoveryClient()
	if err != nil {
		return nil, nil, err
	}

	groupResources, err := discovery.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, nil, err
	}

	mapper := discovery.NewRESTMapper(groupResources, meta.InterfacesForUnstructured)
	typer := discovery.NewUnstructuredObjectTyper(groupResources)

	return mapper, typer, nil
}
