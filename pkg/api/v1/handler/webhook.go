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

// HandleGithubWebhook handles the webhook request from Github.
// 1. Parse the pipeline id from request path;
// 2. Get the pipeline by id;
// 3. Parse the payload from request header;
// 4. Parse the payload from request body;
// 5. First get the event type, and handle it according to its type.
func HandleGithubWebhook(ctx context.Context, pipelineID string) (string, error) {
	pipeline, err := pipelineManager.GetPipelineByID(pipelineID)
	if err != nil {
		return "", err
	}

	if pipeline.AutoTrigger == nil || pipeline.AutoTrigger.SCMTrigger == nil {
		return "SCM auto trigger is not set", nil
	}
	scmTrigger := pipeline.AutoTrigger.SCMTrigger

	// TODO (robin) Validate the payload.
	// Ref: https://github.com/google/go-github/blob/df47db1628185875602e66d3356ae7337b52bba3/github/messages.go#L213-L233
	//github.ValidatePayload(contextutil.GetHttpRequest(ctx), secretKey)
	request := contextutil.GetHttpRequest(ctx)
	payload, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return "", fmt.Errorf("Fail to read the request body")
	}
	event, err := github.ParseWebHook(github.WebHookType(request), payload)
	if err != nil {
		return "", err
	}
	log.Infof("Github webhook event: %v", event)

	// Handle the event.
	var performParams *api.PipelinePerformParams
	switch event := event.(type) {
	case *github.ReleaseEvent:
		if scmTrigger.TagRelease == nil {
			return "Release trigger is not enabled", nil
		}

		performParams = &api.PipelinePerformParams{
			Name:        *event.Release.TagName,
			Ref:         fmt.Sprintf(tagRefTemplate, *event.Release.TagName),
			Description: "Triggered by tag release",
			Stages:      scmTrigger.TagRelease.Stages,
		}

		log.Info("Triggered by Github release event")
	case *github.PullRequestEvent:
		// Only handle when the pull request are created.
		if *event.Action != "opened" {
			return "Only handle when pull request is created", nil
		}

		if scmTrigger.PullRequest == nil {
			return "Pull request trigger is not enabled", nil
		}

		performParams = &api.PipelinePerformParams{
			Ref:         fmt.Sprintf(githubPullRefTemplate, *event.PullRequest.Number),
			Description: "Triggered by pull request",
			Stages:      scmTrigger.PullRequest.Stages,
		}

		log.Info("Triggered by Github pull request event")
	case *github.IssueCommentEvent:
		if event.Issue.PullRequestLinks == nil {
			log.Infof("Only handle when issues type is pull request")
			return "Only handle when issues type is pull request", nil
		}

		// Only handle when the pull request comments are created.
		if *event.Action != "created" {
			return "Only handle when pull request comment is created", nil
		}

		if scmTrigger.PullRequestComment == nil {
			return "Pull request comment trigger is not enabled", nil
		}

		comment := event.Comment
		trigger := false
		if comment != nil {
			for _, c := range scmTrigger.PullRequestComment.Comments {
				if *comment.Body == c {
					trigger = true
					break
				}
			}
		}

		if trigger {
			performParams = &api.PipelinePerformParams{
				Ref:         fmt.Sprintf(githubPullRefTemplate, *event.Issue.Number),
				Description: "Triggered by pull request comments",
				Stages:      scmTrigger.PullRequestComment.Stages,
			}
			log.Info("Triggered by Github pull request review comment event")
		}
	case *github.PushEvent:
		if scmTrigger.Push == nil {
			return "Push trigger is not enabled", nil
		}

		ref := *event.Ref
		trigger := false
		for _, b := range scmTrigger.Push.Branches {
			if strings.HasSuffix(ref, b) {
				trigger = true
				break
			}
		}

		if trigger {
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
			Name:          performParams.Name,
			PipelineID:    pipeline.ID,
			PerformParams: performParams,
			Trigger:       api.TriggerSCM,
		}
		if _, err = pipelineRecordManager.CreatePipelineRecord(pipelineRecord); err != nil {
			return "", err
		}

		return "Successfully triggered", nil
	} else {
		return "Is ignored", nil
	}
}

