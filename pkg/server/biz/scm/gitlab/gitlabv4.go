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
	"net/http"

	"github.com/caicloud/nirvana/log"
	v4 "gopkg.in/xanzy/go-gitlab.v0"

	c_v1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// V4 represents the SCM provider of API V4 GitLab.
type V4 struct {
	scmCfg *v1alpha1.SCMSource
	client *v4.Client
}

// GetToken gets the token by the username and password of SCM config.
func (g *V4) GetToken() (string, error) {
	return getOauthToken(g.scmCfg)
}

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token.
func (g *V4) CheckToken() error {
	if _, err := g.listReposInner(false); err != nil {
		return err
	}
	return nil
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
	opt := &v4.ListProjectsOptions{
		ListOptions: v4.ListOptions{
			PerPage: scm.ListOptPerPage,
		},
		Membership: &trueVar,
	}

	// Get all pages of results.
	var allProjects []*v4.Project
	for {
		projects, resp, err := g.client.Projects.ListProjects(opt)
		if err != nil {
			return nil, convertGitlabV4Error(err, resp)
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
	opts := &v4.ListBranchesOptions{
		PerPage: scm.ListOptPerPage,
	}

	var allBranches []string
	for {
		branches, resp, err := g.client.Branches.ListBranches(repo, opts)
		if err != nil {
			log.Errorf("Fail to list branches for %s", repo)
			return nil, convertGitlabV4Error(err, resp)
		}

		for _, b := range branches {
			allBranches = append(allBranches, b.Name)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allBranches, nil
}

// ListTags lists the tags for specified repo.
func (g *V4) ListTags(repo string) ([]string, error) {
	opts := &v4.ListTagsOptions{
		ListOptions: v4.ListOptions{
			PerPage: scm.ListOptPerPage,
		},
	}

	var allTags []string
	for {
		tags, resp, err := g.client.Tags.ListTags(repo, opts)
		if err != nil {
			log.Errorf("Fail to list tags for %s", repo)
			return nil, convertGitlabV4Error(err, resp)
		}

		for _, t := range tags {
			allTags = append(allTags, t.Name)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allTags, nil
}

// ListDockerfiles lists the Dockerfiles for specified repo.
func (g *V4) ListDockerfiles(repo string) ([]string, error) {
	recursive := true
	opt := &v4.ListTreeOptions{
		Recursive: &recursive,
		ListOptions: v4.ListOptions{
			PerPage: scm.ListOptPerPage,
		},
	}

	treeNodes := []*v4.TreeNode{}
	for {
		treeNode, resp, err := g.client.Repositories.ListTree(repo, opt)
		if err != nil {
			log.Errorf("Fail to list dockerfile for %s", repo)
			return nil, convertGitlabV4Error(err, resp)
		}

		treeNodes = append(treeNodes, treeNode...)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	files := []string{}
	for _, t := range treeNodes {
		if t.Type == "blob" && scm.IsDockerfile(t.Path) {
			files = append(files, t.Path)
		}
	}

	return files, nil
}

// CreateStatus generate a new status for repository.
func (g *V4) CreateStatus(status c_v1alpha1.StatusPhase, targetURL, repo, commitSha string) error {
	state, description := transStatus(status)

	context := "continuous-integration/cyclone"
	opt := &v4.SetCommitStatusOptions{
		State:       v4.BuildStateValue(state),
		Description: &description,
		TargetURL:   &targetURL,
		Context:     &context,
	}
	_, resp, err := g.client.Commits.SetCommitStatus(repo, commitSha, opt)
	return convertGitlabV4Error(err, resp)
}

// GetPullRequestSHA gets latest commit SHA of pull request.
func (g *V4) GetPullRequestSHA(repo string, number int) (string, error) {
	mr, resp, err := g.client.MergeRequests.GetMergeRequest(repo, number, nil)
	if err != nil {
		return "", convertGitlabV4Error(err, resp)
	}

	return mr.SHA, nil
}

// GetWebhook gets webhook from specified repo.
func (g *V4) GetWebhook(repo string, webhookURL string) (*v4.ProjectHook, error) {
	// log.Infof("repo: %s", url.PathEscape(repo))
	hooks, resp, err := g.client.Projects.ListProjectHooks(repo, nil)
	if err != nil {
		return nil, convertGitlabV4Error(err, resp)
	}

	for _, hook := range hooks {
		if hook.URL == webhookURL {
			return hook, nil
		}
	}

	return nil, cerr.ErrorContentNotFound.Error(fmt.Sprintf("webhook url %s", webhookURL))
}

// CreateWebhook creates webhook for specified repo.
func (g *V4) CreateWebhook(repo string, webhook *scm.Webhook) error {
	if webhook == nil || len(webhook.URL) == 0 || len(webhook.Events) == 0 {
		return fmt.Errorf("The webhook %v is not correct", webhook)
	}

	_, err := g.GetWebhook(repo, webhook.URL)
	if err != nil {
		if !cerr.ErrorContentNotFound.Derived(err) {
			return err
		}

		log.Infof("webhook url: %s", webhook.URL)
		hook := generateV4ProjectHook(webhook)
		_, resp, err := g.client.Projects.AddProjectHook(repo, hook)
		if err != nil {
			return convertGitlabV4Error(err, resp)
		}
		return nil
	}

	log.Warningf("Webhook already existed: %+v", webhook)
	return err
}

// DeleteWebhook deletes webhook from specified repo.
func (g *V4) DeleteWebhook(repo string, webhookURL string) error {
	hook, err := g.GetWebhook(repo, webhookURL)
	if err != nil {
		return err
	}

	if resp, err := g.client.Projects.DeleteProjectHook(repo, hook.ID); err != nil {
		log.Errorf("delete project hook %s for %s/%s error: %v", hook.ID, repo, err)
		return convertGitlabV4Error(err, resp)
	}

	return nil
}

func generateV4ProjectHook(webhook *scm.Webhook) *v4.AddProjectHookOptions {
	enableState, disableState := true, false
	// Push event is enable for Gitlab webhook in default, so need to remove this default option.
	hook := &v4.AddProjectHookOptions{
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

	return hook
}

func convertGitlabV4Error(err error, resp *v4.Response) error {
	if err == nil {
		return nil
	}

	if resp != nil && resp.StatusCode == http.StatusInternalServerError {
		return cerr.ErrorSCMServerInternalError.Error(err)
	}

	return err
}
