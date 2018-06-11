/*
Copyright 2018 caicloud authors. All rights reserved.

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

package cloud

import (
	"fmt"
	"io/ioutil"
	"time"

	log "github.com/golang/glog"
	mgo "gopkg.in/mgo.v2"
	apiv1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/store"
)

const (
	DefaultCloudPingTimeout = time.Duration(5 * time.Second)

	WorkerTimeout = time.Duration(2 * time.Hour)

	defaultCloudName = "inCluster"
)

type newCloudFunc func(c *api.Cloud) (Provider, error)

var cloudProviderFactory map[api.CloudType]newCloudFunc

func init() {
	cloudProviderFactory = make(map[api.CloudType]newCloudFunc)
}

func RegistryCloudProvider(ct api.CloudType, ncf newCloudFunc) error {
	if _, ok := cloudProviderFactory[ct]; ok {
		return fmt.Errorf("cloud provider %s has been registried", ct)
	}

	cloudProviderFactory[ct] = ncf

	return nil
}

func InitCloud(autoDiscovery bool) error {
	ds := store.NewStore()
	defer ds.Close()

	_, err := ds.FindCloudByName(defaultCloudName)
	if err != nil && err == mgo.ErrNotFound {
		if autoDiscovery {
			log.Info("add the default incluster cloud %s", defaultCloudName)

			namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/" + apiv1.ServiceAccountNamespaceKey)
			if err != nil {
				log.Error(err)
				return err
			}

			defaultK8sCloud := &api.Cloud{
				Name: defaultCloudName,
				Type: api.CloudTypeKubernetes,
				Kubernetes: &api.CloudKubernetes{
					InCluster: true,
					Namespace: string(namespace),
				},
			}

			return ds.InsertCloud(defaultK8sCloud)
		}
	}

	return nil
}

type Provider interface {
	CanProvision(quota options.Quota) (bool, error)
	Resource() (*options.Resource, error)
	Provision(info *api.WorkerInfo, opts *options.WorkerOptions) (*api.WorkerInfo, error)
	TerminateWorker(string) error
	Ping() error
}

func NewCloudProvider(c *api.Cloud) (Provider, error) {
	ncf, ok := cloudProviderFactory[c.Type]
	if !ok {
		err := fmt.Errorf("cloud type %s is not supported", c.Type)
		log.Error(err)
		return nil, err
	}

	return ncf(c)
}
