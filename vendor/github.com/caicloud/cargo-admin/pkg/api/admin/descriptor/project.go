package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(projects)
}

var projects = definition.Descriptor{
	Description: "Project API",
	Children: []definition.Descriptor{
		{
			Path:        "/registries/{registry}/projects",
			Definitions: []definition.Definition{createProject, listProjects},
		},
		{
			Path:        "/registries/{registry}/projects/{project}",
			Definitions: []definition.Definition{getProject, updateProject, deleteProject},
		},
		{
			Path:        "/registries/{registry}/projects/{project}/stats",
			Definitions: []definition.Definition{listProjectStats},
		},
	},
}

var createProject = definition.Definition{
	Method:      definition.Create,
	Description: "Create project",
	Function:    handlers.CreateProject,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Body,
			Description: "create project request body",
		},
	},
	Results: definition.DataErrorResults("project"),
}

var listProjects = definition.Definition{
	Method:      definition.List,
	Description: "List projects",
	Function:    handlers.ListProjects,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderSeqID,
			Description: "sequence id",
		},
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Query,
			Name:        "includePublic",
			Default:     false,
			Description: "whether include public project",
		},
		{
			Source:      definition.Query,
			Name:        "q",
			Description: "query keyword",
		},
		{
			Source:      definition.Auto,
			Name:        "pagination",
			Description: "pagination",
		},
	},
	Results: []definition.Result{
		{
			Destination: definition.Data,
			Description: "project list",
		},
		{
			Destination: definition.Meta,
		},
		{
			Destination: definition.Error,
		},
	},
}

var getProject = definition.Definition{
	Method:      definition.Get,
	Description: "Get project",
	Function:    handlers.GetProject,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Path,
			Name:        "project",
			Description: "Project name",
		},
	},
	Results: definition.DataErrorResults("project"),
}

var updateProject = definition.Definition{
	Method:      definition.Update,
	Description: "Update project",
	Function:    handlers.UpdateProject,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Path,
			Name:        "project",
			Description: "Project name",
		},
		{
			Source:      definition.Body,
			Description: "update public project request body",
		},
	},
	Results: definition.DataErrorResults("project"),
}

var deleteProject = definition.Definition{
	Method:      definition.Delete,
	Description: "Delete project",
	Function:    handlers.DeleteProject,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Path,
			Name:        "project",
			Description: "Project name",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}

var listProjectStats = definition.Definition{
	Method:      definition.List,
	Description: "List project's stats",
	Function:    handlers.ListProjectStats,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Path,
			Name:        "project",
			Description: "project name",
		},
		{
			Source:      definition.Query,
			Name:        "action",
			Description: "whitch kind of action",
		},
		{
			Source:      definition.Query,
			Name:        "startTime",
			Description: "start time",
		},
		{
			Source:      definition.Query,
			Name:        "endTime",
			Description: "end time",
		},
	},
	Results: definition.DataErrorResults("project stats list"),
}
