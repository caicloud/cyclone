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
	wscm "github.com/caicloud/cyclone/pkg/worker/scm"
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
// TODO(robin) Refactor this interface to avoid the SCM config param for each method.
type SCMProvider interface {
	GetToken(scm *api.SCMConfig) (string, error)
	ListRepos(scm *api.SCMConfig) ([]api.Repository, error)
	ListBranches(scm *api.SCMConfig, repo string) ([]string, error)
	ListTags(scm *api.SCMConfig, repo string) ([]string, error)
	CheckToken(scm *api.SCMConfig) bool
	NewTagFromLatest(tagName, description, commitID, url, token string) error
	CreateWebHook(scm *api.SCMConfig, repoURL string, webHook *WebHook) error
	DeleteWebHook(scm *api.SCMConfig, repoURL string, webHookUrl string) error
}

// WebHook represents the params for SCM webhook.
type WebHook struct {
	Events []EventType
	Url    string
}

type EventType string

const (
	PullRequestEventType        EventType = "PullRequest"
	PullRequestCommentEventType EventType = "PullRequestComment"
	PushEventType               EventType = "Push"
	TagReleaseEventType         EventType = "TagRelease"
)

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
	provider, err := GetSCMProvider(scmType)
	if err != nil {
		return err
	}

	var generatedToken string

	switch scmType {
	case api.GitHub:
		// Github username is required.
		if len(config.Username) == 0 {
			err := fmt.Errorf("username of Github is required")
			return err
		}

		// If Github password is provided, generate the new token.
		if len(config.Password) != 0 {
			generatedToken, err = provider.GetToken(config)
			if err != nil {
				log.Errorf("fail to get SCM token for user %s as %s", config.Username, err.Error())
				return err
			}
		}
	case api.GitLab:
		// If username and password is provided, generate the new token.
		if len(config.Username) != 0 && len(config.Password) != 0 {
			generatedToken, err = provider.GetToken(config)
			if err != nil {
				log.Errorf("fail to get SCM token for user %s as %s", config.Username, err.Error())
				return err
			}
		}
	case api.SVN:
		generatedToken, _ = provider.GetToken(config)
	default:
		return fmt.Errorf("SCM type %s is unknow", scmType)
	}

	if generatedToken != "" {
		config.Token = generatedToken
	} else if !provider.CheckToken(config) {
		return fmt.Errorf("token is unauthorized to repos")
	}

	// Cleanup the password for security.
	config.Password = ""

	return nil
}

func NewTagFromLatest(codeSource *api.CodeSource, tagName, description, token string) error {
	commitID, err := wscm.GetCommitID(codeSource, "")
	if err != nil {
		return err
	}
	scmType := codeSource.Type

	gitSource, err := api.GetGitSource(codeSource)
	if err != nil {
		return err
	}

	url := gitSource.Url

	p, err := GetSCMProvider(scmType)
	if err != nil {
		return err
	}

	err = p.NewTagFromLatest(tagName, description, commitID, url, token)
	if err != nil {
		return err
	}

	return nil
}