// HandleGitlabWebhook handles the webhook request from Gitlab.
// 1. Parse the pipeline id from request path;
// 2. Get the pipeline by id;
// 3. Parse the payload from request header;
// 4. Parse the payload from request body;
// 5. First get the event type, and handle it according to its type.
func HandleGitlabWebhook(ctx context.Context, pipelineID string) (string, error) {
	pipeline, err := pipelineManager.GetPipelineByID(pipelineID)
	if err != nil {
		return "", err
	}

	if pipeline.AutoTrigger == nil || pipeline.AutoTrigger.SCMTrigger == nil {
		return "SCM auto trigger is not set", nil
	}
	scmTrigger := pipeline.AutoTrigger.SCMTrigger

	// TODO (robin) Validate the payload.
	// Ref: https://github.com/google/go-github/blob/df47db1628185875602e66d3356ae7337b52bba3/github/messages.go#L213-L233
	// github.ValidatePayload(contextutil.GetHttpRequest(ctx), secretKey)
	request := contextutil.GetHttpRequest(ctx)
	event, err := gitlabuitl.ParseWebHook(request)
	if err != nil {
		return "", err
	}
	log.Infof("Gitlab webhook event: %v", event)

	// Handle the event.
	var performParams *api.PipelinePerformParams
	switch event := event.(type) {
	case *gitlab.TagEvent:
		if scmTrigger.TagRelease == nil {
			return "Release trigger is not enabled", nil
		}

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
		if objectAttributes.Action != "open" {
			return "Only handle when merge request is created", nil
		}

		if scmTrigger.PullRequest == nil {
			return "Pull request trigger is not enabled", nil
		}

		performParams = &api.PipelinePerformParams{
			Ref:         fmt.Sprintf(gitlabMergeRefTemplate, objectAttributes.Iid, objectAttributes.TargetBranch),
			Description: objectAttributes.Title,
			Stages:      scmTrigger.PullRequest.Stages,
		}

		log.Info("Triggered by Gitlab merge event")
	case *gitlabuitl.MergeCommentEvent:
		if event.MergeRequest == nil {
			log.Infof("Only handle comments on merge request")
			return "Only handle comments on merge request", nil
		}

		if scmTrigger.PullRequestComment == nil {
			return "Pull request comment trigger is not enabled", nil
		}
		objectAttributes := event.ObjectAttributes
		trigger := false
		if objectAttributes.Note != "" {
			for _, c := range scmTrigger.PullRequestComment.Comments {
				if objectAttributes.Note == c {
					trigger = true
					break
				}
			}
		}

		if trigger {
			performParams = &api.PipelinePerformParams{
				Ref:         fmt.Sprintf(gitlabMergeRefTemplate, event.MergeRequest.IID, event.MergeRequest.TargetBranch),
				Description: "Triggered by pull request comments",
				Stages:      scmTrigger.PullRequestComment.Stages,
			}
			log.Info("Triggered by Gitlab merge comment event")
		}
	case *gitlab.PushEvent:
		if scmTrigger.Push == nil {
			return "Push trigger is not enabled", nil
		}

		ref := event.Ref
		trigger := false
		for _, b := range scmTrigger.Push.Branches {
			if strings.HasSuffix(ref, b) {
				trigger = true
				break
			}
		}

		if trigger {
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
			Name:          performParams.Name,
			PipelineID:    pipeline.ID,
			PerformParams: performParams,
			Trigger:       api.TriggerSCM,
		}
		if _, err = pipelineRecordManager.CreatePipelineRecord(pipelineRecord); err != nil {
			return "", err
		}

		return "Successfully triggered", nil
	} else {
		return "Is ignored", nil
	}
}
