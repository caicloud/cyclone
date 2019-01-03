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
	register(workflow...)
}

var workflow = []definition.Descriptor{
	{
		Path:        "/workflows",
		Description: "workflow APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.CreateWorkflow,
				Description: "Create workflow",
				Results:     definition.DataErrorResults("workflow"),
			},
			{
				Method:      definition.Get,
				Function:    handler.ListWorkflows,
				Description: "List workflows",
				Parameters: []definition.Parameter{
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
					},
				},
				Results: definition.DataErrorResults("workflows"),
			},
		},
	},
	{
		Path:        "/workflows/{workflow}",
		Description: "workflow APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetWorkflow,
				Description: "Get workflow",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
					},
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
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
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
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
						Source: definition.Path,
						Name:   httputil.WorkflowNamePathParameterName,
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
}
