/*
Copyright 2018 caicloud authors. All rights reserved.
*/

package descriptor

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1/handler"
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
		},
	},
	{
		Path:        "/workflowruns{workflowrun-name}",
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
						Source: definition.Query,
						Name:   httputil.NamespaceQueryParameter,
					},
				},
				Results: definition.DataErrorResults("workflowrun"),
			},
		},
	},
}
