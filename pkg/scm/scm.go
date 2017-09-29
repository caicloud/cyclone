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

package scm

import (
	"fmt"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm/provider"
	"github.com/caicloud/cyclone/store"
)

// SCM is the interface of all operations needed for scm repository.
type SCM interface {
	Authcallback(code, state string) (string, error)
	GetRepos(string) ([]api.Repository, string, string, error)
	LogOut(projectID string) error
	GetAuthCodeURL(projectID string) (string, error)
	// CreateWebHook
	// DeleteWebHook
	// PostCommitStatus
}

// Manager represents the manager for scm.
type Manager struct {
	DataStore *store.DataStore
}

// FindScm returns the scm by scm type.
func (scm *Manager) FindScm(scmType string) (SCM, error) {
	switch scmType {
	case api.GITHUB:
		return &provider.GitHubManager{DataStore: scm.DataStore}, nil
	case api.GITLAB:
		return &provider.GitLabManager{DataStore: scm.DataStore}, nil
	default:
		return nil, fmt.Errorf("Unknown scm type %s", scmType)
	}
}
