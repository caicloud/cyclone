package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(tags)
}

var tags = definition.Descriptor{
	Description: "Tag API",
	Children: []definition.Descriptor{
		{
			Path:        "/registries/{registry}/projects/{project}/repositories/{repository}/tags",
			Definitions: []definition.Definition{listTags},
		},
		{
			Path:        "/registries/{registry}/projects/{project}/repositories/{repository}/tags/{tag}",
			Definitions: []definition.Definition{getTag, deleteTag},
		},
		{
			Path:        "/registries/{registry}/publicprojects/{project}/repositories/{repository}/tags",
			Definitions: []definition.Definition{listTags},
		},
		{
			Path:        "/registries/{registry}/publicprojects/{project}/repositories/{repository}/tags/{tag}",
			Definitions: []definition.Definition{getTag, deleteTag},
		},
	},
}

var listTags = definition.Definition{
	Method:      definition.List,
	Description: "List tags",
	Function:    handlers.ListTags,
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
	Results: definition.DataErrorResults("tag list"),
}

var getTag = definition.Definition{
	Method:      definition.Get,
	Description: "Get tag",
	Function:    handlers.GetTag,
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
			Source:      definition.Path,
			Name:        "tag",
			Description: "image tag",
		},
	},
	Results: definition.DataErrorResults("tag"),
}

var deleteTag = definition.Definition{
	Method:      definition.Delete,
	Description: "Delete tag",
	Function:    handlers.DeleteTag,
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
			Source:      definition.Path,
			Name:        "tag",
			Description: "image tag",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}
