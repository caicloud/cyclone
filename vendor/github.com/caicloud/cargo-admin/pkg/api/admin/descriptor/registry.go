package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(registries)
}

var registries = definition.Descriptor{
	Description: "Registry API",
	Children: []definition.Descriptor{
		{
			Path:        "/registries",
			Definitions: []definition.Definition{listRegistries},
		},
		{
			Path:        "/registries/{registry}",
			Definitions: []definition.Definition{getRegistry},
		},
		{
			Path:        "/registries/{registry}/stats",
			Definitions: []definition.Definition{listRegistryStats},
		},
		{
			Path:        "/registries/{registry}/usages",
			Definitions: []definition.Definition{listRegistryUsages},
		},
	},
}

var listRegistries = definition.Definition{
	Method:      definition.List,
	Description: "List registries",
	Function:    handlers.ListRegistries,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Auto,
			Name:        "pagination",
			Description: "pagination",
		},
	},
	Results: definition.DataErrorResults("registry list"),
}

var getRegistry = definition.Definition{
	Method:      definition.Get,
	Description: "Get registry",
	Function:    handlers.GetRegistry,
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
	},
	Results: definition.DataErrorResults("repo"),
}

var listRegistryStats = definition.Definition{
	Method:      definition.List,
	Description: "List registry stats",
	Function:    handlers.ListRegistryStats,
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
	Results: definition.DataErrorResults("registry stats list"),
}

var listRegistryUsages = definition.Definition{
	Method:      definition.List,
	Description: "Get registry usages",
	Function:    handlers.ListRegistryUsages,
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
			Source:      definition.Auto,
			Name:        "pagination",
			Description: "pagination",
		},
	},
	Results: definition.DataErrorResults("registry usages list"),
}
