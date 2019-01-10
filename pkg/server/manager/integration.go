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

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/integrate"
	"github.com/caicloud/cyclone/pkg/store"
	"github.com/caicloud/cyclone/pkg/util/http/errors"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	slug "github.com/caicloud/cyclone/pkg/util/slugify"
)

// IntegrationManager represents the interface to manage integration.
type IntegrationManager interface {
	CreateIntegration(integration *api.Integration) (*api.Integration, error)
	ListIntegrations() ([]api.Integration, error)
	DeleteIntegration(name string) error
	UpdateIntegration(name string, ni *api.Integration) (*api.Integration, error)
	GetIntegration(name string) (*api.Integration, error)
}

// integrationManager represents the manager for integration.
type integrationManager struct {
	ds *store.DataStore
}

// NewIntegrationManager creates an integration manager.
func NewIntegrationManager(dataStore *store.DataStore) (IntegrationManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new integration manager as data store is nil.")
	}

	return &integrationManager{dataStore}, nil
}

// CreateIntegration creates an integration.
func (m *integrationManager) CreateIntegration(i *api.Integration) (*api.Integration, error) {
	if i.Name == "" && i.Alias == "" {
		return nil, httperror.ErrorValidationFailed.Error("integration name and alias", "can not neither be empty")
	}

	nameEmpty := false
	if i.Name == "" && i.Alias != "" {
		i.Name = slug.Slugify(i.Alias, false, -1)
		nameEmpty = true
	}

	if ei, err := m.ds.GetIntegration(i.Name); err == nil {
		log.Errorf("name %s conflict, integration alias:%s, exist integration alias:%s",
			i.Name, i.Alias, ei.Alias)
		if nameEmpty {
			i.Name = slug.Slugify(i.Name, true, -1)
		} else {
			return nil, httperror.ErrorAlreadyExist.Error(i.Name)
		}
	}

	// check auth info
	valid, err := integrate.Validate(i)
	if err != nil || !valid {
		log.Errorf("Valid not pass, valid:%s, error:%v", valid, err)
		return nil, errors.ErrorAuthenticationFailed.Error()
	}

	if err := m.ds.InsertIntegration(i); err != nil {
		return nil, err
	}

	return i, nil
}

// UpdateIntegration updates an integration.
func (m *integrationManager) UpdateIntegration(name string, ni *api.Integration) (*api.Integration, error) {
	i, err := m.ds.GetIntegration(ni.Name)
	if err != nil {
		return nil, err
	}

	if ni.Alias != "" {
		i.Alias = ni.Alias
	}

	if ni.SonarQube != nil {
		i.SonarQube = ni.SonarQube
	}

	// check auth info
	valid, err := integrate.Validate(i)
	if err != nil || !valid {
		log.Errorf("Valid not pass, valid:%s, error:%v", valid, err)
		return nil, errors.ErrorAuthenticationFailed.Error()
	}

	if err = m.ds.UpdateIntegration(i); err != nil {
		return nil, err
	}

	return i, nil
}

// ListIntegrations lists all integrations.
func (m *integrationManager) ListIntegrations() ([]api.Integration, error) {
	return m.ds.FindAllIntegrations()
}

// DeleteIntegration deletes the integration.
func (m *integrationManager) DeleteIntegration(name string) error {
	return m.ds.DeleteIntegrationByName(name)
}

// GetIntegration gets the pipeline record by name.
func (m *integrationManager) GetIntegration(name string) (*api.Integration, error) {
	return m.ds.GetIntegration(name)
}
