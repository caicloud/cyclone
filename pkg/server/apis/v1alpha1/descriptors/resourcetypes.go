package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(resourceTypes...)
}

var resourceTypes = []definition.Descriptor{
	{
		Path:        "/resourcetypes",
		Description: "Resource types APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.List,
				Function:    handler.ListResourceTypes,
				Description: "List supported resource types",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Query,
						Name:        httputil.OperationQueryParameter,
						Description: "Operation the resource type should support, e.g. 'pull', 'push'",
					},
				},
				Results: definition.DataErrorResults("resource types"),
			},
		},
	},
	{
		Path:        "/resourcetypes/{resourceType}",
		Description: "Resource type APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetResourceType,
				Description: "Get resource type",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ResourceTypePathParameterName,
					},
				},
				Results: definition.DataErrorResults("resource"),
			},
		},
	},
}
