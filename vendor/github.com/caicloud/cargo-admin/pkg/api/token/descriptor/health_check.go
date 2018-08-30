package descriptor

import (
	"github.com/caicloud/cargo-admin/pkg/api/token/handlers"

	"github.com/caicloud/nirvana/definition"
)

func init() {
	register(status)
}

// This API is used by Cargo to check whether Cargo-Token is healthy to provide token service.
// It will be exposed to outside cluster as NodePort
var status = definition.Descriptor{
	Description: "Health Check API",
	Children: []definition.Descriptor{
		{
			Path:        "/registryproxy/v2",
			Definitions: []definition.Definition{getStatus},
		},
	},
}

var getStatus = definition.Definition{
	Method:      definition.Get,
	Description: "Get health check results",
	Function:    handlers.HealthCheck,
	Parameters: []definition.Parameter{
		{
			Source:      definition.Header,
			Name:        headerCargo,
			Description: "domain of registry",
		},
	},
	Results: definition.DataErrorResults("health check result"),
}
