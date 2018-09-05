package descriptor

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(webhooks...)
}

var webhooks = []definition.Descriptor{
	{
		Path:        "/pipelines/{pipelineid}/githubwebhook",
		Description: "Cloud API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.HandleGithubWebhook,
				Description: "Trigger the pipeline by github webhook",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.PipelineIDPathParameterName,
					},
				},
				Results: definition.DataErrorResults("message"),
			},
		},
	},
	{
		Path:        "/pipelines/{pipelineid}/gitlabwebhook",
		Description: "Cloud API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.HandleGitlabWebhook,
				Description: "Trigger the pipeline by gitlab webhook",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.PipelineIDPathParameterName,
					},
				},
				Results: definition.DataErrorResults("message"),
			},
		},
	},
}
