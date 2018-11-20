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
		},
	},
	{
		Path:        "/workflows/{workflow-name}",
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
						Source: definition.Query,
						Name:   httputil.NamespaceQueryParameter,
					},
				},
				Results: definition.DataErrorResults("workflow"),
			},
		},
	},
}
