package descriptor

import (
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/operators/validator"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/api/v1/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(pipelines...)
}

var pipelines = []definition.Descriptor{
	{
		Path:        "/projects/{project}/pipelines",
		Description: "Pipeline API",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListPipelines,
				Description: "List all pipelines",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source:    definition.Query,
						Name:      api.RecentPipelineRecordCount,
						Operators: []definition.Operator{validator.Int("")},
					},
					{
						Source:    definition.Query,
						Name:      api.RecentSuccessPipelineRecordCount,
						Operators: []definition.Operator{validator.Int("")},
					},
					{
						Source:    definition.Query,
						Name:      api.RecentFailedPipelineRecordCount,
						Operators: []definition.Operator{validator.Int("")},
					},
				},
				Results: definition.DataErrorResults("pipelines"),
			},
			{
				Method:      definition.Create,
				Function:    handler.CreatePipeline,
				Description: "Add a pipeline",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source: definition.Header,
						Name:   httputil.HeaderUser,
					},
				},
				Results: definition.DataErrorResults("pipeline"),
			},
		},
	},
	{
		Path:        "/projects/{project}/pipelines/{pipeline}",
		Description: "Pipeline API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetPipeline,
				Description: "Get tht pipeline",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.PipelinePathParameterName,
					},
					{
						Source:    definition.Query,
						Name:      api.RecentPipelineRecordCount,
						Operators: []definition.Operator{validator.Int("")},
					},
					{
						Source:    definition.Query,
						Name:      api.RecentSuccessPipelineRecordCount,
						Operators: []definition.Operator{validator.Int("")},
					},
					{
						Source:    definition.Query,
						Name:      api.RecentFailedPipelineRecordCount,
						Operators: []definition.Operator{validator.Int("")},
					},
				},
				Results: definition.DataErrorResults("pipeline"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdatePipeline,
				Description: "Update the pipeline",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.PipelinePathParameterName,
					},
				},
				Results: definition.DataErrorResults("pipeline"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeletePipeline,
				Description: "Delete the project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.PipelinePathParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path:        "/projects/{project}/pipelines/{pipeline}/stats",
		Description: "Pipeline stats API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetPipelineStatistics,
				Description: "Get statistics of the pipeline",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.PipelinePathParameterName,
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
				Results: definition.DataErrorResults("pipeline stats"),
			},
		},
	},
}
