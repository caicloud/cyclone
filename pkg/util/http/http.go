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
	"strings"

	"github.com/caicloud/nirvana/service"
)

const (

	// APIVersion is the version of API.
	APIVersion = "/apis/v1alpha1"

	// TenantNamePathParameterName represents the name of the path parameter for tenant name.
	TenantNamePathParameterName = "tenant"

	// ProjectNamePathParameterName represents the name of the path parameter for project name.
	ProjectNamePathParameterName = "project"

	// ResourceNamePathParameterName represents the name of the path parameter for resource name.
	ResourceNamePathParameterName = "resource"

	// StageNamePathParameterName represents the name of the path parameter for stage name.
	StageNamePathParameterName = "stage"

	// TemplateNamePathParameterName represents the name of the path parameter for template name.
	TemplateNamePathParameterName = "template"

	// WorkflowNamePathParameterName represents the name of the path parameter for workflow name.
	WorkflowNamePathParameterName = "workflow"

	// WorkflowRunNamePathParameterName represents the name of the path parameter for workflowrun name.
	WorkflowRunNamePathParameterName = "workflowrun"

	// ArtifactNamePathParameterName represents the name of the path parameter for artifact name.
	ArtifactNamePathParameterName = "artifact"

	// WorkflowTriggerNamePathParameterName represents the name of the path parameter for workflowtrigger name.
	WorkflowTriggerNamePathParameterName = "workflowtrigger"

	// NamespaceQueryParameter represents namespace query parameter.
	NamespaceQueryParameter = "namespace"

	// StageNameQueryParameter represents the query param stage name.
	StageNameQueryParameter = "stage"

	// LabelQueryParameter represents the label query.
	LabelQueryParameter = "label"

	// ContainerNameQueryParameter represents the query param container name.
	ContainerNameQueryParameter = "container"

	// PaginationAutoParameter represents the auto param pagination.
	PaginationAutoParameter = "pagination"

	// TenantHeaderName is name of tenant header name in http request
	TenantHeaderName = "X-Tenant"

	// NamespaceHeaderName is name of namespace header
	NamespaceHeaderName = "X-Namespace"

	// HeaderContentType represents the key of Content-Type.
	HeaderContentType = "Content-Type"

	// HeaderContentTypeJSON represents the JSON Content-Type value.
	HeaderContentTypeJSON = "application/json"

	// PublicHeaderName is a header name to indicate whether the resource is public
	PublicHeaderName = "X-Public"

	// HeaderDryRun is a header name to indicate do a rehearsal of a performance or procedure before the real one.
	HeaderDryRun = "X-Dry-Run"

	// DefaultNamespace represents the default namespace 'default'.
	DefaultNamespace = "default"

	// DownloadQueryParameter represents a download flag of the query parameter.
	DownloadQueryParameter = "download"

	// StatusQueryParameter represents a status of the query parameter.
	StatusQueryParameter = "status"

	// IncludePublicQueryParameter indicates whether include system level resources, for example, when list
	// stage templates in a tenant, whether to include system level templates. Default is true.
	IncludePublicQueryParameter = "includePublic"

	// StartTimeQueryParameter represents the query param start time.
	StartTimeQueryParameter string = "startTime"

	// EndTimeQueryParameter represents the query param end time.
	EndTimeQueryParameter string = "endTime"

	// OperationQueryParameter ...
	OperationQueryParameter string = "operation"

	// ResourceTypePathParameterName ...
	ResourceTypePathParameterName = "resourceType"
)

// GetHTTPRequest gets request from context.
func GetHTTPRequest(ctx context.Context) *http.Request {
	return service.HTTPContextFrom(ctx).Request()
}

// EnsureProtocolScheme ensures URL has protocol sheme set
func EnsureProtocolScheme(url string) string {
	if !strings.Contains(url, "://") {
		return "http://" + url
	}
	return url
}
