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
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/caicloud/nirvana/log"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	c_v1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

const (
	// EventTypeHeader represents the header key for event type of Github.
	EventTypeHeader = "X-Github-Event"

	// branchRefTemplate represents reference template for branches.
	// branchRefTemplate = "refs/heads/%s"

	// tagRefTemplate represents reference template for tags.
	tagRefTemplate = "refs/tags/%s"

	// pullRefTemplate represents reference template for pull request.
	pullRefTemplate = "refs/pull/%d/merge"

	// errorFieldResponse represents the field `Response` of Github error.
	errorFieldResponse = "Response"
)

var (
	statusURLRegexpStr = `^https://api.github.com/repos/[\S]+/[\S]+/statuses/([\w]+)$`
	statusURLRegexp    = regexp.MustCompile(statusURLRegexpStr)
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
	if scmCfg.Token == "" {
		client = newClientByBasicAuth(scmCfg.User, scmCfg.Password)
	} else {
		client = newClientByBasicAuth(scmCfg.User, scmCfg.Token)
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
		PerPage: scm.ListOptPerPage,
	}
	for {
		auths, resp, err := g.client.Authorizations.List(g.ctx, opt)
		if err != nil {
			return "", convertGithubError(err)
		}

		for _, auth := range auths {
			if *auth.App.Name == oauthAppName {
				// The token of existed authorization can not be got, so delete it and recreate a new one.
				if _, err = g.client.Authorizations.Delete(g.ctx, *auth.ID); err != nil {
					log.Errorf("Fail to delete the token for %s as %s", oauthAppName, err.Error())
					return "", convertGithubError(err)
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
		return "", convertGithubError(err)
	}

	return *auth.Token, nil
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token
func (g *Github) CheckToken() error {
	if _, err := g.listReposInner(false); err != nil {
		return err
	}
	return nil
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
			PerPage: scm.ListOptPerPage,
		},
	}

	// Get all pages of results.
	var allRepos []*github.Repository
	for {
		repos, resp, err := g.client.Repositories.List(g.ctx, "", opt)
		if err != nil {
			return nil, convertGithubError(err)
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
		PerPage: scm.ListOptPerPage,
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

	var allBranches []string
	for {
		branches, resp, err := g.client.Repositories.ListBranches(g.ctx, owner, repo, opt)
		if err != nil {
			return nil, convertGithubError(err)
		}

		for _, b := range branches {
			allBranches = append(allBranches, *b.Name)
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allBranches, nil
}

// ListTags lists the tags for specified repo.
func (g *Github) ListTags(repo string) ([]string, error) {
	opt := &github.ListOptions{
		PerPage: scm.ListOptPerPage,
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
			return nil, convertGithubError(err)
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
			PerPage: scm.ListOptPerPage,
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
			return nil, convertGithubError(err)
		}

		allCodeResult = append(allCodeResult, csr.CodeResults...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	crs := []string{}
	for _, c := range allCodeResult {
		if scm.IsDockerfile(*c.Path) {
			crs = append(crs, *c.Path)
		}
	}

	return crs, nil
}

// CreateStatus generate a new status for repository.
func (g *Github) CreateStatus(status c_v1alpha1.StatusPhase, targetURL, repoURL, commitSHA string) error {
	// GitHub : error, failure, pending, or success.
	state := "pending"
	description := ""

	switch status {
	case c_v1alpha1.StatusRunning:
		state = "pending"
		description = "Cyclone CI is in progress."
	case c_v1alpha1.StatusSucceeded:
		state = "success"
		description = "Cyclone CI passed."
	case c_v1alpha1.StatusFailed:
		state = "failure"
		description = "Cyclone CI failed."
	case c_v1alpha1.StatusCancelled:
		state = "failure"
		description = "Cyclone CI failed."
	default:
		err := fmt.Errorf("not supported state:%s", status)
		log.Error(err)
		return err
	}

	owner, repo := scm.ParseRepo(repoURL)
	client := newClientByToken(g.scmCfg.Token)
	email := "cyclone@caicloud.dev"
	name := "cyclone"
	context := "continuous-integration/cyclone"
	creator := github.User{
		Name:  &name,
		Email: &email,
	}
	repoStatus := &github.RepoStatus{
		State:       &state,
		Description: &description,
		TargetURL:   &targetURL,
		Context:     &context,
		Creator:     &creator,
	}
	_, _, err := client.Repositories.CreateStatus(g.ctx, owner, repo, commitSHA, repoStatus)
	return err
}

// GetPullRequestSHA gets latest commit SHA of pull request.
func (g *Github) GetPullRequestSHA(repoURL string, number int) (string, error) {
	owner, repo := scm.ParseRepo(repoURL)
	pr, _, err := g.client.PullRequests.Get(g.ctx, owner, repo, number)
	if err != nil {
		log.Error(err)
	}

	return *pr.Head.SHA, nil
}

// newClientByBasicAuth news Github client by basic auth, supports two types: username with password; username
// with OAuth token.
// Refer to https://developer.github.com/v3/auth/#basic-authentication
func newClientByBasicAuth(username, password string) *github.Client {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				req.SetBasicAuth(username, password)
				return nil, nil
			},
		},
	}

	return github.NewClient(client)
}

// newClientByToken news Github client by token.
func newClientByToken(token string) *github.Client {
	// Use token to new Github client.
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.TODO(), tokenSource)

	return github.NewClient(httpClient)
}

// GetWebhook gets webhook from specified repo.
func (g *Github) GetWebhook(repo string, webhookURL string) (*github.Hook, error) {
	owner, name := scm.ParseRepo(repo)
	hooks, _, err := g.client.Repositories.ListHooks(g.ctx, owner, name, nil)
	if err != nil {
		return nil, convertGithubError(err)
	}

	for _, hook := range hooks {
		if hookurl, ok := hook.Config["url"].(string); ok {
			if hookurl == webhookURL {
				return hook, nil
			}
		}
	}

	return nil, cerr.ErrorContentNotFound.Error(fmt.Sprintf("webhook url %s", webhookURL))
}

// CreateWebhook creates webhook for specified repo.
func (g *Github) CreateWebhook(repo string, webhook *scm.Webhook) error {
	if webhook == nil || len(webhook.URL) == 0 || len(webhook.Events) == 0 {
		return fmt.Errorf("the webhook %v is not correct", webhook)
	}

	_, err := g.GetWebhook(repo, webhook.URL)
	if err != nil {
		if !cerr.ErrorContentNotFound.Derived(err) {
			return err
		}

		// Hook name must be passed as "web".
		// Ref: https://developer.github.com/v3/repos/hooks/#create-a-hook
		hook := github.Hook{
			Events: convertToGithubEvents(webhook.Events),
			Config: map[string]interface{}{
				"url":          webhook.URL,
				"content_type": "json",
			},
		}
		owner, name := scm.ParseRepo(repo)
		_, _, err = g.client.Repositories.CreateHook(g.ctx, owner, name, &hook)
		if err != nil {
			return convertGithubError(err)
		}
		return nil
	}

	log.Warningf("Webhook already existed: %+v", webhook)
	return nil
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

// DeleteWebhook deletes webhook from specified repo.
func (g *Github) DeleteWebhook(repo string, webhookURL string) error {
	hook, err := g.GetWebhook(repo, webhookURL)
	if err != nil {
		return err
	}

	owner, name := scm.ParseRepo(repo)
	if _, err = g.client.Repositories.DeleteHook(g.ctx, owner, name, *hook.ID); err != nil {
		log.Errorf("delete hook %s for %s/%s error: %v", hook.ID, owner, name, err)
		return convertGithubError(err)
	}

	return nil
}

// ParseEvent parses data from Github events.
func ParseEvent(scmCfg *v1alpha1.SCMSource, request *http.Request) *scm.EventData {
	payload, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	event, err := github.ParseWebHook(github.WebHookType(request), payload)
	if err != nil {
		log.Errorf("Failed to parse Github webhook as %v", err)
		return nil
	}

	switch event := event.(type) {
	case *github.ReleaseEvent:
		// Only handle when the tag is created.
		// When you create a new release in your GitHub, GitHub will send
		// two release X-GitHub-Event(and one push X-GitHub-Event, not discussed here):
		// - one action is 'created'
		// - the other one is 'published'
		// So we only process the 'created' one, otherwise, one release will trigger a workflow twice.
		action := *event.Action
		if action != "created" {
			log.Warningf("Skip unsupported action %s of Github release event, only support created action.", action)
			return nil
		}
		return &scm.EventData{
			Type: scm.TagReleaseEventType,
			Repo: *event.Repo.FullName,
			Ref:  fmt.Sprintf(tagRefTemplate, *event.Release.TagName),
		}
	case *github.PullRequestEvent:
		// Only handle when the pull request are created.
		action := *event.Action
		if action != "opened" && action != "synchronize" {
			log.Warningf("Skip unsupported action %s of Github pull request event, only support opened and synchronize action.", action)
			return nil
		}
		commitSHA, err := extractCommitSHA(*event.PullRequest.StatusesURL)
		if err != nil {
			log.Errorf("Fail to get commit SHA: %v", err)
			return nil
		}
		return &scm.EventData{
			Type:      scm.PullRequestEventType,
			Repo:      *event.Repo.FullName,
			Ref:       fmt.Sprintf(pullRefTemplate, *event.PullRequest.Number),
			CommitSHA: commitSHA,
		}
	case *github.IssueCommentEvent:
		if event.Issue.PullRequestLinks == nil {
			log.Warningln("Only handle comments on pull requests.")
			return nil
		}

		// Only handle when the pull request comments are created.
		if *event.Action != "created" {
			log.Warningln("Only handle comments when they are created.")
			return nil
		}

		if *event.Issue.State != "open" {
			log.Warningln("Only handle comments on opened pull requests.")
			return nil
		}

		issueNumber := *event.Issue.Number
		commitSHA, err := getLastCommitSHA(scmCfg, *event.Repo.FullName, issueNumber)
		if err != nil {
			log.Errorf("Failed to get latest commit SHA for issue %d", issueNumber)
			return nil
		}

		return &scm.EventData{
			Type:      scm.PullRequestCommentEventType,
			Repo:      *event.Repo.FullName,
			Ref:       fmt.Sprintf(pullRefTemplate, issueNumber),
			Comment:   *event.Comment.Body,
			CommitSHA: commitSHA,
		}
	case *github.PushEvent:
		return &scm.EventData{
			Type:   scm.PushEventType,
			Repo:   *event.Repo.FullName,
			Ref:    *event.Ref,
			Branch: *event.Ref,
		}
	default:
		log.Warningln("Skip unsupported Github event")
		return nil
	}
}

// input   : `https://api.github.com/repos/aaa/bbb/statuses/ccc`
// output  : ccc
func extractCommitSHA(url string) (string, error) {
	results := statusURLRegexp.FindStringSubmatch(url)
	if len(results) < 2 {
		return "", fmt.Errorf("statusesURL is invalid")
	}
	return results[1], nil
}

func getLastCommitSHA(scmCfg *v1alpha1.SCMSource, repo string, number int) (string, error) {
	p, err := scm.GetSCMProvider(scmCfg)
	if err != nil {
		return "", err
	}

	return p.GetPullRequestSHA(repo, number)
}

func convertGithubError(err error) error {
	if err == nil {
		return nil
	}

	value := reflect.ValueOf(err).Elem()
	respField := value.FieldByName(errorFieldResponse)
	if !respField.IsValid() {
		log.Warningf("response filed of Github error is invalid: %v", err)
		return err
	}

	resp, ok := respField.Interface().(*http.Response)
	if !ok {
		return err
	}

	if resp != nil && resp.StatusCode == http.StatusInternalServerError {
		return cerr.ErrorExternalSystemError.Error("GitHub", err)
	}

	if resp != nil && resp.StatusCode == http.StatusUnauthorized {
		return cerr.ErrorExternalAuthorizationFailed.Error(err)
	}

	if resp != nil && resp.StatusCode == http.StatusForbidden {
		return cerr.ErrorExternalAuthenticationFailed.Error(err)
	}

	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return cerr.ErrorExternalNotFound.Error(err)
	}

	return err
}
