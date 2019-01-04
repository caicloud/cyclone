/*
Copyright 2018 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package descriptors

import (
	"github.com/caicloud/nirvana/definition"

	"github.com/caicloud/cyclone/pkg/server/handler"
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
