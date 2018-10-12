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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
	"github.com/caicloud/cyclone/pkg/scm/provider"
)

func init() {
	if err := scm.RegisterProvider(api.Github, NewGithub); err != nil {
		log.Errorln(err)
	}
}

// Github represents the SCM provider of Github.
type Github struct {
	scmCfg *api.SCMConfig
	client *github.Client
}

// NewGithub new Github client.
func NewGithub(scmCfg *api.SCMConfig) (scm.SCMProvider, error) {
	var client *github.Client
	var err error
	if scmCfg.Token == "" {
		client, err = newClientByBasicAuth(scmCfg.Username, scmCfg.Password)
		if err != nil {
			return nil, err
		}
	} else {
		client, err = newClientByBasicAuth(scmCfg.Username, scmCfg.Token)
		if err != nil {
			return nil, err
		}
	}

	return &Github{scmCfg, client}, nil
}

// GetToken gets the token by the username and password of SCM config.
func (g *Github) GetToken() (string, error) {
	scmCfg := g.scmCfg
	if len(scmCfg.Username) == 0 || len(scmCfg.Password) == 0 {
		return "", fmt.Errorf("Github username or password is missing")
	}

	// oauthAppName represents the oauth app name for Cyclone.
	oauthAppName := "Caicloud"
	opt := &github.ListOptions{
		PerPage: provider.ListPerPageOpt,
	}
	for {
		auths, resp, err := g.client.Authorizations.List(opt)
		if err != nil {
			return "", err
		}

		for _, auth := range auths {
			if *auth.App.Name == oauthAppName {
				// The token of existed authorization can not be got, so delete it and recreate a new one.
				if _, err := g.client.Authorizations.Delete(*auth.ID); err != nil {
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
	auth, _, err := g.client.Authorizations.Create(authReq)
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
func (g *Github) ListRepos() ([]api.Repository, error) {
	return g.listReposInner(true)
}

// listReposInner lists the repos by the SCM config,
// list all repos while the parameter 'listAll' is true,
// otherwise, list repos by default 'listPerPageOpt' number.
func (g *Github) listReposInner(listAll bool) ([]api.Repository, error) {
	// List all repositories for the authenticated user.
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: provider.ListPerPageOpt,
		},
	}
	// Get all pages of results.
	var allRepos []*github.Repository
	for {
		repos, resp, err := g.client.Repositories.List("", opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 || !listAll {
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
func (g *Github) ListBranches(repo string) ([]string, error) {
	opt := &github.ListOptions{
		PerPage: provider.ListPerPageOpt,
	}

	owner := g.scmCfg.Username
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
		branches, resp, err := g.client.Repositories.ListBranches(owner, repo, opt)
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
		PerPage: provider.ListPerPageOpt,
	}

	owner := g.scmCfg.Username
	if strings.Contains(repo, "/") {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("repo %s is not correct, only supports one left slash", repo)
			log.Error(err.Error())
			return nil, err
		}
		owner, repo = parts[0], parts[1]
	}

	var allTags []*github.RepositoryTag
	for {
		tags, resp, err := g.client.Repositories.ListTags(owner, repo, opt)
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

// CreateWebHook creates webhook for specified repo.
func (g *Github) CreateWebHook(repoURL string, webHook *scm.WebHook) error {
	if webHook == nil || len(webHook.Url) == 0 || len(webHook.Events) == 0 {
		return fmt.Errorf("The webhook %v is not correct", webHook)
	}

	// Hook name must be passed as "web".
	// Ref: https://developer.github.com/v3/repos/hooks/#create-a-hook
	hookName := "web"
	hook := github.Hook{
		Name:   &hookName,
		Events: convertToGithubEvents(webHook.Events),
		Config: map[string]interface{}{
			"url":          webHook.Url,
			"content_type": "json",
		},
	}
	owner, name := provider.ParseRepoURL(repoURL)
	_, _, err := g.client.Repositories.CreateHook(owner, name, &hook)
	return err
}

// DeleteWebHook deletes webhook from specified repo.
func (g *Github) DeleteWebHook(repoURL string, webHookUrl string) error {
	owner, name := provider.ParseRepoURL(repoURL)
	hooks, _, err := g.client.Repositories.ListHooks(owner, name, nil)
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		if hookurl, ok := hook.Config["url"].(string); ok {
			if strings.HasPrefix(hookurl, webHookUrl) {
				_, err = g.client.Repositories.DeleteHook(owner, name, *hook.ID)
				return nil
			}
		}
	}

	return nil
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

// NewTagFromLatest generate a new tag.
func (g *Github) NewTagFromLatest(tagName, description, commitID, url string) error {
	client := newClientByToken(g.scmCfg.Token)

	objecttype := "commit"
	curtime := time.Now()
	email := "circle@caicloud.io"
	name := "circle"

	tag := &github.Tag{
		Tag:     &tagName,
		Message: &(description),
		Object: &github.GitObject{
			Type: &objecttype,
			SHA:  &commitID,
		},
		Tagger: &github.CommitAuthor{
			Date:  &curtime,
			Name:  &name,
			Email: &email,
		},
	}

	owner, repo := provider.ParseRepoURL(url)
	_, _, err := client.Git.CreateTag(owner, repo, tag)
	if err != nil {
		return err
	}

	ref := "refs/tags/" + tagName
	reference := &github.Reference{
		Ref: &ref,
		Object: &github.GitObject{
			Type: &objecttype,
			SHA:  &commitID,
		},
	}
	refs, _, err := client.Git.CreateRef(owner, repo, reference)
	log.Info(refs)
	return err
}

// convertToGithubEvents converts the defined event types to Github event types.
func convertToGithubEvents(events []scm.EventType) []string {
	var ge []string
	for _, e := range events {
		switch e {
		case scm.PullRequestEventType:
			ge = append(ge, "pull_request")
		case scm.PullRequestCommentEventType:
			ge = append(ge, "issue_comment")
		case scm.PushEventType:
			ge = append(ge, "push")
		case scm.TagReleaseEventType:
			ge = append(ge, "release")
		default:
			log.Errorf("The event type %s is not supported, will be ignored", e)
		}
	}

	return ge
}

func (g *Github) GetTemplateType(repo string) (string, error) {
	owner := g.scmCfg.Username
	if strings.Contains(repo, "/") {
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			err := fmt.Errorf("repo %s is not correct, only supports one left slash", repo)
			log.Error(err.Error())
			return "", err
		}
		owner, repo = parts[0], parts[1]
	}

	languages, r, err := g.client.Repositories.ListLanguages(owner, repo)
	log.Error(r, err)
	if err != nil {
		log.Error("list language failed:%v", err)
		return "", err
	}

	language := getTopLanguage(languages)

	switch language {
	case api.JavaRepoType, api.JavaScriptRepoType:
		opt := &github.RepositoryContentGetOptions{}
		_, directories, _, err := g.client.Repositories.GetContents(owner, repo, "", opt)
		if err != nil {
			log.Error("get contents failed:%v", err)
			return language, nil
		}

		for _, d := range directories {
			if language == api.JavaRepoType && strings.Contains(*d.Name, "pom.xml") {
				return api.MavenRepoType, nil
			}
			if language == api.JavaRepoType && strings.Contains(*d.Name, "build.gradle") {
				return api.GradleRepoType, nil
			}
			if language == api.JavaScriptRepoType && strings.Contains(*d.Name, "package.json") {
				return api.NodeRepoType, nil
			}
		}
	}

	return language, nil
}

// getTopLanguage get the top usage language name.
func getTopLanguage(languages map[string]int) string {
	var language string
	var max int
	for l, value := range languages {
		if value > max {
			max = value
			language = l
		}
	}
	return language
}

// CreateStatus generate a new status for repository.
func (g *Github) CreateStatus(recordStatus api.Status, targetURL, repoURL, commitSHA string) error {
	// GitHub : error, failure, pending, or success.
	state := "pending"
	description := ""

	switch recordStatus {
	case api.Running:
		state = "pending"
		description = "The Cyclone CI build is in progress."
	case api.Success:
		state = "success"
		description = "The Cyclone CI build passed."
	case api.Failed:
		state = "failure"
		description = "The Cyclone CI build failed."
	case api.Aborted:
		state = "failure"
		description = "The Cyclone CI build failed."
	default:
		log.Errorf("not supported state:%s", recordStatus)
	}

	owner, repo := provider.ParseRepoURL(repoURL)
	client := newClientByToken(g.scmCfg.Token)
	email := "cyclone@caicloud.io"
	name := "cyclone"
	context := "continuous-integration/cyclone"
	creator := github.User{
		Name:  &name,
		Email: &email,
	}
	status := &github.RepoStatus{
		State:       &state,
		Description: &description,
		TargetURL:   &targetURL,
		Context:     &context,
		Creator:     &creator,
	}
	//var owner, repo, ref string
	_, _, err := client.Repositories.CreateStatus(owner, repo, commitSHA, status)
	log.Error(err)
	return err
}

// GetPullRequestSHA get the commit sha for specified pr.
func (g *Github) GetPullRequestSHA(repoURL string, number int) (string, error) {
	owner, repo := provider.ParseRepoURL(repoURL)
	pr, _, err := g.client.PullRequests.Get(owner, repo, number)
	if err != nil {
		log.Error(err)
	}

	return *pr.Head.SHA, nil
}
