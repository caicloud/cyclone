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

package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	log "github.com/golang/glog"
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"

	"github.com/caicloud/cyclone/pkg/api"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	gitlabuitl "github.com/caicloud/cyclone/pkg/util/gitlab"
)

const (
	// branchRefTemplate represents reference template for branches.
	branchRefTemplate = "refs/heads/%s"

	// tagRefTemplate represents reference template for tags.
	tagRefTemplate = "refs/tags/%s"

	// githubPullRefTemplate represents reference template for Github pull request.
	githubPullRefTemplate = "refs/pull/%d/merge"

	// gitlabMergeRefTemplate represents reference template for Gitlab merge request and merge target branch
	gitlabMergeRefTemplate = "refs/merge-requests/%d/head:%s"

	// gitlabEventTypeHeader represents the Gitlab header key used to pass the event type.
	gitlabEventTypeHeader = "X-Gitlab-Event"
)

// githubRepoNameRegexp represents the regexp of github status url.
var githubStatusURLRegexp *regexp.Regexp

func init() {
	var statusURLRegexp = `^https://api.github.com/repos/[\S]+/[\S]+/statuses/([\w]+)$`
	githubStatusURLRegexp = regexp.MustCompile(statusURLRegexp)
}

// HandleGithubWebhook handles the webhook request from Github.
// 1. Parse the pipeline id from request path;
// 2. Get the pipeline by id;
// 3. Parse the payload from request header;
// 4. Parse the payload from request body;
// 5. First get the event type, and handle it according to its type.
func HandleGithubWebhook(ctx context.Context, pipelineID string) (webhookResponse, error) {
	response := webhookResponse{}
	pipeline, err := pipelineManager.GetPipelineByID(pipelineID)
	if err != nil {
		return response, err
	}

	if pipeline.AutoTrigger == nil || pipeline.AutoTrigger.SCMTrigger == nil {
		response.Message = "SCM auto trigger is not set"
		return response, nil
	}
	scmTrigger := pipeline.AutoTrigger.SCMTrigger

	// TODO (robin) Validate the payload.
	// Ref: https://github.com/google/go-github/blob/df47db1628185875602e66d3356ae7337b52bba3/github/messages.go#L213-L233
	//github.ValidatePayload(contextutil.GetHttpRequest(ctx), secretKey)
	request := contextutil.GetHttpRequest(ctx)
	payload, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return response, fmt.Errorf("Fail to read the request body")
	}
	event, err := github.ParseWebHook(github.WebHookType(request), payload)
	if err != nil {
		return response, err
	}
	log.Infof("Github webhook event: %v", event)

	// Handle the event.
	var performParams *api.PipelinePerformParams
	var commitSHA string
	trigger := api.TriggerSCM

	switch event := event.(type) {
	case *github.ReleaseEvent:
		if scmTrigger.TagRelease == nil {
			response.Message = "Release trigger is not enabled"
			return response, nil
		}
		trigger = api.TriggerWebhookTagRelease

		performParams = &api.PipelinePerformParams{
			Name:        *event.Release.TagName,
			Ref:         fmt.Sprintf(tagRefTemplate, *event.Release.TagName),
			Description: "Triggered by tag release",
			Stages:      scmTrigger.TagRelease.Stages,
		}

		log.Info("Triggered by Github release event")
	case *github.PullRequestEvent:
		// Only handle when the pull request are created.
		if *event.Action != "opened" && *event.Action != "synchronize" {
			response.Message = "Only handle when pull request is created or synchronized"
			return response, nil
		}

		if scmTrigger.PullRequest == nil {
			response.Message = "Pull request trigger is not enabled"
			return response, nil
		}

		commitSHA, err = extractCommitSha(*event.PullRequest.StatusesURL)
		if err != nil {
			response.Message = "Get last commit sha failed"
			return response, err
		}
		trigger = api.TriggerWebhookPullRequest

		performParams = &api.PipelinePerformParams{
			Ref:         fmt.Sprintf(githubPullRefTemplate, *event.PullRequest.Number),
			Description: "Triggered by pull request",
			Stages:      scmTrigger.PullRequest.Stages,
		}

		log.Info("Triggered by Github pull request event")
	case *github.IssueCommentEvent:
		if event.Issue.PullRequestLinks == nil {
			log.Infof("Only handle when issues type is pull request")
			response.Message = "Only handle when issues type is pull request"
			return response, nil
		}

		// Only handle when the pull request comments are created.
		if *event.Action != "created" {
			response.Message = "Only handle when pull request comment is created"
			return response, nil
		}

		if scmTrigger.PullRequestComment == nil {
			response.Message = "Pull request comment trigger is not enabled"
			return response, nil
		}

		comment := event.Comment
		match := false
		if comment != nil {
			for _, c := range scmTrigger.PullRequestComment.Comments {
				if *comment.Body == c {
					match = true
					break
				}
			}
		}

		if match {
			trigger = api.TriggerWebhookPullRequestComment
			performParams = &api.PipelinePerformParams{
				Ref:         fmt.Sprintf(githubPullRefTemplate, *event.Issue.Number),
				Description: "Triggered by pull request comments",
				Stages:      scmTrigger.PullRequestComment.Stages,
			}
			log.Info("Triggered by Github pull request review comment event")
		}
	case *github.PushEvent:
		if scmTrigger.Push == nil {
			response.Message = "Push trigger is not enabled"
			return response, nil
		}

		ref := *event.Ref
		match := false
		for _, b := range scmTrigger.Push.Branches {
			if strings.HasSuffix(ref, b) {
				match = true
				break
			}
		}

		if match {
			trigger = api.TriggerWebhookPush
			performParams = &api.PipelinePerformParams{
				Ref:         ref,
				Description: "Triggered by push",
				Stages:      scmTrigger.Push.Stages,
			}

			log.Info("Triggered by Github push event")
		}
	default:
		log.Errorf("event type not support.")

	}

	if performParams != nil {
		pipelineRecord := &api.PipelineRecord{
			Name:            performParams.Name,
			PipelineID:      pipeline.ID,
			PerformParams:   performParams,
			Trigger:         trigger,
			PRLastCommitSHA: commitSHA,
		}
		if _, err = pipelineRecordManager.CreatePipelineRecord(pipelineRecord); err != nil {
			return response, err
		}

		response.Message = "Successfully triggered"
		return response, nil
	} else {
		response.Message = "Is ignored"
		return response, nil
	}
}

