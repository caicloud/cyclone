/*
Copyright 2017 caicloud authors. All rights reserved.

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

package manager

import (
	"fmt"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/cloud"
	"github.com/caicloud/cyclone/pkg/store"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
)

// CloudManager represents the interface to manage cloud.
type CloudManager interface {
	CreateCloud(*api.Cloud) (*api.Cloud, error)
	ListClouds() ([]api.Cloud, error)
	DeleteCloud(name string) error
	PingCloud(name string) error
	ListWorkers(name string, extendInfo string) ([]api.WorkerInstance, error)
}

// cloudManager represents the manager for cloud.
type cloudManager struct {
	ds *store.DataStore
}

// NewCloudManager creates a cloud manager.
func NewCloudManager(dataStore *store.DataStore) (CloudManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new cloud manager as data store is nil.")
	}

	return &cloudManager{dataStore}, nil
}

// CreateCloud creates a cloud.
func (m *cloudManager) CreateCloud(cloud *api.Cloud) (*api.Cloud, error) {
	cloudName := cloud.Name

	if _, err := m.ds.FindCloudByName(cloudName); err == nil {
		return nil, httperror.ErrorAlreadyExist.Format(cloudName)
	}

	if err := m.ds.InsertCloud(cloud); err != nil {
		return nil, err
	}

	return cloud, nil
}

// ListClouds lists all clouds.
func (m *cloudManager) ListClouds() ([]api.Cloud, error) {
	return m.ds.FindAllClouds()
}

// DeleteCloud deletes the cloud.
func (m *cloudManager) DeleteCloud(name string) error {
	return m.ds.DeleteCloudByName(name)
}

// PingCloud pings the cloud to check its health.
func (m *cloudManager) PingCloud(name string) error {
	c, err := m.ds.FindCloudByName(name)
	if err != nil {
		return httperror.ErrorContentNotFound.Format(name)
	}
	cp, err := cloud.NewCloudProvider(c)

	return cp.Ping()
}

// ListWorkers lists all workers.
func (m *cloudManager) ListWorkers(name string, extendInfo string) ([]api.WorkerInstance, error) {
	c, err := m.ds.FindCloudByName(name)
	if err != nil {
		return nil, httperror.ErrorContentNotFound.Format(name)
	}

	if c.Type == api.CloudTypeKubernetes && extendInfo == "" {
		err := fmt.Errorf("query parameter namespace can not be empty because cloud type is %v.",
			api.CloudTypeKubernetes)
		return nil, err
	}

	if c.Kubernetes != nil {
		if c.Kubernetes.InCluster {
			// default cluster, get default namespace.
			c.Kubernetes.Namespace = cloud.DefaultNamespace
		} else {
			c.Kubernetes.Namespace = extendInfo
		}
	}

	cp, err := cloud.NewCloudProvider(c)

	return cp.ListWorkers()
}
