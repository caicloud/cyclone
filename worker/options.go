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

package worker

import (
	"github.com/caicloud/cyclone/cloud"
	cli "gopkg.in/urfave/cli.v1"
)

// Config contains all options(config) for worker
type Config struct {
	ID         string
	DockerHost string
}

// NewConfig returns a new worker config
func NewConfig() *Config {
	return &Config{
		DockerHost: "unix:///var/run/docker.sock",
	}
}

// AddFlags adds flags to cli.App
func (opts *Config) AddFlags(app *cli.App) {

	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "id",
			Usage:       "worker event id",
			EnvVar:      cloud.WorkerEventID,
			Destination: &opts.ID,
		},
		cli.StringFlag{
			Name:        "docker-host",
			Value:       "unix:///var/run/docker.sock",
			Usage:       "worker used docker host",
			EnvVar:      "DOCKER_HOST",
			Destination: &opts.DockerHost,
		},
	}

	app.Flags = append(app.Flags, flags...)

}
