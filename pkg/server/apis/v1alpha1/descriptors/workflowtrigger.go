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

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(workflowtrigger...)
}

var workflowtrigger = []definition.Descriptor{
	{
		Path:        "/projects/{project}/workflowtriggers",
		Description: "workflowtrigger APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.CreateWorkflowTrigger,
				Description: "Create workflowtrigger",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the new workflowtrigger",
					},
				},
				Results: definition.DataErrorResults("workflowtrigger"),
			},
			{
				Method:      definition.Get,
				Function:    handler.ListWorkflowTriggers,
				Description: "List workflowtriggers",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Auto,
						Name:        "pagination",
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("workflowtrigger"),
			},
		},
	},
	{
		Path:        "/projects/{project}/workflowtriggers/{workflowtrigger}",
		Description: "workflowtrigger APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetWorkflowTrigger,
				Description: "Get workflowtrigger",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowTriggerNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
				},
				Results: definition.DataErrorResults("workflowtrigger"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateWorkflowTrigger,
				Description: "Update workflowtrigger",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowTriggerNamePathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the updated workflowtrigger",
					},
				},
				Results: definition.DataErrorResults("workflowtrigger"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteWorkflowTrigger,
				Description: "Delete workflowtrigger",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.WorkflowTriggerNamePathParameterName,
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
}
