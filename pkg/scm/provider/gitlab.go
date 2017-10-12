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

package provider

import (
	"fmt"

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/osutil"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/store"
	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	mgo "gopkg.in/mgo.v2"
)

// GitLabManager represents the manager for gitlab.
type GitLabManager struct {
	DataStore *store.DataStore
}

// Authcallback is the callback handler.
func (g *GitLabManager) Authcallback(code, state string) (string, error) {
	if code == "" || state == "" {
		return "", fmt.Errorf("code: %s or state: %s is nil", code, state)
	}

	//caicloud web address,eg caicloud.io
	uiPath := osutil.GetStringEnv(cloud.ConsoleWebEndpoint, "http://localhost:8000")
	redirectURL := fmt.Sprintf("%s/cyclone/add?type=gitlab&code=%s&state=%s", uiPath, code, state)

	if err := g.setToken(code, state); err != nil {
		return "", err
	}
	return redirectURL, nil
}

// setToken sets the token.
func (g *GitLabManager) setToken(code, state string) error {
	// Get the oauth config to request token.
	config, err := getConfig(api.GITLAB)
	if err != nil {
		return err
	}

	// To communicate with gitlab or other scm to get token.
	var token *oauth2.Token
	token, err = config.Exchange(oauth2.NoContext, code) // Post a token request and receive toeken.
	if err != nil {
		return err
	}

	if !token.Valid() {
		return fmt.Errorf("Token invalid. Got: %#v", token)
	}

	// Create token in database (but not ready to use yet).
	scmToken := api.ScmToken{
		ProjectID: state,
		ScmType:   api.GITLAB,
		Token:     *token,
	}

	if _, err = g.DataStore.Findtoken(state, api.GitHub); err != nil {
		if err == mgo.ErrNotFound {
			if _, err = g.DataStore.CreateToken(&scmToken); err != nil {
				return err
			}
		}
		return err
	}

	if err = g.DataStore.UpdateToken2(&scmToken); err != nil {
		return err
	}

	return nil
}

// GetRepos gets the list of repositories with token from gitlab.
func (g *GitLabManager) GetRepos(projectID string) (repos []api.Repository, username string, avatarURL string, err error) {
	token, err := g.DataStore.Findtoken(projectID, api.GITLAB)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, username, avatarURL, httperror.ErrorContentNotFound.Format("token", projectID, api.GitHub)
		}
		return nil, username, avatarURL, err
	}

	gitlabServer := osutil.GetStringEnv(cloud.GitlabURL, "https://gitlab.com")
	client := gitlab.NewOAuthClient(nil, token.Token.AccessToken)
	client.SetBaseURL(gitlabServer + "/api/v3/")

	opt := &gitlab.ListProjectsOptions{}
	projects, _, err := client.Projects.ListProjects(opt)
	if err != nil {
		return nil, username, avatarURL, err
	}

	repos = make([]api.Repository, len(projects))
	for i, repo := range projects {
		repos[i].Name = repo.Name
		repos[i].URL = repo.HTTPURLToRepo
		repos[i].Owner = (*repo.Namespace).Name
		if repo.Owner != nil {
			username = repos[i].Owner
		}
	}

	optUsers := &gitlab.ListUsersOptions{Username: &username}
	users, _, err := client.Users.ListUsers(optUsers)
	if err != nil {
		return repos, username, avatarURL, err
	}
	avatarURL = users[0].AvatarURL

	return repos, username, avatarURL, nil
}

// LogOut logs out and deletes the token.
func (g *GitLabManager) LogOut(projectID string) error {
	return g.DataStore.DeleteToken(projectID, api.GITLAB)
}

// GetAuthCodeURL gets the URL for token request.
func (g *GitLabManager) GetAuthCodeURL(projectID string) (string, error) {
	return getAuthCodeURL(projectID)
}
