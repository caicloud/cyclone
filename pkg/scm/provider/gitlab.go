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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/golang/glog"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
)

// gitLabServer represents the server address for public GitLab.
const gitLabServer = "https://gitlab.com"

// GitLab represents the SCM provider of GitLab.
type GitLab struct{}

func init() {
	if err := scm.RegisterProvider(api.GitLab, new(GitLab)); err != nil {
		log.Errorln(err)
	}
}

// GetToken gets the token by the username and password of SCM config.
func (g *GitLab) GetToken(scm *api.SCMConfig) (string, error) {
	if len(scm.Username) == 0 || len(scm.Password) == 0 {
		return "", fmt.Errorf("GitHub username or password is missing")
	}

	bodyData := struct {
		GrantType string `json:"grant_type"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}{
		GrantType: "password",
		Username:  scm.Username,
		Password:  scm.Password,
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return "", fmt.Errorf("fail to new request body for token as %s", err.Error())
	}

	// If use the public Gitlab, must use the HTTPS protocol.
	if strings.Contains(scm.Server, "gitlab.com") && strings.HasPrefix(scm.Server, "http://") {
		log.Infof("Convert SCM server from %s to %s to use HTTPS protocol for public Gitlab", scm.Server, gitLabServer)
		scm.Server = gitLabServer
	}

	tokenURL := fmt.Sprintf("%s%s", scm.Server, "/oauth/token")
	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Errorf("Fail to new the request for token as %s", err.Error())
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to request for token as %s", err.Error())
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to request for token as %s", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		var token oauth2.Token
		err := json.Unmarshal(body, &token)
		if err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}

	err = fmt.Errorf("Fail to request for token as %s", body)
	return "", err
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token.
func (g *GitLab) CheckToken(scm *api.SCMConfig) bool {
	if _, err := g.ListRepos(scm); err != nil {
		return false
	}
	return true
}

// ListRepos lists the repos by the SCM config.
func (g *GitLab) ListRepos(scm *api.SCMConfig) ([]api.Repository, error) {
	client, err := newGitLabClient(scm.Server, scm.Username, scm.Token)
	if err != nil {
		return nil, err
	}

	opt := &gitlab.ListProjectsOptions{}
	projects, _, err := client.Projects.ListProjects(opt)
	if err != nil {
		log.Errorf("Fail to list projects for %s", scm.Username)
		return nil, err
	}

	repos := make([]api.Repository, len(projects))
	for i, repo := range projects {
		repos[i].Name = repo.PathWithNamespace
		repos[i].URL = repo.HTTPURLToRepo
	}

	return repos, nil
}

// ListBranches lists the branches for specified repo.
func (g *GitLab) ListBranches(scm *api.SCMConfig, repo string) ([]string, error) {
	client, err := newGitLabClient(scm.Server, scm.Username, scm.Token)
	if err != nil {
		return nil, err
	}

	branches, _, err := client.Branches.ListBranches(repo)
	if err != nil {
		log.Errorf("Fail to list branches for %s", repo)
		return nil, err
	}

	branchNames := make([]string, len(branches))
	for i, branch := range branches {
		branchNames[i] = branch.Name
	}

	return branchNames, nil
}

// ListTags lists the tags for specified repo.
func (g *GitLab) ListTags(scm *api.SCMConfig, repo string) ([]string, error) {
	client, err := newGitLabClient(scm.Server, scm.Username, scm.Token)
	if err != nil {
		return nil, err
	}

	tags, _, err := client.Tags.ListTags(repo)
	if err != nil {
		log.Errorf("Fail to list tags for %s", repo)
		return nil, err
	}

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return tagNames, nil
}

// CreateWebHook creates webhook for specified repo.
func (g *GitLab) CreateWebHook(cfg *api.SCMConfig, repoURL string, webHook *scm.WebHook) error {
	if webHook == nil || len(webHook.Url) == 0 || len(webHook.Events) == 0 {
		return fmt.Errorf("The webhook %v is not correct", webHook)
	}

	client, err := newGitLabClient(cfg.Server, cfg.Username, cfg.Token)
	if err != nil {
		return err
	}

	enableState, disableState := true, false
	// Push event is enable for Gitlab webhook in default, so need to remove this default option.
	hook := gitlab.AddProjectHookOptions{
		PushEvents: &disableState,
	}

	for _, e := range webHook.Events {
		switch e {
		case scm.PullRequestEventType:
			hook.MergeRequestsEvents = &enableState
		case scm.PullRequestCommentEventType:
			hook.NoteEvents = &enableState
		case scm.PushEventType:
			hook.PushEvents = &enableState
		case scm.TagReleaseEventType:
			hook.TagPushEvents = &enableState
		default:
			log.Errorf("The event type %s is not supported, will be ignored", e)
			return nil
		}
	}
	hook.URL = &webHook.Url

	onwer, name := parseURL(repoURL)
	_, _, err = client.Projects.AddProjectHook(onwer+"/"+name, &hook)
	return err
}

// DeleteWebHook deletes webhook from specified repo.
func (g *GitLab) DeleteWebHook(cfg *api.SCMConfig, repoURL string, webHookUrl string) error {
	client, err := newGitLabClient(cfg.Server, cfg.Username, cfg.Token)
	if err != nil {
		return err
	}

	owner, name := parseURL(repoURL)
	hooks, _, err := client.Projects.ListProjectHooks(owner+"/"+name, nil)
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		if strings.HasPrefix(hook.URL, webHookUrl) {
			_, err = client.Projects.DeleteProjectHook(owner+"/"+name, hook.ID)
			return nil
		}
	}

	return nil
}

// newGitLabClient news GitLab client by token.If username is empty, use private-token instead of oauth2.0 token.
func newGitLabClient(server, username, token string) (*gitlab.Client, error) {
	var client *gitlab.Client

	if len(username) == 0 {
		client = gitlab.NewClient(nil, token)
	} else {
		client = gitlab.NewOAuthClient(nil, token)
	}

	if err := client.SetBaseURL(server + "/api/v3/"); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return client, nil
}

// NewTagFromLatest generate a new tag
func (g *GitLab) NewTagFromLatest(tagName, description, commitID, url, token string) error {
	owner, name := parseURL(url)
	ref := "master"
	tag := &gitlab.CreateTagOptions{
		TagName: &tagName,
		Ref:     &ref,
		Message: &description,
	}

	gitlabServer := "https://gitlab.com"
	client := gitlab.NewOAuthClient(nil, token)
	client.SetBaseURL(gitlabServer + "/api/v3/")

	_, _, err := client.Tags.CreateTag(owner+"/"+name, tag)
	return err
}
