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
	"gopkg.in/mgo.v2"
	apiv1 "k8s.io/api/core/v1"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/store"
	"github.com/caicloud/cyclone/pkg/util/http/errors"
)

const (
	DefaultCloudPingTimeout = time.Duration(5 * time.Second)

	WorkerTimeout = time.Duration(2 * time.Hour)

	DefaultCloudName = "_inCluster"

	// RegistryCertPath is the path of docker registry cert file
	// that used for cyclone worker to pull images.
	//
	// Cyclone server will read the cert contents from RegistryCertPath,
	// and set it to the Env whose key is ENV_CERT_DATA in cyclone worker.
	RegistryCertPath string = "/tmp/certs/registry.crt"

	// ENV_CERT_DATA is a environment name of cert path
	// that used for cyclone worker to pull images.
	ENV_CERT_DATA string = "CERT_DATA"
)

var DefaultNamespace = "default"

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

	_, err := ds.FindCloudByName(DefaultCloudName)
	if err != nil && err == mgo.ErrNotFound {
		if autoDiscovery {
			log.Info("add the default incluster cloud %s", DefaultCloudName)

			namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/" + apiv1.ServiceAccountNamespaceKey)
			if err != nil {
				log.Error(err)
				// error when running cyclone not in a k8s system, e.g. local docker system.
				//return err
			}

			DefaultNamespace = string(namespace)

			defaultK8sCloud := &api.Cloud{
				Name: DefaultCloudName,
				Type: api.CloudTypeKubernetes,
				Kubernetes: &api.CloudKubernetes{
					InCluster: true,
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
	ListWorkers() ([]api.WorkerInstance, error)
}

func NewCloudProvider(c *api.Cloud) (Provider, error) {
	if c == nil {
		err := fmt.Errorf("Cloud config is nil")
		log.Error(err)
		return nil, err
	}

	ncf, ok := cloudProviderFactory[c.Type]
	if !ok {
		err := errors.ErrorUnsupported.Error("cloud type", c.Type)
		log.Error(err)
		return nil, err
	}

	return ncf(c)
}
