/*
Copyright 2016 caicloud authors. All rights reserved.

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
	"crypto/aes"

	log "github.com/golang/glog"
	"gopkg.in/urfave/cli.v1"

	"github.com/caicloud/cyclone/api/server"
	"github.com/caicloud/cyclone/cmd/worker/options"
)

// ServerOptions ...
type ServerOptions struct {
	WorkerOptions    *options.WorkerOptions
	APIServerOptions *server.APIServerOptions
}

// NewServerOptions creates a new ServerOptions
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		WorkerOptions:    options.NewWorkerOptions(),
		APIServerOptions: server.NewAPIServerOptions(),
	}
}

// AddFlags adds flags to app.Flags
func (opts *ServerOptions) AddFlags(app *cli.App) {
	opts.WorkerOptions.AddFlags(app)
	opts.APIServerOptions.AddFlags(app)

}

// NewAPIServer returns a new APIServer with config
func (opts *ServerOptions) NewAPIServer() *server.APIServer {
	s := &server.APIServer{
		Config:        opts.APIServerOptions,
		WorkerOptions: opts.WorkerOptions,
	}

	log.Infof("Start server with: server options: %+v; worker options: %+v", opts.APIServerOptions, opts.WorkerOptions)
	return s
}

// Validate validates options.
func (opts *ServerOptions) Validate() error {
	serverOpts := opts.APIServerOptions
	_, err := aes.NewCipher([]byte(serverOpts.SaltKey))
	return err
}
