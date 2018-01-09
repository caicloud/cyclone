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
	"strings"

	log "github.com/golang/glog"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
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
				// The token of existed authorization can not be got, so delete it and recreate a new one.
				if _, err := client.Authorizations.Delete(*auth.ID); err != nil {
					log.Errorf("Fail to delete the token for %s as %s", oauthAppName, err.Error())
					return "", err
				}

				break
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Create the oauth token for Caicloud.
	authReq := &github.AuthorizationRequest{
		Scopes: []github.Scope{github.ScopeRepo},
		Note:   &oauthAppName,
	}
	auth, _, err := client.Authorizations.Create(authReq)
	if err != nil {
		return "", err
	}

	return *auth.Token, nil
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token
func (g *GitHub) CheckToken(scm *api.SCMConfig) bool {
	if _, err := g.ListRepos(scm); err != nil {
		return false
	}
	return true
}

// ListRepos lists the repos by the SCM config.
func (g *GitHub) ListRepos(scm *api.SCMConfig) ([]api.Repository, error) {
	client, err := newClientByBasicAuth(scm.Username, scm.Token)
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
	client, err := newClientByBasicAuth(scm.Username, scm.Token)
	if err != nil {
		return nil, err
	}

	opt := &github.ListOptions{
		PerPage: listPerPageOpt,
	}

	owner := scm.Username
	if strings.Contains(repo, "/") {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("repo %s is not correct, only supports one left slash", repo)
			log.Error(err.Error())
			return nil, err
		}
		owner, repo = parts[0], parts[1]
	}

	var allBranches []*github.Branch
	for {
		branches, resp, err := client.Repositories.ListBranches(owner, repo, opt)
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

// newClientByBasicAuth news GitHub client by basic auth, supports two types: username with password; username
// with OAuth token.
// Refer to https://developer.github.com/v3/auth/#basic-authentication
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
