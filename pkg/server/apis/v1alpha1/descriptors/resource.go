package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	handler "github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

func init() {
	register(resource...)
}

var resource = []definition.Descriptor{
	{
		Path:        "/resources",
		Description: "Resource APIs",
		Definitions: []definition.Definition{
			{
				Method:      definition.Create,
				Function:    handler.CreateResource,
				Description: "Create resource",
				Results:     definition.DataErrorResults("resource"),
			},
			{
				Method:      definition.Get,
				Function:    handler.ListResources,
				Description: "List resources",
				Parameters: []definition.Parameter{
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
					},
				},
				Results: definition.DataErrorResults("resources"),
			},
		},
	},
	{
		Path: "/resources/{resource}",
		Definitions: []definition.Definition{
			{
				Method:      definition.Get,
				Function:    handler.GetResource,
				Description: "Get resource",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ResourceNamePathParameterName,
					},
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
					},
				},
				Results: definition.DataErrorResults("resource"),
			},
			{
				Method:      definition.Update,
				Function:    handler.UpdateResource,
				Description: "Update resource",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ResourceNamePathParameterName,
					},
				},
				Results: definition.DataErrorResults("resource"),
			},
			{
				Method:      definition.Delete,
				Function:    handler.DeleteResource,
				Description: "Delete resource",
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   httputil.ResourceNamePathParameterName,
					},
					{
						Source:  definition.Query,
						Name:    httputil.NamespaceQueryParameter,
						Default: httputil.DefaultNamespace,
					},
				},
				Results: []definition.Result{definition.ErrorResult()},
			},
		},
	},
}
