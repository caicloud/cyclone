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
	register(workflowtrigger...)
}

var workflowtrigger = []definition.Descriptor{
	{
		Path:        "/workflowtriggers",
		Description: "workflowtrigger APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.CreateWorkflowTrigger,
				Description: "Create workflowtrigger",
				Results:     definition.DataErrorResults("workflowtrigger"),
			},
		},
	},
	{
		Path:        "/workflowtriggers{workflowtrigger-name}",
		Description: "workflowtrigger APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetWorkflowTrigger,
				Description: "Get workflowtrigger",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.WorkflowTriggerNamePathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.NamespaceQueryParameter,
					},
				},
				Results: definition.DataErrorResults("workflowtrigger"),
			},
		},
	},
}
