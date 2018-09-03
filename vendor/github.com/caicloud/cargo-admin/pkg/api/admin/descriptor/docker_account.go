package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"
	"github.com/caicloud/cargo-admin/pkg/api/admin/types"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(dockerAccounts)
}

var dockerAccounts = definition.Descriptor{
	Description: "Docker Account API",
	Children: []definition.Descriptor{
		{
			Path:        "/registries/{registry}/dockeraccount",
			Definitions: []definition.Definition{getDockerAccount},
		},
	},
}

var getDockerAccount = definition.Definition{
	Method:      definition.Get,
	Description: "Get docker account",
	Function:    handlers.GetOrNewDockerAccount,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Path,
			Name:        "registry",
			Description: "registry name",
		},
		{
			Source:      definition.Header,
			Name:        types.HeaderUser,
			Description: "username",
		},
	},
	Results: definition.DataErrorResults("docker account"),
}
