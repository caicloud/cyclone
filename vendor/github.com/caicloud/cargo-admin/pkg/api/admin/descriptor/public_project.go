package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(publicProejcts)
}

var publicProejcts = definition.Descriptor{
	Description: "Public Project API",
	Children: []definition.Descriptor{
		{
			Path:        "/registries/{registry}/publicprojects",
			Definitions: []definition.Definition{createPublicProject, listPublicProjects},
		},
		{
			Path:        "/registries/{registry}/publicprojects/{publicproject}",
			Definitions: []definition.Definition{updatePublicProject, getPublicProject, deletePublicProject},
		},
		{
			Path:        "/registries/{registry}/publicprojects/{publicproject}/stats",
			Definitions: []definition.Definition{listPublicProjectStats},
		},
	},
}

var createPublicProject = definition.Definition{
	Method:      definition.Create,
	Description: "Create Public Project",
	Function:    handlers.CreatePublicProject,
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
			Description: "create public project request body",
		},
	},
	Results: definition.DataErrorResults("public project"),
}

var listPublicProjects = definition.Definition{
	Method:      definition.Get,
	Description: "List Public Projects",
	Function:    handlers.ListPublicProjects,
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
			Source:      definition.Auto,
			Name:        "pagination",
			Description: "pagination",
		},
	},
	Results: []definition.Result{
		{
			Destination: definition.Data,
			Description: "public project list",
		},
		{
			Destination: definition.Meta,
		},
		{
			Destination: definition.Error,
		},
	},
}

var getPublicProject = definition.Definition{
	Method:      definition.Get,
	Description: "Get public project",
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
			Name:        "publicproject",
			Description: "Public Project name",
		},
	},
	Results: definition.DataErrorResults("project"),
}

var updatePublicProject = definition.Definition{
	Method:      definition.Update,
	Description: "Update public project",
	Function:    handlers.UpdatePublicProject,
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
			Name:        "publicproject",
			Description: "public project name",
		},
		{
			Source:      definition.Body,
			Description: "update public project request body",
		},
	},
	Results: definition.DataErrorResults("public project"),
}

var deletePublicProject = definition.Definition{
	Method:      definition.Delete,
	Description: "Delete public project",
	Function:    handlers.DeletePublicProject,
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
			Name:        "publicproject",
			Description: "public project name",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}

var listPublicProjectStats = definition.Definition{
	Method:      definition.List,
	Description: "List Public Project Stats",
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
			Name:        "publicproject",
			Description: "public project name",
		},
		{
			Source:      definition.Query,
			Name:        "action",
			Description: "action",
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
	Results: definition.DataErrorResults("public project stats list"),
}
