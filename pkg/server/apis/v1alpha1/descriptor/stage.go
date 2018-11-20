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
	register(stage...)
}

var stage = []definition.Descriptor{
	{
		Path:        "/stages",
		Description: "Stage APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.CreateStage,
				Description: "Create stages",
				Results:     definition.DataErrorResults("stage"),
			},
		},
	},
	{
		Path:        "/stages/{stage-name}",
		Description: "Stage APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetStage,
				Description: "Get stages",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.StageNamePathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.NamespaceQueryParameter,
					},
				},
				Results: definition.DataErrorResults("stage"),
			},
		},
	},
	{
		Path: "/workflowruns/{workflowrun-name}/stages/{stage-name}/streamlogs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ReceivePipelineRecordLogStream,
				Description: "Used for collecting stage logs",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.WorkflowRunNamePathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.StageNamePathParameterName,
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
}
