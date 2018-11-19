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
		Description: "GitHub webhook API",
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
		Description: "GitLab webhook API",
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
	{
		Path:        "/subversion/{svnrepoid}/postcommithook",
		Description: "SVN hooks API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.HandleSVNHooks,
				Description: "Trigger the pipeline by svn hooks",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.SVNRepoIDPathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.SVNRevisionQueryParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
