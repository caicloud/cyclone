package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(registryClaims)
}

var registryClaims = definition.Descriptor{
	Description: "Registry Claim API",
	Children: []definition.Descriptor{
		{
			Path:        "/registryclaims",
			Definitions: []definition.Definition{createRegistryClaim, listRegistryClaims},
		},
		{
			Path:        "/registryclaims/{registryclaims}",
			Definitions: []definition.Definition{getRegistryClaim, updateRegistryClaim, deleteRegistryClaim},
		},
	},
}

var createRegistryClaim = definition.Definition{
	Method:      definition.Create,
	Description: "Create registryclaim",
	Function:    handlers.CreateRegistry,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Body,
			Description: "create registryclaim request body",
		},
	},
	Results: definition.DataErrorResults("registryclaim"),
}

var listRegistryClaims = definition.Definition{
	Method:      definition.List,
	Description: "List registryclaims",
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
	Results: definition.DataErrorResults("registryclaim list"),
}

var getRegistryClaim = definition.Definition{
	Method:      definition.Get,
	Description: "Get registryclaim",
	Function:    handlers.GetRegistry,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registryclaims",
			Description: "registryclaim name",
		},
	},
	Results: definition.DataErrorResults("repo"),
}

var updateRegistryClaim = definition.Definition{
	Method:      definition.Update,
	Description: "Update registryclaim",
	Function:    handlers.UpdateRegistry,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registryclaims",
			Description: "registryclaim name",
		},
		{
			Source:      definition.Body,
			Description: "update registryclaim request body",
		},
	},
	Results: definition.DataErrorResults("registryclaim"),
}

var deleteRegistryClaim = definition.Definition{
	Method:      definition.Delete,
	Description: "Delete registryclaim",
	Function:    handlers.DeleteRegistry,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        types.HeaderTenant,
			Description: "tenant name",
		},
		{
			Source:      definition.Path,
			Name:        "registryclaims",
			Description: "registryclaim name",
		},
	},
	Results: []definition.Result{{Destination: definition.Error}},
}
