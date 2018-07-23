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

package gitlab

import (
	"fmt"
	"strings"

	log "github.com/golang/glog"
	gitlabv3 "github.com/xanzy/go-gitlab"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
	"github.com/caicloud/cyclone/pkg/scm/provider"
)

// GitlabV3 represents the SCM provider of GitlabV3 with API V3.
type GitlabV3 struct {
	scmCfg *api.SCMConfig
	client *gitlabv3.Client
}

// GetToken gets the token by the username and password of SCM config.
func (g *GitlabV3) GetToken() (string, error) {
	return getOauthToken(g.scmCfg)
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token.
func (g *GitlabV3) CheckToken() bool {
	if _, err := g.listReposInner(false); err != nil {
		return false
	}
	return true
}

// ListRepos lists the repos by the SCM config.
func (g *GitlabV3) ListRepos() ([]api.Repository, error) {
	return g.listReposInner(true)
}

// listReposInner lists the projects by the SCM config,
// list all projects while the parameter 'listAll' is true,
// otherwise, list projects by default 'ListPerPageOpt' number.
func (g *GitlabV3) listReposInner(listAll bool) ([]api.Repository, error) {
	opt := &gitlabv3.ListProjectsOptions{
		ListOptions: gitlabv3.ListOptions{
			PerPage: provider.ListPerPageOpt,
		},
	}

	// Get all pages of results.
	var allProjects []*gitlabv3.Project
	for {
		projects, resp, err := g.client.Projects.ListProjects(opt)
		if err != nil {
			return nil, err
		}

		allProjects = append(allProjects, projects...)
		if resp.NextPage == 0 || !listAll {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	repos := make([]api.Repository, len(allProjects))
	for i, repo := range allProjects {
		repos[i].Name = repo.PathWithNamespace
		repos[i].URL = repo.HTTPURLToRepo
	}

	return repos, nil
}

// ListBranches lists the branches for specified repo.
func (g *GitlabV3) ListBranches(repo string) ([]string, error) {
	branches, _, err := g.client.Branches.ListBranches(repo)
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
func (g *GitlabV3) ListTags(repo string) ([]string, error) {
	tags, _, err := g.client.Tags.ListTags(repo)
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
func (g *GitlabV3) CreateWebHook(repoURL string, webHook *scm.WebHook) error {
	if webHook == nil || len(webHook.Url) == 0 || len(webHook.Events) == 0 {
		return fmt.Errorf("The webhook %v is not correct", webHook)
	}

	enableState, disableState := true, false
	// Push event is enable for Gitlab webhook in default, so need to remove this default option.
	hook := gitlabv3.AddProjectHookOptions{
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

	onwer, name := provider.ParseRepoURL(repoURL)
	_, _, err := g.client.Projects.AddProjectHook(onwer+"/"+name, &hook)
	log.Error(err)
	return err
}

// DeleteWebHook deletes webhook from specified repo.
func (g *GitlabV3) DeleteWebHook(repoURL string, webHookUrl string) error {
	owner, name := provider.ParseRepoURL(repoURL)
	hooks, _, err := g.client.Projects.ListProjectHooks(owner+"/"+name, nil)
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		if strings.HasPrefix(hook.URL, webHookUrl) {
			_, err = g.client.Projects.DeleteProjectHook(owner+"/"+name, hook.ID)
			return nil
		}
	}

	return nil
}

// NewTagFromLatest generate a new tag
func (g *GitlabV3) NewTagFromLatest(tagName, description, commitID, url string) error {
	owner, name := provider.ParseRepoURL(url)
	tag := &gitlabv3.CreateTagOptions{
		TagName: &tagName,
		Ref:     &commitID,
		Message: &description,
	}

	_, _, err := g.client.Tags.CreateTag(owner+"/"+name, tag)
	log.Error(err)
	return err
}

// CreateStatuses generate a new status for repository.
func (g *GitlabV3) CreateStatuses(state, description, targetURL, statusesURL string) error {
	owner, project, commitSha, err := splitStatusesURL(statusesURL)
	if err != nil {
		return err
	}

	context := "continuous-integration/cyclone"

	status := &gitlabv3.SetCommitStatusOptions{
		State:       gitlabv3.BuildState(state),
		Description: &description,
		TargetURL:   &targetURL,
		Context:     &context,
	}

	_, _, err = g.client.Commits.SetCommitStatus(owner+"/"+project, commitSha, status)
	log.Error(err)
	return nil
}
