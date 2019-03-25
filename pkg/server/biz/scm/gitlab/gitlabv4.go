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

	"github.com/caicloud/nirvana/log"
	"gopkg.in/xanzy/go-gitlab.v0"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
)

// V4 represents the SCM provider of API V4 GitLab.
type V4 struct {
	scmCfg *v1alpha1.SCMSource
	client *gitlab.Client
}

func newGitlabV4(scmCfg *v1alpha1.SCMSource) (scm.Provider, error) {
	client, err := newGitlabV4Client(scmCfg.Server, scmCfg.User, scmCfg.Token)
	if err != nil {
		log.Errorf("fail to new gitlab client as %v", err)
		return nil, err
	}

	return &V4{scmCfg, client}, nil
}

// GetToken gets the token by the username and password of SCM config.
func (g *V4) GetToken() (string, error) {
	return getOauthToken(g.scmCfg)
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token.
func (g *V4) CheckToken() bool {
	if _, err := g.listReposInner(false); err != nil {
		return false
	}
	return true
}

// ListRepos lists the repos by the SCM config.
func (g *V4) ListRepos() ([]scm.Repository, error) {
	return g.listReposInner(true)
}

// listReposInner lists the projects by the SCM config,
// list all projects while the parameter 'listAll' is true,
// otherwise, list projects by default 'provider.ListPerPageOpt' number.
func (g *V4) listReposInner(listAll bool) ([]scm.Repository, error) {
	trueVar := true
	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: scm.ListPerPageOpt,
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

	repos := make([]scm.Repository, len(allProjects))
	for i, repo := range allProjects {
		repos[i].Name = repo.PathWithNamespace
		repos[i].URL = repo.HTTPURLToRepo
	}

	return repos, nil
}

// ListBranches lists the branches for specified repo.
func (g *V4) ListBranches(repo string) ([]string, error) {
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
func (g *V4) ListTags(repo string) ([]string, error) {
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

// ListDockerfiles lists the Dockerfiles for specified repo.
func (g *V4) ListDockerfiles(repo string) ([]string, error) {
	recursive := true
	opt := &gitlab.ListTreeOptions{
		Recursive: &recursive,
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	treeNodes := []*gitlab.TreeNode{}
	for {
		treeNode, resp, err := g.client.Repositories.ListTree(repo, opt)
		if err != nil {
			log.Errorf("Fail to list dockerfile for %s", repo)
			return nil, err
		}

		treeNodes = append(treeNodes, treeNode...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	files := []string{}
	for _, t := range treeNodes {
		if t.Type == "blob" && t.Name == "Dockerfile" {
			files = append(files, t.Path)
		}
	}

	return files, nil
}

// CreateWebhook creates webhook for specified repo.
func (g *V4) CreateWebhook(repo string, webhook *scm.Webhook) error {
	if webhook == nil || len(webhook.URL) == 0 || len(webhook.Events) == 0 {
		return fmt.Errorf("The webhook %v is not correct", webhook)
	}

	enableState, disableState := true, false
	// Push event is enable for Gitlab webhook in default, so need to remove this default option.
	hook := gitlab.AddProjectHookOptions{
		PushEvents: &disableState,
	}

	for _, e := range webhook.Events {
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
	hook.URL = &webhook.URL

	onwer, name := scm.ParseRepo(repo)
	_, _, err := g.client.Projects.AddProjectHook(onwer+"/"+name, &hook)
	return err
}

// DeleteWebhook deletes webhook from specified repo.
func (g *V4) DeleteWebhook(repo string, webhookURL string) error {
	owner, name := scm.ParseRepo(repo)
	hooks, _, err := g.client.Projects.ListProjectHooks(owner+"/"+name, nil)
	if err != nil {
		return err
	}

	for _, hook := range hooks {
		if strings.HasPrefix(hook.URL, webhookURL) {
			if _, err = g.client.Projects.DeleteProjectHook(owner+"/"+name, hook.ID); err != nil {
				return err
			}
		}
	}

	return nil
}
