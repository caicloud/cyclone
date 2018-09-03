package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/token/handlers"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(tokens)
}

var tokens = definition.Descriptor{
	Description: "Service Token API",
	Children: []definition.Descriptor{
		{
			Path:        "/service/token",
			Definitions: []definition.Definition{getToken},
		},
	},
}

var getToken = definition.Definition{
	Method:      definition.Get,
	Description: "Get Service Token",
	Function:    handlers.GetToken,
	Parameters:  []definition.Parameter{},
	Results:     definition.DataErrorResults("token"),
}
