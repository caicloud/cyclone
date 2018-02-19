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

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/event"
	"github.com/caicloud/cyclone/pkg/store"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
)

// CloudManager represents the interface to manage cloud.
type CloudManager interface {
	CreateCloud(*cloud.Options) (*cloud.Options, error)
	ListClouds() map[string]cloud.Options
	DeleteCloud(name string) error
	PingCloud(name string) error
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
func (m *cloudManager) CreateCloud(cloudOpt *cloud.Options) (*cloud.Options, error) {
	cloudName := cloudOpt.Name

	if _, ok := event.CloudController.GetCloud(cloudName); ok {
		return nil, httperror.ErrorAlreadyExist.Format(cloudName)
	}

	if err := event.CloudController.AddClouds(*cloudOpt); err != nil {
		return nil, err
	}

	gotCloud, _ := event.CloudController.GetCloud(cloudName)
	opts := gotCloud.GetOptions()

	if err := m.ds.InsertCloud(&opts); err != nil {
		return nil, err
	}

	return &opts, nil
}

// ListClouds lists all clouds.
func (m *cloudManager) ListClouds() map[string]cloud.Options {
	clouds := make(map[string]cloud.Options)
	for name, cloud := range event.CloudController.Clouds {
		clouds[name] = cloud.GetOptions()
	}

	return clouds
}

// DeleteCloud deletes the cloud.
func (m *cloudManager) DeleteCloud(name string) error {
	event.CloudController.DeleteCloud(name)

	return m.ds.DeleteCloudByName(name)
}

// PingCloud pings the cloud to check its health.
func (m *cloudManager) PingCloud(name string) error {
	cloud, ok := event.CloudController.GetCloud(name)
	if !ok {
		return httperror.ErrorContentNotFound.Format(name)
	}

	return cloud.Ping()
}
