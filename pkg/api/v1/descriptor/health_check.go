/*
Copyright 2018 caicloud authors. All rights reserved.
*/

package descriptor

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
)

func init() {
	register(healthCheck...)
}

var healthCheck = []definition.Descriptor{
	{
		Path:        "/healthcheck",
		Description: "health check API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.HealthCheck,
				Description: "Health check for cyclone",
				Results: []definition.Result{
					{
						Destination: definition.Error,
					},
				},
			},
		},
	},
}
