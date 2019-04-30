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

	"github.com/caicloud/nirvana/log"

	c_v1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
)

type newProviderFunc func(source *v1alpha1.SCMSource) (Provider, error)

// scmProviders represents the set of SCM providers.
var scmProviders map[v1alpha1.SCMType]newProviderFunc

func init() {
	scmProviders = make(map[v1alpha1.SCMType]newProviderFunc)
}

// RegisterProvider registers SCM providers.
func RegisterProvider(scmType v1alpha1.SCMType, pFunc newProviderFunc) error {
	if _, ok := scmProviders[scmType]; ok {
		return fmt.Errorf("scm provider %s already exists", scmType)
	}

	scmProviders[scmType] = pFunc
	return nil
}

// Provider represents the interface of SCM provider.
type Provider interface {
	GetToken() (string, error)
	ListRepos() ([]Repository, error)
	// ListBranches list branches of repo, repo format must be {owner}/{repo}.
	ListBranches(repo string) ([]string, error)
	// ListTags list tags of repo, repo format must be {owner}/{repo}.
	ListTags(repo string) ([]string, error)
	ListDockerfiles(repo string) ([]string, error)
	CreateStatus(status c_v1alpha1.StatusPhase, targetURL, repoURL, commitSHA string) error
	GetPullRequestSHA(repoURL string, number int) (string, error)
	CheckToken() error
	CreateWebhook(repo string, webhook *Webhook) error
	DeleteWebhook(repo string, webhookURL string) error
}

// GetSCMProvider gets the SCM provider by the type.
func GetSCMProvider(scm *v1alpha1.SCMSource) (Provider, error) {
	if scm == nil {
		err := fmt.Errorf("SCM config is nil")
		log.Error(err)
		return nil, err
	}

	scmType := scm.Type
	pFunc, ok := scmProviders[scmType]
	if !ok {
		return nil, fmt.Errorf("SCM type %s upsupported", scmType)
	}

	return pFunc(scm)
}

// GenerateSCMToken generates the SCM token according to the config.
// Make sure the type, server of the SCM is provided. If the SCM is Github, the username is required.
// If the access token is provided, it should be checked whether has authority of repos.
// Generate new token only when the username and password are provided at the same time.
func GenerateSCMToken(config *v1alpha1.SCMSource) error {
	if config == nil {
		return fmt.Errorf("SCM config %s not found", config)
	}

	if config.AuthType != v1alpha1.AuthTypePassword && config.AuthType != v1alpha1.AuthTypeToken {
		return fmt.Errorf("SCM auth type %s upsupported", config.AuthType)
	}

	// Trim suffix '/' of Gitlab server to ensure that the token can work, otherwise there will be 401 error.
	config.Server = strings.TrimSuffix(config.Server, "/")

	scmType := config.Type
	provider, err := GetSCMProvider(config)
	if err != nil {
		return err
	}

	var generatedToken string

	switch scmType {
	case v1alpha1.GitHub, v1alpha1.GitLab:
		// If username and password is provided, generate the new token.
		if len(config.User) != 0 && len(config.Password) != 0 {
			generatedToken, err = provider.GetToken()
			if err != nil {
				log.Errorf("fail to get SCM token for user %s as %s", config.User, err.Error())
				return err
			}
		}
	default:
		return fmt.Errorf("SCM type %s unsupported", scmType)
	}

	if generatedToken != "" && config.AuthType == v1alpha1.AuthTypePassword {
		config.Token = generatedToken
	} else {
		err := provider.CheckToken()
		if err != nil {
			return fmt.Errorf("token unauthorized to repos, error:%v", err)
		}
	}

	// Cleanup the password for security.
	config.Password = ""

	return nil
}

// Webhook represents the params for SCM webhook.
type Webhook struct {
	Events []EventType
	URL    string
}

// EventType represents event types of SCM.
type EventType string

const (
	// PullRequestEventType represents pull request events.
	PullRequestEventType EventType = "scm-pull-request"
	// PullRequestCommentEventType represents pull request comment events.
	PullRequestCommentEventType EventType = "scm-pull-request-comment"
	// PushEventType represents commit push events.
	PushEventType EventType = "scm-push"
	// TagReleaseEventType represents tag release events.
	TagReleaseEventType EventType = "scm-tag-release"
)

// EventData represents the data parsed from SCM events.
type EventData struct {
	Type      EventType
	Repo      string
	Ref       string
	Branch    string
	Comment   string
	CommitSHA string
}
