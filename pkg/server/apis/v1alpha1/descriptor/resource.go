/*
Copyright 2018 caicloud authors. All rights reserved.
*/

package descriptor

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1/handler"
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
		},
	},
	{
		Path: "/resources/{resource-name}",
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
						Source: definition.Query,
						Name:   httputil.NamespaceQueryParameter,
					},
				},
				Results: definition.DataErrorResults("resource"),
			},
		},
	},
}
