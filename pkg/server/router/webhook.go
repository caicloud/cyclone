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

package router

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	restful "github.com/emicklei/go-restful"
	log "github.com/golang/glog"
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"

	"github.com/caicloud/cyclone/pkg/api"
	gitlabuitl "github.com/caicloud/cyclone/pkg/util/gitlab"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
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

// handleGithubWebhook handles the webhook request from Github.
// 1. Parse the pipeline id from request path;
// 2. Get the pipeline by id;
// 3. Parse the payload from request header;
// 4. Parse the payload from request body;
// 5. First get the event type, and handle it according to its type.
func (router *router) handleGithubWebhook(request *restful.Request, response *restful.Response) {
	pipelineID := request.PathParameter(pipelineIDPathParameterName)
	pipeline, err := router.pipelineManager.GetPipelineByID(pipelineID)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	if pipeline.AutoTrigger == nil || pipeline.AutoTrigger.SCMTrigger == nil {
		httputil.ResponseWithError(response, fmt.Errorf("SCM auto trigger is not set"))
		return
	}
	scmTrigger := pipeline.AutoTrigger.SCMTrigger

	// TODO (robin) Validate the payload.
	// Ref: https://github.com/google/go-github/blob/df47db1628185875602e66d3356ae7337b52bba3/github/messages.go#L213-L233
	// github.ValidatePayload(request.Request, secretKey)
	payload, err := ioutil.ReadAll(request.Request.Body)
	if err != nil {
		httputil.ResponseWithError(response, fmt.Errorf("Fail to read the request body"))
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(request.Request), payload)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}
	log.Infof("Github webhook event: %v", event)

	// Handle the event.
	var performParams *api.PipelinePerformParams
	switch event := event.(type) {
	case *github.ReleaseEvent:
		if scmTrigger.TagRelease == nil {
			response.WriteHeaderAndEntity(http.StatusOK, "Release trigger is not enabled")
			return
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
			response.WriteHeaderAndEntity(http.StatusOK, "Only handle when pull request is created")
			return
		}

		if scmTrigger.PullRequest == nil {
			response.WriteHeaderAndEntity(http.StatusOK, "Pull request trigger is not enabled")
			return
		}

		performParams = &api.PipelinePerformParams{
			Ref:         fmt.Sprintf(githubPullRefTemplate, *event.PullRequest.Number),
			Description: "Triggered by pull request",
			Stages:      scmTrigger.PullRequest.Stages,
		}
		log.Info("Triggered by Github pull request event")
	case *github.PullRequestReviewCommentEvent:
		// Only handle when the pull request comments are created.
		if *event.Action != "created" {
			response.WriteHeaderAndEntity(http.StatusOK, "Only handle when pull request comment is created")
			return
		}

		if scmTrigger.PullRequestComment == nil {
			response.WriteHeaderAndEntity(http.StatusOK, "Pull request comment trigger is not enabled")
			return
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
				Ref:         fmt.Sprintf(githubPullRefTemplate, *event.PullRequest.Number),
				Description: "Triggered by pull request comments",
				Stages:      scmTrigger.PullRequestComment.Stages,
			}
			log.Info("Triggered by Github pull request review comment event")
		}
	}

	if performParams != nil {
		pipelineRecord := &api.PipelineRecord{
			Name:          performParams.Name,
			PipelineID:    pipeline.ID,
			PerformParams: performParams,
			Trigger:       api.TriggerSCM,
		}
		if _, err = router.pipelineRecordManager.CreatePipelineRecord(pipelineRecord); err != nil {
			httputil.ResponseWithError(response, err)
			return
		}

		response.WriteHeaderAndEntity(http.StatusOK, "Successfully triggered")
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, "Is ignored")
	}
}

// handleGitlabWebhook handles the webhook request from Gitlab.
// 1. Parse the pipeline id from request path;
// 2. Get the pipeline by id;
// 3. Parse the payload from request header;
// 4. Parse the payload from request body;
// 5. First get the event type, and handle it according to its type.
func (router *router) handleGitlabWebhook(request *restful.Request, response *restful.Response) {
	pipelineID := request.PathParameter(pipelineIDPathParameterName)
	pipeline, err := router.pipelineManager.GetPipelineByID(pipelineID)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	if pipeline.AutoTrigger == nil || pipeline.AutoTrigger.SCMTrigger == nil {
		httputil.ResponseWithError(response, fmt.Errorf("SCM auto trigger is not set"))
		return
	}
	scmTrigger := pipeline.AutoTrigger.SCMTrigger

	// TODO (robin) Validate the payload.
	// Ref: https://github.com/google/go-github/blob/df47db1628185875602e66d3356ae7337b52bba3/github/messages.go#L213-L233
	// github.ValidatePayload(request.Request, secretKey)
	event, err := gitlabuitl.ParseWebHook(request.Request)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}
	log.Infof("Gitlab webhook event: %v", event)

	// Handle the event.
	var performParams *api.PipelinePerformParams
	switch event := event.(type) {
	case *gitlab.TagEvent:
		if scmTrigger.TagRelease == nil {
			response.WriteHeaderAndEntity(http.StatusOK, "Release trigger is not enabled")
			return
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
			response.WriteHeaderAndEntity(http.StatusOK, "Only handle when merge request is created")
			return
		}

		if scmTrigger.PullRequest == nil {
			response.WriteHeaderAndEntity(http.StatusOK, "Pull request trigger is not enabled")
			return
		}

		performParams = &api.PipelinePerformParams{
			Ref:         fmt.Sprintf(gitlabMergeRefTemplate, objectAttributes.Iid, objectAttributes.TargetBranch),
			Description: objectAttributes.Title,
			Stages:      scmTrigger.PullRequest.Stages,
		}

		log.Info("Triggered by Gitlab merge event")
	case *gitlab.MergeCommentEvent:
		if scmTrigger.PullRequestComment == nil {
			response.WriteHeaderAndEntity(http.StatusOK, "Pull request comment trigger is not enabled")
			return
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
	}

	if performParams != nil {
		pipelineRecord := &api.PipelineRecord{
			Name:          performParams.Name,
			PipelineID:    pipeline.ID,
			PerformParams: performParams,
			Trigger:       api.TriggerSCM,
		}
		if _, err = router.pipelineRecordManager.CreatePipelineRecord(pipelineRecord); err != nil {
			httputil.ResponseWithError(response, err)
			return
		}

		response.WriteHeaderAndEntity(http.StatusOK, "Successfully triggered")
	} else {
		response.WriteHeaderAndEntity(http.StatusOK, "Is ignored")
	}
}