type GitHubStatusURL struct {
}

// input   : `https://api.github.com/repos/aaa/bbb/statuses/ccc`
// output  : ccc
func extractCommitSha(url string) (string, error) {
	results := githubStatusURLRegexp.FindStringSubmatch(url)
	if len(results) < 2 {
		return "", fmt.Errorf("statusesURL is invalid")
	}
	return results[1], nil
}

// HandleGitlabWebhook handles the webhook request from Gitlab.
// 1. Parse the pipeline id from request path;
// 2. Get the pipeline by id;
// 3. Parse the payload from request header;
// 4. Parse the payload from request body;
// 5. First get the event type, and handle it according to its type.
func HandleGitlabWebhook(ctx context.Context, pipelineID string) (webhookResponse, error) {
	response := webhookResponse{}
	pipeline, err := pipelineManager.GetPipelineByID(pipelineID)
	if err != nil {
		return response, err
	}

	if pipeline.AutoTrigger == nil || pipeline.AutoTrigger.SCMTrigger == nil {
		response.Message = "SCM auto trigger is not set"
		return response, nil
	}
	scmTrigger := pipeline.AutoTrigger.SCMTrigger

	// TODO (robin) Validate the payload.
	// Ref: https://github.com/google/go-github/blob/df47db1628185875602e66d3356ae7337b52bba3/github/messages.go#L213-L233
	// github.ValidatePayload(contextutil.GetHttpRequest(ctx), secretKey)
	request := contextutil.GetHttpRequest(ctx)
	event, err := gitlabuitl.ParseWebHook(request)
	if err != nil {
		return response, err
	}
	log.Infof("Gitlab webhook event: %v", event)

	// Handle the event.
	var performParams *api.PipelinePerformParams
	var commitSHA string
	trigger := api.TriggerSCM

	switch event := event.(type) {
	case *gitlab.TagEvent:
		if scmTrigger.TagRelease == nil {
			response.Message = "Release trigger is not enabled"
			return response, nil
		}

		trigger = api.TriggerWebhookTagRelease
		performParams = &api.PipelinePerformParams{
			Name:        strings.Split(event.Ref, "/")[2],
			Ref:         event.Ref,
			Description: "Triggered by tag release",
			Stages:      scmTrigger.TagRelease.Stages,
		}

		log.Info("Triggered by Gitlab tag event")
	case *gitlab.MergeEvent:
		// Only handle when the pull request are created.
		objectAttributes := event.ObjectAttributes
		if objectAttributes.Action != "open" && objectAttributes.Action != "update" {
			response.Message = "Only handle when merge request is created or updated"
			return response, nil
		}

		if scmTrigger.PullRequest == nil {
			response.Message = "Pull request trigger is not enabled"
			return response, nil
		}

		commitSHA = objectAttributes.LastCommit.ID
		trigger = api.TriggerWebhookPullRequest
		performParams = &api.PipelinePerformParams{
			Ref:         fmt.Sprintf(gitlabMergeRefTemplate, objectAttributes.Iid, objectAttributes.TargetBranch),
			Description: objectAttributes.Title,
			Stages:      scmTrigger.PullRequest.Stages,
		}

		log.Info("Triggered by Gitlab merge event")
	case *gitlabuitl.MergeCommentEvent:
		if event.MergeRequest == nil {
			log.Infof("Only handle comments on merge request")
			response.Message = "Only handle comments on merge request"
			return response, nil
		}

		if scmTrigger.PullRequestComment == nil {
			response.Message = "Pull request comment trigger is not enabled"
			return response, nil
		}
		objectAttributes := event.ObjectAttributes
		match := false
		if objectAttributes.Note != "" {
			for _, c := range scmTrigger.PullRequestComment.Comments {
				if objectAttributes.Note == c {
					match = true
					break
				}
			}
		}

		if match {
			trigger = api.TriggerWebhookPullRequestComment
			performParams = &api.PipelinePerformParams{
				Ref:         fmt.Sprintf(gitlabMergeRefTemplate, event.MergeRequest.IID, event.MergeRequest.TargetBranch),
				Description: "Triggered by pull request comments",
				Stages:      scmTrigger.PullRequestComment.Stages,
			}
			log.Info("Triggered by Gitlab merge comment event")
		}
	case *gitlab.PushEvent:
		if scmTrigger.Push == nil {
			response.Message = "Push trigger is not enabled"
			return response, nil
		}

		ref := event.Ref
		match := false
		for _, b := range scmTrigger.Push.Branches {
			if strings.HasSuffix(ref, b) {
				match = true
				break
			}
		}

		if match {
			trigger = api.TriggerWebhookPush
			performParams = &api.PipelinePerformParams{
				Ref:         ref,
				Description: "Triggered by push",
				Stages:      scmTrigger.Push.Stages,
			}

			log.Info("Triggered by Gitlab push event")
		}
	default:
		log.Errorf("event type not support.")
	}

	if performParams != nil {
		pipelineRecord := &api.PipelineRecord{
			Name:            performParams.Name,
			PipelineID:      pipeline.ID,
			PerformParams:   performParams,
			Trigger:         trigger,
			PRLastCommitSHA: commitSHA,
		}
		if _, err = pipelineRecordManager.CreatePipelineRecord(pipelineRecord); err != nil {
			return response, err
		}
		response.Message = "Successfully triggered"
		return response, nil
	} else {
		response.Message = "Is ignored"
		return response, nil
	}
}

type webhookResponse struct {
	Message string `json:"message,omitempty"`
}
