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
	"net/http"
	"net/url"

	log "github.com/golang/glog"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	mgo "gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/scm"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/store"
)

const (
	// listPerPageOpt represents the value for PerPage in list options.
	listPerPageOpt = 30
)

// GitHub represents the SCM provider of GitHub.
type GitHub struct{}

func init() {
	if err := scm.RegisterProvider(api.GitHub, new(GitHub)); err != nil {
		log.Errorln(err)
	}
}

// GetToken gets the token by the username and password of SCM config.
func (g *GitHub) GetToken(scm *api.SCMConfig) (string, error) {
	if len(scm.Username) == 0 || len(scm.Password) == 0 {
		return "", fmt.Errorf("GitHub username or password is missing")
	}

	client, err := newClientByBasicAuth(scm.Username, scm.Password)
	if err != nil {
		return "", err
	}

	// oauthAppName represents the oauth app name for Cyclone.
	oauthAppName := "Caicloud"
	opt := &github.ListOptions{
		PerPage: listPerPageOpt,
	}
	for {
		auths, resp, err := client.Authorizations.List(opt)
		if err != nil {
			return "", err
		}

		for _, auth := range auths {
			if *auth.App.Name == oauthAppName {
				return *auth.Token, nil
				// TODO(robin) As the tokens of listed authorizations are empty, and hashed token can not directly use.
				// So delete the already existed authorization, and recreate a new one.
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Create the oauth token for Caicloud after not found.
	authReq := &github.AuthorizationRequest{
		Scopes: []github.Scope{github.ScopeRepo},
		Note: &oauthAppName,
	}
	auth, _, err := client.Authorizations.Create(authReq)
	if err != nil {
		return "", err
	}

	return *auth.Token, nil
}

// ListRepos lists the repos by the SCM config.
func (g *GitHub) ListRepos(scm *api.SCMConfig) ([]api.Repository, error) {
	client, err := newClientByToken(scm.Token)
	if err != nil {
		return nil, err
	}

	// List all repositories for the authenticated user.
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: listPerPageOpt,
		},
	}
	// Get all pages of results.
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.List("", opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	repos := make([]api.Repository, len(allRepos))
	for i, repo := range allRepos {
		repos[i].Name = *repo.FullName
		repos[i].URL = *repo.CloneURL
	}

	return repos, nil
}

// ListBranches lists the branches for specified repo.
func (g *GitHub) ListBranches(scm *api.SCMConfig, repo string) ([]string, error) {
	client, err := newClientByToken(scm.Token)
	if err != nil {
		return nil, err
	}

	opt := &github.ListOptions{
		PerPage: listPerPageOpt,
	}

	var allBranches []*github.Branch
	for {
		branches, resp, err := client.Repositories.ListBranches(scm.Username, repo, opt)
		if err != nil {
			return nil, err
		}
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	branches := make([]string, len(allBranches))
	for i, b := range allBranches {
		branches[i] = *b.Name
	}

	return branches, nil
}

// newClientByBasicAuth news GitHub client by username and password.
func newClientByBasicAuth(username, password string) (*github.Client, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				req.SetBasicAuth(username, password)
				return nil, nil
			},
		},
	}

	return github.NewClient(client), nil
}

// newClientByToken news GitHub client by token.
func newClientByToken(token string) (*github.Client, error) {
	// Use token to new GitHub client.
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(oauth2.NoContext, tokenSource)

	return github.NewClient(httpClient), nil
}

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
