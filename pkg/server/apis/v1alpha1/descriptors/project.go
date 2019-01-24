package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(project...)
}

var project = []definition.Descriptor{
	{
		Path:        "/projects",
		Description: "Projects APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListProjects,
				Description: "List projects",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose projects to list",
					},
					{
						Source:      definition.Auto,
						Name:        "pagination",
						Description: "pagination",
					},
				},
				Results: definition.DataErrorResults("projects"),
			},
			{
				Method:      definition.Create,
				Function:    handler.CreateProject,
				Description: "Create project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant to create project for",
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the new project",
					},
				},
				Results: definition.DataErrorResults("created project"),
			},
		},
	},
	{
		Path:        "/projects/{project}",
		Description: "Projects APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetProject,
				Description: "Get project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose project to get",
					},
					{
						Source:      definition.Path,
						Name:        "project",
						Description: "Name of the project to get",
					},
				},
				Results: definition.DataErrorResults("project gotten"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateProject,
				Description: "Update project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose project to update",
					},
					{
						Source:      definition.Path,
						Name:        "project",
						Description: "Name of the project to update",
					},
					{
						Source:      definition.Body,
						Description: "JSON body to describe the updated project",
					},
				},
				Results: definition.DataErrorResults("project updated"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteProject,
				Description: "Delete project",
				Parameters: []definition.Parameter{
					{
						Source:      definition.Header,
						Name:        httputil.TenantHeaderName,
						Description: "Name of the tenant whose project to delete",
					},
					{
						Source:      definition.Path,
						Name:        "project",
						Description: "Name of the project to delete",
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
