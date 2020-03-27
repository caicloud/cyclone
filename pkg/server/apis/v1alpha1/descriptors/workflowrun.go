/*
Copyright 2018 caicloud authors. All rights reserved.

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

package descriptors

import (
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/operators/validator"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(workflowrun...)
}

var workflowrun = []definition.Descriptor{
	{
		Path:        "/projects/{project}/workflows/{workflow}/workflowruns",
		Description: "workflowrun APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.CreateWorkflowRun,
				Description: "Create workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the new workflowrun",
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
			{
				Method:      definition.Get,
				Function:    handler.ListWorkflowRuns,
				Description: "List workflowruns",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Auto,
						Name:        httputil.PaginationAutoParameter,
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
	{
		Path:        "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}",
		Description: "workflowrun APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetWorkflowRun,
				Description: "Get workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateWorkflowRun,
				Description: "Update workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the updated workflowrun",
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteWorkflowRun,
				Description: "Delete workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path: "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/stop",
		Definitions: []definition.Definition{
			{
				Method:      definition.Update,
				Function:    handler.StopWorkflowRun,
				Description: "Stop a workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
	{
		Path: "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/pause",
		Definitions: []definition.Definition{
			{
				Method:      definition.Update,
				Function:    handler.PauseWorkflowRun,
				Description: "Pause a workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
	{
		Path: "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/resume",
		Definitions: []definition.Definition{
			{
				Method:      definition.Update,
				Function:    handler.ResumeWorkflowRun,
				Description: "Continue ro run WorkflowRun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
	{
		Path: "/workflowruns/{workflowrun}/streamlogs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ReceiveContainerLogStream,
				Description: "Used for collecting stage logs",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.NamespaceQueryParameter,
					},
					{
						Source: definition.Query,
						Name:   httputil.StageNameQueryParameter,
					},
					{
						Source: definition.Query,
						Name:   httputil.ContainerNameQueryParameter,
					},
				},
				Results: []definition.Result{
					{
						Destination: definition.Error,
					},
				},
			},
		},
	},
	{
		Path: "/workflowruns/{workflowrun}/artifacts",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.ReceiveArtifacts,
				Consumes:    []string{definition.MIMEFormData},
				Description: "Collect stage artifacts",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.NamespaceQueryParameter,
					},
					{
						Source:    definition.Query,
						Name:      httputil.StageNameQueryParameter,
						Operators: []definition.Operator{validator.String("required")},
					},
				},
				Results: []definition.Result{
					{
						Destination: definition.Error,
					},
				},
			},
		},
	},
	{
		Path: "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/logstream",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetContainerLogStream,
				Description: "Get log stream of a stage",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Query,
						Name:   httputil.StageNameQueryParameter,
					},
					{
						Source: definition.Query,
						Name:   "tenant",
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path: "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/logs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Produces:    []string{definition.MIMEText},
				Function:    handler.GetContainerLogs,
				Description: "Get log of containers",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Query,
						Name:   httputil.StageNameQueryParameter,
					},
					{
						Source: definition.Query,
						Name:   httputil.ContainerNameQueryParameter,
					},
					{
						Source:    definition.Query,
						Name:      httputil.DownloadQueryParameter,
						Operators: []definition.Operator{validator.Bool("")},
					},
				},
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Description: "container log",
					},
					{
						Destination: definition.Meta,
					},
					{
						Destination: definition.Error,
					},
				},
			},
		},
	},
	{
		Path: "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/artifacts",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListArtifacts,
				Description: "List artifacts produced in the workflowRun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
	{
		Path: "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/artifacts/{artifact}",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Produces:    []string{definition.MIMEOctetStream},
				Function:    handler.DownloadArtifact,
				Description: "download artifact",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ArtifactNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Query,
						Name:   httputil.StageNameQueryParameter,
					},
				},
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Description: "artifact",
					},
					{
						Destination: definition.Meta,
					},
					{
						Destination: definition.Error,
					},
				},
			},
		},
	},
	{
		Path: "/projects/{project}/workflows/{workflow}/workflowruns/{workflowrun}/artifacts/{artifact}",
		Definitions: []definition.Definition{
			{
				Method:      definition.Delete,
				Function:    handler.DeleteArtifact,
				Description: "delete artifact",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ArtifactNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Query,
						Name:   httputil.StageNameQueryParameter,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
