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

package api

import (
	"context"

	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/tracing/services/models"
	"github.com/caicloud/nirvana/examples/tracing/services/store"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/tracing"
)

func NewServer() error {
	cmd := config.NewNirvanaCommand(&config.Option{
		Port: 8080,
	})
	cmd.EnablePlugin(
		&tracing.Option{
			ServiceName:   "api",
			AgentHostPort: "127.0.0.1:6831",
		},
	)

	log.Info("API server listening on localhost:8080")
	return cmd.Execute(app)
}

type appForm struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Config    string `json:"config"`
}

var app = definition.Descriptor{
	Path:        "/applications",
	Produces:    []string{definition.MIMEJSON},
	Consumes:    []string{definition.MIMEJSON},
	Description: "app",
	Definitions: []definition.Definition{
		{
			Method: definition.Create,
			Parameters: []definition.Parameter{
				{
					Source:      definition.Body,
					Name:        "params",
					Description: "app form",
				},
			},
			Function: createApp,
			Results:  definition.DataErrorResults("create app"),
		},
	},
}

func createApp(ctx context.Context, form *appForm) (*models.Application, error) {
	return store.CreateApplication(ctx, form.Namespace, form.Name, form.Config)
}
