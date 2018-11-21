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

package config

import (
	"context"

	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/tracing/services/store"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/tracing"
)

func NewServer() error {
	cmd := config.NewNirvanaCommand(&config.Option{
		Port: 8081,
	})
	cmd.EnablePlugin(
		&tracing.Option{
			ServiceName:   "config",
			AgentHostPort: "127.0.0.1:6831",
		},
	)
	log.Info("API server listening on localhost:8081")
	return cmd.Execute(app)
}

var app = definition.Descriptor{
	Path:        "/config",
	Produces:    []string{definition.MIMEJSON},
	Consumes:    []string{definition.MIMEJSON},
	Description: "config api",
	Definitions: []definition.Definition{
		{
			Method: definition.Get,
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "config",
					Description: "config name",
				},
			},
			Function: getConfig,
			Results:  definition.DataErrorResults("create app"),
		},
	},
}

func getConfig(ctx context.Context, config string) (map[string]string, error) {
	return store.GetConfig(ctx, config)
}
