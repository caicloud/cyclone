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

	"github.com/caicloud/cyclone/pkg/server/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(workflowrun...)
}

var workflowrun = []definition.Descriptor{
	{
		Path:        "/workflowruns",
		Description: "workflowrun APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.CreateWorkflowRun,
				Description: "Create workflowrun",
				Results:     definition.DataErrorResults("workflowrun"),
			},
			{
				Method:      definition.Get,
				Function:    handler.ListWorkflowRuns,
				Description: "List workflowruns",
				Parameters: []definition.Parameter{
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
	{
		Path:        "/workflowruns/{workflowrun}",
		Description: "workflowrun APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetWorkflowRun,
				Description: "Get workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
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
						Name:   httputil.WorkflowRunNamePathParameterName,
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
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path: "/workflowruns/{workflowrun}/cancel",
		Definitions: []definition.Definition{
			{
				Method:      definition.Update,
				Function:    handler.CancelWorkflowRun,
				Description: "Cancel workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
	{
		Path: "/workflowruns/{workflowrun}/continue",
		Definitions: []definition.Definition{
			{
				Method:      definition.Update,
				Function:    handler.ContinueWorkflowRun,
				Description: "Continue ro run workflowrun",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
}
