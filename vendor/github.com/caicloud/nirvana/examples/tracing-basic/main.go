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

package main

import (
	"context"
	"errors"

	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/tracing"
	"github.com/caicloud/nirvana/service"
)

func main() {
	example := definition.Descriptor{
		Path:        "/",
		Description: "trace example",
		Definitions: []definition.Definition{
			{
				Method: definition.Get,
				Function: func(ctx context.Context) (string, error) {
					msg := service.HTTPContextFrom(ctx).Request().URL.Query().Get("msg")
					if msg != "" {
						return "", errors.New(msg)
					}
					return "success", nil
				},
				Consumes: []string{"application/json"},
				Produces: []string{"application/json"},
				Results: []definition.Result{
					{
						Destination: definition.Data,
					},
					{
						Destination: definition.Error,
					},
				},
			},
		},
	}

	cmd := config.NewDefaultNirvanaCommand()
	cmd.EnablePlugin(
		&tracing.Option{
			ServiceName:   "example",
			AgentHostPort: "127.0.0.1:6831",
		},
	)

	log.Infof("Listening on localhost:8080")
	if err := cmd.Execute(example); err != nil {
		log.Error(err)
	}
}
