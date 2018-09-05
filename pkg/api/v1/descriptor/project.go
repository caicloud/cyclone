package descriptor

import (
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/operators/validator"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(projects...)
}

var projects = []definition.Descriptor{
	{
		Path:        "/projects",
		Description: "Projects API",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListProjects,
				Description: "List all projects",
				Parameters:  []definition.Parameter{},
				Results:     definition.DataErrorResults("list projects"),
			},
			{
				Method:      definition.Create,
				Function:    handler.CreateProject,
				Description: "Add a project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.HeaderUser,
					},
				},
				Results: definition.DataErrorResults("project"),
			},
		},
	},
	{
		Path:        "/projects/{project}",
		Description: "Projects API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetProject,
				Description: "Get tht project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
				},
				Results: definition.DataErrorResults("project"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateProject,
				Description: "Update the project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
				},
				Results: definition.DataErrorResults("project"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteProject,
				Description: "Delete the project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path:        "/projects/{project}/repos",
		Description: "Project repos API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListRepos,
				Description: "List accessible repos of the project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
				},
				Results: definition.DataErrorResults("repos"),
			},
		},
	},
	{
		Path:        "/projects/{project}/branches",
		Description: "Project branches API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListBranches,
				Description: "List branches of the repo for the project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source:    definition.Query,
						Name:      httputil.RepoQueryParameter,
						Operators: []definition.Operator{validator.String("required")},
					},
				},
				Results: definition.DataErrorResults("branches"),
			},
		},
	},
	{
		Path:        "/projects/{project}/tags",
		Description: "Project tags API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListTags,
				Description: "List tags of the repo for the project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source:    definition.Query,
						Name:      httputil.RepoQueryParameter,
						Operators: []definition.Operator{validator.String("required")},
					},
				},
				Results: definition.DataErrorResults("tags"),
			},
		},
	},
	{
		Path:        "/projects/{project}/templatetype",
		Description: "Project template API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetTemplateType,
				Description: "Get template type of the repo for the project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
					},
					{
						Source:    definition.Query,
						Name:      httputil.RepoQueryParameter,
						Operators: []definition.Operator{validator.String("required")},
					},
				},
				Results: definition.DataErrorResults("template type"),
			},
		},
	},
	{
		Path:        "/projects/{project}/stats",
		Description: "Project stats API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetProjectStatistics,
				Description: "Get statistics of the project",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ProjectPathParameterName,
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
				Results: definition.DataErrorResults("project stats"),
			},
		},
	},
}
