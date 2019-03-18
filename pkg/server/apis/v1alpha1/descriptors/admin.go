/*
Copyright 2019 caicloud authors. All rights reserved.

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
	register(all...)
}

var all = []definition.Descriptor{
	{
		Path:        "/workflows",
		Description: "APIs to list all workflows",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.AllWorkflows,
				Description: "List all workflows regardless of tenant and project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Query,
						Name:        httputil.LabelQueryParameter,
						Description: "label to filter workflows",
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
		Path:        "/stages",
		Description: "APIs to list all stages",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.AllStages,
				Description: "List all stages regardless of tenant and project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Query,
						Name:        httputil.LabelQueryParameter,
						Description: "label to filter stages",
					},
					{
						Source:      definition.Auto,
						Name:        httputil.PaginationAutoParameter,
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("stages"),
			},
		},
	},
	{
		Path:        "/resources",
		Description: "APIs to list all resources",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.AllResources,
				Description: "List all resources regardless of tenant and project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Query,
						Name:        httputil.LabelQueryParameter,
						Description: "label to filter resources",
					},
					{
						Source:      definition.Auto,
						Name:        httputil.PaginationAutoParameter,
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("resources"),
			},
		},
	},
	{
		Path:        "/workflowruns",
		Description: "APIs to list all workflowruns",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.AllWorkflowRuns,
				Description: "List all workflowruns regardless of tenant and project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Query,
						Name:        httputil.LabelQueryParameter,
						Description: "label to filter workflowruns",
					},
					{
						Source:      definition.Auto,
						Name:        httputil.PaginationAutoParameter,
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("workflowruns"),
			},
		},
	},
	{
		Path:        "/workflowtriggers",
		Description: "APIs to list all workflowtriggers",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.AllWorkflowTriggers,
				Description: "List all workflowtriggers regardless of tenant and project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Query,
						Name:        httputil.LabelQueryParameter,
						Description: "label to filter workflowtriggers",
					},
					{
						Source:      definition.Auto,
						Name:        httputil.PaginationAutoParameter,
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("workflowtriggers"),
			},
		},
	},
}
