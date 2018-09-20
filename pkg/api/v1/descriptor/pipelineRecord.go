package descriptor

import (
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/operators/validator"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(records...)
}

var records = []definition.Descriptor{
	{
		Path:        "/projects/{project}/pipelines/{pipeline}/records",
		Description: "Pipeline record API",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListPipelineRecords,
				Description: "Get all pipeline records of one pipeline",
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
				Results: definition.DataErrorResults("pipeline records"),
			},
			{
				Method:      definition.Create,
				Function:    handler.CreatePipelineRecord,
				Description: "Perform pipeline, which will create a pipeline record",
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
						Source: definition.Header,
						Name:   httputil.HeaderUser,
					},
				},
				Results: definition.DataErrorResults("pipeline record"),
			},
		},
	},
	{
		Path:        "/projects/{project}/pipelines/{pipeline}/records/{recordid}",
		Description: "Pipeline record API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetPipelineRecord,
				Description: "Get the pipeline record",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
				},
				Results: definition.DataErrorResults("pipeline record"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeletePipelineRecord,
				Description: "Delete a pipeline record",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path:        "/projects/{project}/pipelines/{pipeline}/records/{recordid}/status",
		Description: "Pipeline record status API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Patch,
				Function:    handler.UpdatePipelineRecordStatus,
				Description: "Update the status of pipeline record, only support to set the status as Aborted for running pipeline record",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
				},
				Results: definition.DataErrorResults("pipeline record"),
			},
		},
	},
	{
		Path:        "/projects/{project}/pipelines/{pipeline}/records/{recordid}/logs",
		Description: "Pipeline record log API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetPipelineRecordLogs,
				Description: "Get the pipeline record log",
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
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.PipelineRecordStageQueryParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.PipelineRecordTaskQueryParameterName,
					},
					{
						Source:    definition.Query,
						Name:      httputil.PipelineRecordDownloadQueryParameter,
						Operators: []definition.Operator{validator.Bool("")},
						//Default:   false,
					},
				},
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Description: "pipeline record log",
					},
					{
						Destination: definition.Meta,
					},
					{
						Destination: definition.Error,
					},
				},
			},
		},
	},
	{
		Path:        "/projects/{project}/pipelines/{pipeline}/records/{recordid}/testresults",
		Description: "Pipeline record test results API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListPipelineRecordTestResults,
				Description: "Get the pipeline record test results",
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
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
				},
				Results: definition.DataErrorResults("pipeline record test results"),
			},
			{
				Method:      definition.Create,
				Function:    handler.ReceivePipelineRecordTestResult,
				Description: "Receive test result of pipeline record from worker",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path:        "/projects/{project}/pipelines/{pipeline}/records/{recordid}/testresults/{filename}",
		Description: "Pipeline record test result API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetPipelineRecordTestResult,
				Description: "Get the pipeline record test result",
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
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
					{
						Source: definition.Path,
						Name:   httputil.FileNamePathParameterName,
					},
					{
						Source:    definition.Query,
						Name:      httputil.PipelineRecordDownloadQueryParameter,
						Operators: []definition.Operator{validator.Bool("")},
						//Default:   false,
					},
				},
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Description: "pipeline record test result",
					},
					{
						Destination: definition.Meta,
					},
					{
						Destination: definition.Error,
					},
				},
			},
		},
	},
	{
		Path:        "/projects/{project}/pipelines/{pipeline}/records/{recordid}/logstream",
		Description: "Pipeline record log API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetPipelineRecordLogStream,
				Description: "Get log stream of pipeline record",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.PipelineRecordStageQueryParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.PipelineRecordTaskQueryParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		// TODO gorilla/websocket only supports GET method. This a workaround as this API is only used by workers,
		// but still need a better way.
		Path:        "/projects/{project}/pipelines/{pipeline}/records/{recordid}/stagelogstream",
		Description: "Pipeline record log API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ReceivePipelineRecordLogStream,
				Description: "Receive log stream of pipeline record from worker",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.PipelineRecordPathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.PipelineRecordStageQueryParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.PipelineRecordTaskQueryParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
