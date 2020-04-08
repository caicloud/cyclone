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
	register(workflow...)
}

var workflow = []definition.Descriptor{
	{
		Path:        "/projects/{project}/workflows",
		Description: "workflow APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.CreateWorkflow,
				Description: "Create workflow",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the new workflow",
					},
				},
				Results: definition.DataErrorResults("workflow"),
			},
			{
				Method:      definition.List,
				Function:    handler.ListWorkflows,
				Description: "List workflows",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source:      definition.Auto,
						Name:        httputil.PaginationAutoParameter,
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("workflows"),
			},
		},
	},
	{
		Path:        "/projects/{project}/workflows/{workflow}",
		Description: "workflow APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetWorkflow,
				Description: "Get workflow",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
				},
				Results: definition.DataErrorResults("workflow"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateWorkflow,
				Description: "Update workflow",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the updated workflow",
					},
				},
				Results: definition.DataErrorResults("workflow"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteWorkflow,
				Description: "Delete workflow",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
		Children: []definition.Descriptor{
			{
				Path: "/stats",
				Definitions: []definition.Definition{
					{
						Method:      definition.Get,
						Function:    handler.GetWFStatistics,
						Description: "Get statistics of the workflow",
						Parameters: []definition.Parameter{
							{
								Source:      definition.Header,
								Name:        httputil.TenantHeaderName,
								Description: "Name of the tenant whose project to stats",
							},
							{
								Source: definition.Path,
								Name:   httputil.ProjectNamePathParameterName,
							},
							{
								Source: definition.Path,
								Name:   httputil.WorkflowNamePathParameterName,
							},
							{
								Source:    definition.Query,
								Name:      httputil.StartTimeQueryParameter,
								Operators: []definition.Operator{validator.String("required")},
							},
							{
								Source:    definition.Query,
								Name:      httputil.EndTimeQueryParameter,
								Operators: []definition.Operator{validator.String("required")},
							},
						},
						Results: definition.DataErrorResults("workflow stats"),
					},
				},
			},
		},
	},
}
