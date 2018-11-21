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

package http

import (
	"context"
	"net/http"

	"github.com/caicloud/nirvana/service"
)

const (

	// APIVersion is the version of API.
	APIVersion = "/apis/v1alpha1"

	// ResourceNamePathParameterName represents the name of the path parameter for resource name.
	ResourceNamePathParameterName = "resource-name"

	// StageNamePathParameterName represents the name of the path parameter for stage name.
	StageNamePathParameterName = "stage-name"

	// WorkflowNamePathParameterName represents the name of the path parameter for workflow name.
	WorkflowNamePathParameterName = "workflow-name"

	// WorkflowRunNamePathParameterName represents the name of the path parameter for workflowrun name.
	WorkflowRunNamePathParameterName = "workflowrun-name"

	// WorkflowTriggerNamePathParameterName represents the name of the path parameter for workflowtrigger name.
	WorkflowTriggerNamePathParameterName = "workflowtrigger-name"

	// ContainerNameQueryParameter represents the query param container name.
	ContainerNameQueryParameter = "container"

	// NamespaceQueryParameter represents the query param namespace.
	NamespaceQueryParameter string = "namespace"

	// ProjectPathParameterName represents the name of the path parameter for project.
	ProjectPathParameterName = "project"

	// PipelinePathParameterName represents the name of the path parameter for pipeline.
	PipelinePathParameterName = "pipeline"

	// PipelineRecordPathParameterName represents the name of the path parameter for pipeline record.
	PipelineRecordPathParameterName = "recordid"

	// FileNamePathParameterName represents the name of the path parameter for file name.
	FileNamePathParameterName = "filename"

	// PipelineRecordStagePathParameterName represents the name of the query parameter for pipeline record stage.
	PipelineRecordStageQueryParameterName = "stage"

	// PipelineRecordTaskQueryParameterName represents the name of the query parameter for pipeline record task.
	PipelineRecordTaskQueryParameterName = "task"

	// PipelineRecordDownloadQueryParameter represents a download flag of the query parameter for pipeline record task.
	PipelineRecordDownloadQueryParameter = "download"

	// EventPathParameterName represents the name of the path parameter for event.
	EventPathParameterName = "eventid"

	// CloudPathParameterName represents the name of the path parameter for cloud.
	CloudPathParameterName = "cloud"

	// RepoQueryParameterName represents the repo name of the query parameter.
	RepoQueryParameter = "repo"

	// EndTimeQueryParameter represents the query param end time.
	EndTimeQueryParameter string = "endTime"

	// HeaderUser represents the the key of user in request header.
	HeaderUser = "X-User"

	HEADER_ContentType = "Content-Type"

	// SVNRepoIDPathParameterName represents a svn repository's uuid.
	SVNRepoIDPathParameterName = "svnrepoid"

	// SVNRevisionQueryParameterName represents the svn commit revision.
	SVNRevisionQueryParameterName = "revision"
)

// GetHttpRequest gets request from context.
func GetHttpRequest(ctx context.Context) *http.Request {
	return service.HTTPContextFrom(ctx).Request()
}
