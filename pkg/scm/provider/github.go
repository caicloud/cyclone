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
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	mgo "gopkg.in/mgo.v2"
)

// GitHubManager represents the manager for github.
type GitHubManager struct {
	DataStore *store.DataStore
}

// Authcallback is the callback handler.
func (g *GitHubManager) Authcallback(code, state string) (string, error) {
	if code == "" || state == "" {
		return "", fmt.Errorf("code: %s or state: %s is nil", code, state)
	}

	// Caicloud web address,eg caicloud.io
	uiPath := osutil.GetStringEnv(cloud.ConsoleWebEndpoint, "http://localhost:8000")
	redirectURL := fmt.Sprintf("%s/cyclone/add?type=github&code=%s&state=%s", uiPath, code, state)

	if err := g.setToken(code, state); err != nil {
		return "", err
	}

	return redirectURL, nil
}

// setToken sets the token.
func (g *GitHubManager) setToken(code, state string) error {
	// Get the oauth config to request token.
	config, err := getConfig(api.GITHUB)
	if err != nil {
		return err
	}

	// To communicate with github or other scm to get token.
	var token *oauth2.Token
	token, err = config.Exchange(oauth2.NoContext, code) // Post a token request and receive token.
	if err != nil {
		return err
	}

	if !token.Valid() {
		return fmt.Errorf("Token invalid. Got: %#v", token)
	}

	// Create token in database (but not ready to use yet).
	scmToken := api.ScmToken{
		ProjectID: state,
		ScmType:   api.GitHub,
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

// GetRepos gets the list of repositories with token from github.
func (g *GitHubManager) GetRepos(projectID string) (Repos []api.Repository, username, avatarURL string, err error) {
	token, err := g.DataStore.Findtoken(projectID, api.GitHub)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, username, avatarURL, httperror.ErrorContentNotFound.Format("token", projectID, api.GitHub)
		}
		return nil, username, avatarURL, err
	}

	// Use token to get repository list.
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token.Token.AccessToken},
	)
	httpClient := oauth2.NewClient(oauth2.NoContext, tokenSource)

	client := github.NewClient(httpClient)

	// List all repositories for the authenticated user.
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}
	// Get all pages of results.
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.List("", opt)
		if err != nil {
			return nil, username, avatarURL, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	Repos = make([]api.Repository, len(allRepos))
	for i, repo := range allRepos {
		Repos[i].Name = *repo.Name
		Repos[i].URL = *repo.CloneURL
		Repos[i].Owner = *repo.Owner.Login
	}

	user, _, err := client.Users.Get("")
	if err != nil {
		return Repos, username, avatarURL, err
	}
	username = *user.Login
	avatarURL = *user.AvatarURL

	return Repos, username, avatarURL, nil
}

// LogOut logs out and deletes the token.
func (g *GitHubManager) LogOut(projectID string) error {
	token, err := g.DataStore.Findtoken(projectID, api.GitHub)
	if err != nil {
		return err
	}

	config, err := getConfig(api.GITHUB)
	if err != nil {
		return err
	}

	tp := github.BasicAuthTransport{
		Username: config.ClientID,
		Password: config.ClientSecret,
	}

	client := github.NewClient(tp.Client())
	if _, err = client.Authorizations.Revoke(config.ClientID, token.Token.AccessToken); err != nil {
		return err
	}

	if err = g.DataStore.DeleteToken(projectID, api.GitHub); err != nil {
		return err
	}

	return nil
}

// GetAuthCodeURL gets the URL for token request.
func (g *GitHubManager) GetAuthCodeURL(projectID string) (string, error) {
	return getAuthCodeURL(projectID)
}
