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
	v3 "github.com/xanzy/go-gitlab"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// V3 represents the SCM provider of API V3 GitLab.
type V3 struct {
	scmCfg *v1alpha1.SCMSource
	client *v3.Client
}

// GetToken gets the token by the username and password of SCM config.
func (g *V3) GetToken() (string, error) {
	return getOauthToken(g.scmCfg)
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token.
func (g *V3) CheckToken() error {
	if _, err := g.listReposInner(false); err != nil {
		return err
	}
	return nil
}

// ListRepos lists the repos by the SCM config.
func (g *V3) ListRepos() ([]scm.Repository, error) {
	return g.listReposInner(true)
}

// listReposInner lists the projects by the SCM config,
// list all projects while the parameter 'listAll' is true,
// otherwise, list projects by default 'ListPerPageOpt' number.
func (g *V3) listReposInner(listAll bool) ([]scm.Repository, error) {
	opt := &v3.ListProjectsOptions{
		ListOptions: v3.ListOptions{
			PerPage: scm.ListPerPageOpt,
		},
	}

	// Get all pages of results.
	var allProjects []*v3.Project
	for {
		projects, resp, err := g.client.Projects.ListProjects(opt)
		if err != nil {
			if resp.StatusCode == 500 {
				return nil, cerr.ErrorSCMServerInternalError.Error(g.scmCfg.Server, err)
			}
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
func (g *V3) ListBranches(repo string) ([]string, error) {
	branches, resp, err := g.client.Branches.ListBranches(repo, nil)
	if err != nil {
		log.Errorf("Fail to list branches for %s", repo)
		if resp.StatusCode == 500 {
			return nil, cerr.ErrorSCMServerInternalError.Error(g.scmCfg.Server, err)
		}
		return nil, err
	}

	branchNames := make([]string, len(branches))
	for i, branch := range branches {
		branchNames[i] = branch.Name
	}

	return branchNames, nil
}

// ListTags lists the tags for specified repo.
func (g *V3) ListTags(repo string) ([]string, error) {
	tags, resp, err := g.client.Tags.ListTags(repo, nil)
	if err != nil {
		log.Errorf("Fail to list tags for %s", repo)
		if resp.StatusCode == 500 {
			return nil, cerr.ErrorSCMServerInternalError.Error(g.scmCfg.Server, err)
		}
		return nil, err
	}

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	return tagNames, nil
}

// ListDockerfiles lists the Dockerfiles for specified repo.
func (g *V3) ListDockerfiles(repo string) ([]string, error) {
	// List Dockerfiles in a project with gitlab v3 api is very inefficient.
	// There is not a proper api can be used to do this with GitLab v3.
	//
	// FYI:
	// https://stackoverflow.com/questions/25127695/search-filenames-with-gitlab-api
	return nil, fmt.Errorf("list gitlab v3 dockerfiles not implemented")
}

// CreateWebhook creates webhook for specified repo.
func (g *V3) CreateWebhook(repo string, webhook *scm.Webhook) error {
	if webhook == nil || len(webhook.URL) == 0 || len(webhook.Events) == 0 {
		return fmt.Errorf("The webhook %v is not correct", webhook)
	}

	enableState, disableState := true, false
	// Push event is enable for Gitlab webhook in default, so need to remove this default option.
	hook := v3.AddProjectHookOptions{
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
	_, resp, err := g.client.Projects.AddProjectHook(onwer+"/"+name, &hook)
	if resp.StatusCode == 500 {
		return cerr.ErrorSCMServerInternalError.Error(g.scmCfg.Server, err)
	}
	return err
}

// DeleteWebhook deletes webhook from specified repo.
func (g *V3) DeleteWebhook(repo string, webhookURL string) error {
	owner, name := scm.ParseRepo(repo)
	hooks, resp, err := g.client.Projects.ListProjectHooks(owner+"/"+name, nil)
	if err != nil {
		if resp.StatusCode == 500 {
			return cerr.ErrorSCMServerInternalError.Error(g.scmCfg.Server, err)
		}
		return err
	}

	for _, hook := range hooks {
		if strings.HasPrefix(hook.URL, webhookURL) {
			if _, err = g.client.Projects.DeleteProjectHook(owner+"/"+name, hook.ID); err != nil {
				log.Errorf("delete project hook %s for %s/%s error: %v", hook.ID, owner, name, err)
				return err
			}
		}
	}

	return nil
}
