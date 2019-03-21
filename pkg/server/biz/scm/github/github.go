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

package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/caicloud/nirvana/log"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
)

func init() {
	if err := scm.RegisterProvider(v1alpha1.GitHub, NewGithub); err != nil {
		log.Errorln(err)
	}
}

// Github represents the SCM provider of Github.
type Github struct {
	scmCfg *v1alpha1.SCMSource
	client *github.Client
	ctx    context.Context
}

// NewGithub new Github client.
func NewGithub(scmCfg *v1alpha1.SCMSource) (scm.Provider, error) {
	var client *github.Client
	var err error
	if scmCfg.Token == "" {
		client, err = newClientByBasicAuth(scmCfg.User, scmCfg.Password)
		if err != nil {
			return nil, err
		}
	} else {
		client, err = newClientByBasicAuth(scmCfg.User, scmCfg.Token)
		if err != nil {
			return nil, err
		}
	}

	return &Github{scmCfg, client, context.Background()}, nil
}

// GetToken gets the token by the username and password of SCM config.
func (g *Github) GetToken() (string, error) {
	scmCfg := g.scmCfg
	if len(scmCfg.User) == 0 || len(scmCfg.Password) == 0 {
		return "", fmt.Errorf("Github username or password is missing")
	}

	// oauthAppName represents the oauth app name for Cyclone.
	oauthAppName := "Cyclone"
	opt := &github.ListOptions{
		PerPage: scm.ListPerPageOpt,
	}
	for {
		auths, resp, err := g.client.Authorizations.List(g.ctx, opt)
		if err != nil {
			return "", err
		}

		for _, auth := range auths {
			if *auth.App.Name == oauthAppName {
				// The token of existed authorization can not be got, so delete it and recreate a new one.
				if _, err := g.client.Authorizations.Delete(g.ctx, *auth.ID); err != nil {
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

	// Create the oauth token for Cyclone
	authReq := &github.AuthorizationRequest{
		Scopes: []github.Scope{github.ScopeRepo},
		Note:   &oauthAppName,
	}
	auth, _, err := g.client.Authorizations.Create(g.ctx, authReq)
	if err != nil {
		return "", err
	}

	return *auth.Token, nil
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token
func (g *Github) CheckToken() bool {
	if _, err := g.listReposInner(false); err != nil {
		return false
	}
	return true
}

// ListRepos lists the repos by the SCM config.
func (g *Github) ListRepos() ([]scm.Repository, error) {
	return g.listReposInner(true)
}

// listReposInner lists the repos by the SCM config,
// list all repos while the parameter 'listAll' is true,
// otherwise, list repos by default 'listPerPageOpt' number.
func (g *Github) listReposInner(listAll bool) ([]scm.Repository, error) {
	// List all repositories for the authenticated user.
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: scm.ListPerPageOpt,
		},
	}

	// Get all pages of results.
	var allRepos []*github.Repository
	for {
		repos, resp, err := g.client.Repositories.List(g.ctx, "", opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 || !listAll {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	repos := make([]scm.Repository, len(allRepos))
	for i, repo := range allRepos {
		repos[i].Name = *repo.FullName
		repos[i].URL = *repo.CloneURL
	}

	return repos, nil
}

// ListBranches lists the branches for specified repo.
func (g *Github) ListBranches(repo string) ([]string, error) {
	opt := &github.ListOptions{
		PerPage: scm.ListPerPageOpt,
	}

	owner := g.scmCfg.User
	if strings.Contains(repo, "/") {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repo)
			log.Error(err.Error())
			return nil, err
		}
		owner, repo = parts[0], parts[1]
	}

	var allBranches []*github.Branch
	for {
		branches, resp, err := g.client.Repositories.ListBranches(g.ctx, owner, repo, opt)
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

// ListTags lists the tags for specified repo.
func (g *Github) ListTags(repo string) ([]string, error) {
	opt := &github.ListOptions{
		PerPage: scm.ListPerPageOpt,
	}

	owner := g.scmCfg.User
	if strings.Contains(repo, "/") {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repo)
			log.Error(err.Error())
			return nil, err
		}
		owner, repo = parts[0], parts[1]
	}

	var allTags []*github.RepositoryTag
	for {
		tags, resp, err := g.client.Repositories.ListTags(g.ctx, owner, repo, opt)
		if err != nil {
			return nil, err
		}
		allTags = append(allTags, tags...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	tags := make([]string, len(allTags))
	for i, b := range allTags {
		tags[i] = *b.Name
	}

	return tags, nil
}

// ListDockerfiles lists the dockerfiles for specified repo.
func (g *Github) ListDockerfiles(repo string) ([]string, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	owner := g.scmCfg.User
	if strings.Contains(repo, "/") {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("invalid repo %s, must in format of '{owner}/{repo}'", repo)
			log.Error(err.Error())
			return nil, err
		}
		owner, repo = parts[0], parts[1]
	}

	q := fmt.Sprintf("FROM filename:Dockerfile repo:%s/%s", owner, repo)
	var allCodeResult []github.CodeResult
	for {
		csr, resp, err := g.client.Search.Code(g.ctx, q, opt)
		if err != nil {
			return nil, err
		}

		allCodeResult = append(allCodeResult, csr.CodeResults...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	crs := []string{}
	for _, c := range allCodeResult {
		if *c.Name == "Dockerfile" {
			crs = append(crs, *c.Path)
		}
	}

	return crs, nil
}

// newClientByBasicAuth news Github client by basic auth, supports two types: username with password; username
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

// newClientByToken news Github client by token.
func newClientByToken(token string) *github.Client {
	// Use token to new Github client.
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(oauth2.NoContext, tokenSource)

	return github.NewClient(httpClient)
}
