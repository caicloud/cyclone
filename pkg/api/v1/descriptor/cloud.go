package descriptor

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/api/v1/handler"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(clouds...)
}

var clouds = []definition.Descriptor{
	{
		Path:        "/clouds",
		Description: "Cloud API",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListClouds,
				Description: "List all clouds",
				Parameters:  []definition.Parameter{},
				Results:     definition.DataErrorResults("list clouds"),
			},
			{
				Method:      definition.Create,
				Function:    handler.CreateCloud,
				Description: "Add a cloud",
				Results:     definition.DataErrorResults("cloud"),
			},
		},
	},
	{
		Path:        "/clouds/{cloud}",
		Description: "Cloud API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Delete,
				Function:    handler.DeleteCloud,
				Description: "Delete the cloud",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.CloudPathParameterName,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
	{
		Path:        "/clouds/{cloud}/ping",
		Description: "Cloud API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.PingCloud,
				Description: "Ping the cloud to check its health",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.CloudPathParameterName,
					},
				},
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Description: "clouds",
					},
				},
			},
		},
	},
	{
		Path:        "/clouds/{cloud}/workers",
		Description: "Cloud API",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.ListWorkers,
				Description: "Get all cyclone workers in the cloud",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.CloudPathParameterName,
					},
					{
						Source: definition.Query,
						Name:   httputil.NamespaceQueryParameter,
					},
				},
				Results: definition.DataErrorResults("workers"),
			},
		},
	},
}
