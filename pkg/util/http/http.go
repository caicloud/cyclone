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
	ResourceNamePathParameterName = "resource"

	// StageNamePathParameterName represents the name of the path parameter for stage name.
	StageNamePathParameterName = "stage"

	// WorkflowNamePathParameterName represents the name of the path parameter for workflow name.
	WorkflowNamePathParameterName = "workflow"

	// WorkflowRunNamePathParameterName represents the name of the path parameter for workflowrun name.
	WorkflowRunNamePathParameterName = "workflowrun"

	// WorkflowTriggerNamePathParameterName represents the name of the path parameter for workflowtrigger name.
	WorkflowTriggerNamePathParameterName = "workflowtrigger"

	// ContainerNameQueryParameter represents the query param container name.
	ContainerNameQueryParameter = "container"

	// NamespaceQueryParameter represents the query param namespace.
	NamespaceQueryParameter string = "namespace"

	// HeaderContentType represents the the key of Content-Type.
	HeaderContentType = "Content-Type"

	// DefaultNamespace represents the default namespace 'default'.
	DefaultNamespace string = "default"

	// DownloadQueryParameter represents a download flag of the query parameter.
	DownloadQueryParameter = "download"

	// StatusQueryParameter represents a status of the query parameter.
	StatusQueryParameter = "status"
)

// GetHTTPRequest gets request from context.
func GetHTTPRequest(ctx context.Context) *http.Request {
	return service.HTTPContextFrom(ctx).Request()
}
