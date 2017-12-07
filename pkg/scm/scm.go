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
	"strings"

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/store"
)

// scmProviders represents the set of SCM providers.
var scmProviders map[api.SCMType]SCMProvider

func init() {
	scmProviders = make(map[api.SCMType]SCMProvider)
}

// RegisterProvider registers SCM providers.
func RegisterProvider(scmType api.SCMType, provider SCMProvider) error {
	if _, ok := scmProviders[scmType]; ok {
		return fmt.Errorf("SCM provider %s already exists.", scmType)
	}

	scmProviders[scmType] = provider
	return nil
}

// SCMProvider represents the interface of SCM provider.
type SCMProvider interface {
	GetToken(scm *api.SCMConfig) (string, error)
	ListRepos(scm *api.SCMConfig) ([]api.Repository, error)
	ListBranches(scm *api.SCMConfig, repo string) ([]string, error)
	CheckToken(scm *api.SCMConfig) bool
}

// GetSCMProvider gets the SCM provider by the type.
func GetSCMProvider(scmType api.SCMType) (SCMProvider, error) {
	provider, ok := scmProviders[scmType]
	if !ok {
		return nil, fmt.Errorf("unsupported SCM type %s", scmType)
	}

	return provider, nil
}

// GenerateSCMToken generates the SCM token according to the config.
// Make sure the type, server of the SCM is provided. If the SCM is Github, the username is required.
// If the access token is provided, it should be checked whether has authority of repos.
// Generate new token only when the username and password are provided at the same time.
func GenerateSCMToken(config *api.SCMConfig) error {
	if config == nil {
		return httperror.ErrorContentNotFound.Format("SCM config")
	}

	// Trim suffix '/' of Gitlab server to ensure that the token can work, otherwise there will be 401 error.
	config.Server = strings.TrimSuffix(config.Server, "/")

	scmType := config.Type
	token := config.Token
	provider, err := GetSCMProvider(scmType)
	if err != nil {
		return err
	}

	switch scmType {
	case api.GitHub:
		// Github username is required.
		if len(config.Username) == 0 {
			err := fmt.Errorf("username of Github is required")
			return err
		}

		// If Github password is provided, generate the new token.
		if len(config.Password) != 0 {
			token, err = provider.GetToken(config)
			if err != nil {
				log.Errorf("fail to get SCM token for user %s as %s", config.Username, err.Error())
				return err
			}

			// Update the token if generate a new one.
			config.Token = token
		} else if !provider.CheckToken(config) {
			return fmt.Errorf("token is unauthorized to repos")
		}

	case api.GitLab:
		// If username and password is provided, generate the new token.
		if len(config.Username) != 0 && len(config.Password) != 0 {
			token, err = provider.GetToken(config)
			if err != nil {
				log.Errorf("fail to get SCM token for user %s as %s", config.Username, err.Error())
				return err
			}

			// Update the token if generate a new one.
			config.Token = token
		}
	case api.SVN:
		return fmt.Errorf("SCM %s is not supported", scmType)
	default:
		return fmt.Errorf("SCM type %s is unknow", scmType)
	}

	// Cleanup the password for security.
	config.Password = ""

	return nil
}

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
		// return &provider.GitHubManager{DataStore: scm.DataStore}, nil
		return nil, fmt.Errorf("not implemented")
	case api.GITLAB:
		// return &provider.GitLabManager{DataStore: scm.DataStore}, nil
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("Unknown scm type %s", scmType)
	}
}
