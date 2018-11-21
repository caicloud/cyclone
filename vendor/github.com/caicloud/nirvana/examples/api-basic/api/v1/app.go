/*
Copyright 2017 Caicloud Authors

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

package v1

import (
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/api-basic/api/v1/converters"
	"github.com/caicloud/nirvana/examples/api-basic/application"
)

func init() {
	register(app)
}

var app = definition.Descriptor{
	Path:        "/applications",
	Description: "Application API",
	Definitions: []definition.Definition{
		{
			Method:      definition.Create,
			Description: "Create Application",
			Function:    application.CreateApplication,
			Consumes:    []string{definition.MIMEJSON},
			Produces:    []string{definition.MIMEJSON},
			Parameters: []definition.Parameter{
				{
					Source: definition.Body,
					Operators: []definition.Operator{
						converters.ConvertApplicationV1ToApplication(),
					},
				},
			},
			Results: []definition.Result{
				{
					Destination: definition.Data,
					Operators: []definition.Operator{
						converters.ConvertApplicationToApplicationV1(),
					},
				},
				definition.ErrorResult(),
			},
		},
	},
}
