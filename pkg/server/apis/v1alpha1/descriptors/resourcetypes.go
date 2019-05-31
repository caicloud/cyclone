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
			{
				Method:      definition.Create,
				Function:    handler.CreateResourceType,
				Description: "Create new resource type",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source:      definition.Body,
						Description: "resource type to be created",
					},
				},
				Results: definition.DataErrorResults("resource type created"),
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
			{
				Method:      definition.Update,
				Function:    handler.UpdateResourceType,
				Description: "Update resource type",
				Parameters: []definition.Parameter{
					{
						Source: definition.Header,
						Name:   httputil.TenantHeaderName,
					},
					{
						Source: definition.Path,
						Name:   httputil.ResourceTypePathParameterName,
					},
					{
						Source:      definition.Body,
						Description: "resource type to be created",
					},
				},
				Results: definition.DataErrorResults("resource"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteResourceType,
				Description: "Delete resource type",
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
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
