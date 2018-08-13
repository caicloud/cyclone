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

type newSCMProviderFunc func(*api.SCMConfig) (SCMProvider, error)

// scmProviders represents the set of SCM providers.
var scmProviders map[api.SCMType]newSCMProviderFunc

func init() {
	scmProviders = make(map[api.SCMType]newSCMProviderFunc)
}

// RegisterProvider registers SCM providers.
func RegisterProvider(scmType api.SCMType, pFunc newSCMProviderFunc) error {
	if _, ok := scmProviders[scmType]; ok {
		return fmt.Errorf("SCM provider %s already exists.", scmType)
	}

	scmProviders[scmType] = pFunc
	return nil
}

// SCMProvider represents the interface of SCM provider.
type SCMProvider interface {
	GetToken() (string, error)
	ListRepos() ([]api.Repository, error)
	ListBranches(repo string) ([]string, error)
	ListTags(repo string) ([]string, error)
	CheckToken() bool
	NewTagFromLatest(tagName, description, commitID, url string) error
	CreateWebHook(repoURL string, webHook *WebHook) error
	DeleteWebHook(repoURL string, webHookUrl string) error
	//Gitlab Oauth
	GetAuthCodeURL(state string, scmType api.SCMType) (string, error)
	Authcallback(code, state string) (string, error)
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
func GetSCMProvider(scm *api.SCMConfig) (SCMProvider, error) {
	if scm == nil {
		err := fmt.Errorf("SCM config is nil")
		log.Error(err)
		return nil, err
	}

	scmType := scm.Type
	pFunc, ok := scmProviders[scmType]
	if !ok {
		return nil, fmt.Errorf("unsupported SCM type %s", scmType)
	}

	return pFunc(scm)
}

// GenerateSCMToken generates the SCM token according to the config.
// Make sure the type, server of the SCM is provided. If the SCM is Github, the username is required.
// If the access token is provided, it should be checked whether has authority of repos.
// Generate new token only when the username and password are provided at the same time.
func GenerateSCMToken(config *api.SCMConfig) error {
	if config == nil {
		return httperror.ErrorContentNotFound.Format("SCM config")
	}
	
	// when you choose the way that is gitlab oauth, the frontend go back the config which contains the AuthType.
	// the AuthType that is "OAuth" means that you choose gitlab oauth.
	if config.AuthType != api.Password && config.AuthType != api.Token && config.AuthType != "OAuth" {
		return httperror.ErrorValidationFailed.Format("SCM authType %s is unknow", config.AuthType)
	}

	// Trim suffix '/' of Gitlab server to ensure that the token can work, otherwise there will be 401 error.
	config.Server = strings.TrimSuffix(config.Server, "/")

	scmType := config.Type
	// when you choose the way that gitlab oauth ,the frontend go back the config which contains the token and the AuthType.
	// the way that you have token, so you dont execute the get token with provider.
	if config.AuthType == "OAuth" && len(config.Token) != 0 {
		return nil
	}

	provider, err := GetSCMProvider(config)
	if err != nil {
		return err
	}

	var generatedToken string

	switch scmType {
	case api.Github:
		// Github username is required.
		if len(config.Username) == 0 {
			return httperror.ErrorContentNotFound.Format("Github username")
		}

		// If Github password is provided, generate the new token.
		if len(config.Password) != 0 {
			generatedToken, err = provider.GetToken()
			if err != nil {
				log.Errorf("fail to get SCM token for user %s as %s", config.Username, err.Error())
				return err
			}
		}
	case api.Gitlab:
		// If username and password is provided, generate the new token.
		if len(config.Username) != 0 && len(config.Password) != 0 {
			generatedToken, err = provider.GetToken()
			if err != nil {
				log.Errorf("fail to get SCM token for user %s as %s", config.Username, err.Error())
				return err
			}
		}
	case api.SVN:
		generatedToken, _ = provider.GetToken()
	default:
		return httperror.ErrorValidationFailed.Format("SCM type %s is unknow", scmType)
	}

	if generatedToken != "" {
		config.Token = generatedToken
	} else if !provider.CheckToken() {
		return httperror.ErrorValidationFailed.Format("token is unauthorized to repos")
	}

	// Cleanup the password for security.
	config.Password = ""

	return nil
}

func NewTagFromLatest(codeSource *api.CodeSource, scm *api.SCMConfig, tagName, description string) error {
	commitID, err := wscm.GetCommitID(codeSource, "")
	if err != nil {
		return err
	}

	gitSource, err := api.GetGitSource(codeSource)
	if err != nil {
		return err
	}

	url := gitSource.Url

	p, err := GetSCMProvider(scm)
	if err != nil {
		return err
	}

	err = p.NewTagFromLatest(tagName, description, commitID, url)
	if err != nil {
		return err
	}

	return nil
}
