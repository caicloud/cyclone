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
	"github.com/caicloud/nirvana/examples/stream/stream"
)

func init() {
	register(app)
}

var app = definition.Descriptor{
	Path:        "/stream",
	Description: "Example for explaining how to write stream",
	Definitions: []definition.Definition{
		{
			Method:      definition.Get,
			Description: "Return a stream",
			Function:    stream.Stream,
			Produces:    []string{"application/somedata"},
			Results:     definition.DataErrorResults("Data from stream"),
		},
	},
}
