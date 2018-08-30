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
	"gopkg.in/xanzy/go-gitlab.v0"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
	"github.com/caicloud/cyclone/pkg/scm/provider"
)

// GitlabV4 represents the SCM provider of Gitlab with API V4.
type GitlabV4 struct {
	scmCfg *api.SCMConfig
	client *gitlab.Client
}

func NewGitlabV4(scmCfg *api.SCMConfig) (scm.SCMProvider, error) {
	client, err := newGitlabV4Client(scmCfg.Server, scmCfg.Username, scmCfg.Token)
	if err != nil {
		log.Error("fail to new gitlab client as %v", err)
		return nil, err
	}

	return &GitlabV4{scmCfg, client}, nil
}

// GetToken gets the token by the username and password of SCM config.
func (g *GitlabV4) GetToken() (string, error) {
	return getOauthToken(g.scmCfg)
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token.
func (g *GitlabV4) CheckToken() bool {
	if _, err := g.listReposInner(false); err != nil {
		return false
	}
	return true
}

// ListRepos lists the repos by the SCM config.
func (g *GitlabV4) ListRepos() ([]api.Repository, error) {
	return g.listReposInner(true)
}

// listReposInner lists the projects by the SCM config,
// list all projects while the parameter 'listAll' is true,
// otherwise, list projects by default 'provider.ListPerPageOpt' number.
func (g *GitlabV4) listReposInner(listAll bool) ([]api.Repository, error) {
	trueVar := true
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: provider.ListPerPageOpt,
		},
		Membership: &trueVar,
	}

	// Get all pages of results.
	var allProjects []*gitlab.Project
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
func (g *GitlabV4) ListBranches(repo string) ([]string, error) {
	opts := &gitlab.ListBranchesOptions{}
	branches, _, err := g.client.Branches.ListBranches(repo, opts)
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
func (g *GitlabV4) ListTags(repo string) ([]string, error) {
	opts := &gitlab.ListTagsOptions{}
	tags, _, err := g.client.Tags.ListTags(repo, opts)
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
func (g *GitlabV4) CreateWebHook(repoURL string, webHook *scm.WebHook) error {
	if webHook == nil || len(webHook.Url) == 0 || len(webHook.Events) == 0 {
		return fmt.Errorf("The webhook %v is not correct", webHook)
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

	onwer, name := provider.ParseRepoURL(repoURL)
	_, _, err := g.client.Projects.AddProjectHook(onwer+"/"+name, &hook)
	log.Error(err)
	return err
}

// DeleteWebHook deletes webhook from specified repo.
func (g *GitlabV4) DeleteWebHook(repoURL string, webHookUrl string) error {
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
func (g *GitlabV4) NewTagFromLatest(tagName, description, commitID, url string) error {
	owner, name := provider.ParseRepoURL(url)
	tag := &gitlab.CreateTagOptions{
		TagName: &tagName,
		Ref:     &commitID,
		Message: &description,
	}

	_, _, err := g.client.Tags.CreateTag(owner+"/"+name, tag)
	log.Error(err)
	return err
}

// GetRepoType get type of repo
func (g *GitlabV4) GetRepoType(repo string) (string, error) {
	languages, err := getLanguages(g.scmCfg, v4APIVersion, repo)
	if err != nil {
		log.Error("list language failed:%v", err)
		return "", err
	}
	language := getTopLanguage(languages)

	switch language {
	case api.JavaRepoType:
		files, err := getContents(g.scmCfg, v4APIVersion, repo)
		if err != nil {

		}
		log.Error(err)
		for _, f := range files {
			if strings.Contains(f.Name, "pom.xml") {
				return api.MavenRepoType, nil
			}
			if strings.Contains(f.Name, "build.gradle") {
				return api.GradleRepoType, nil
			}
		}
	case api.JavaScriptRepoType:
		files, err := getContents(g.scmCfg, v4APIVersion, repo)
		if err != nil {
			log.Error("get contents failed:%v", err)
			return language, nil
		}

		for _, f := range files {
			if strings.Contains(f.Name, "package.json") {
				return api.NodeRepoType, nil
			}
		}
	}

	return language, nil
}
