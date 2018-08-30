package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(repositories)
}

var repositories = definition.Descriptor{
	Description: "Repository API",
	Children: []definition.Descriptor{
		{
			Path:        "/registries/{registry}/projects/{project}/repositories",
			Definitions: []definition.Definition{listRepositories},
		},
		{
			Path:        "/registries/{registry}/projects/{project}/repositories/{repository}",
			Definitions: []definition.Definition{getRepository, updateRepository, deleteRepository},
		},
		{
			Path:        "/registries/{registry}/publicprojects/{project}/repositories",
			Definitions: []definition.Definition{listRepositories},
		},
		{
			Path:        "/registries/{registry}/publicprojects/{project}/repositories/{repository}",
			Definitions: []definition.Definition{getRepository, updateRepository, deleteRepository},
		},
	},
}

var listRepositories = definition.Definition{
	Method:      definition.List,
	Description: "List repositories",
	Function:    handlers.ListRepositories,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderSeqID,
			Description: "request sequence ID",
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
			Source:      definition.Path,
			Name:        "project",
			Description: "project name",
		},
		{
			Source:      definition.Query,
			Name:        "q",
			Description: "query keyword",
		},
		{
			Source:      definition.Query,
			Name:        "sort",
			Description: "sort method, by name or date",
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
			Description: "repository list",
		},
		{
			Destination: definition.Meta,
		},
		{
			Destination: definition.Error,
		},
	},
}

var getRepository = definition.Definition{
	Method:      definition.Get,
	Description: "Get repository",
	Function:    handlers.GetRepository,
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
			Source:      definition.Path,
			Name:        "repository",
			Description: "repository name",
		},
	},
	Results: definition.DataErrorResults("repository"),
}

var updateRepository = definition.Definition{
	Method:      definition.Update,
	Description: "Update repository",
	Function:    handlers.UpdateRepository,
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
			Source:      definition.Path,
			Name:        "repository",
			Description: "repository name",
		},
		{
			Source:      definition.Body,
			Default:     &types.UpdateRepositoryReq{Spec: &types.RepositorySpec{Description: ""}},
			Description: "repository name",
		},
	},
	Results: definition.DataErrorResults("repository"),
}

var deleteRepository = definition.Definition{
	Method:      definition.Delete,
	Description: "Delete repository",
	Function:    handlers.DeleteRepository,
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
			Source:      definition.Path,
			Name:        "repository",
			Description: "repository name",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}
