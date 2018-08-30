package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/admin/handlers"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(status)
}

var status = definition.Descriptor{
	Description: "Health Check API",
	Children: []definition.Descriptor{
		{
			Path:        "/healthcheck",
			Definitions: []definition.Definition{getStatus},
		},
	},
}

var getStatus = definition.Definition{
	Method:      definition.Get,
	Description: "Get status of Cargo-Admin",
	Function:    handlers.HealthCheck,
	Parameters:  []definition.Parameter{},
	Results:     []definition.Result{{Destination: definition.Error}},
}
