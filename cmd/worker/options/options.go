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

package options

import (
	"time"

	"gopkg.in/urfave/cli.v1"
)

// NewWorkerOptions creates a new WorkerOptions
func NewWorkerOptions() *WorkerOptions {
	return &WorkerOptions{}
}

const (
	CycloneServer      = "CYCLONE_SERVER"
	ConsoleWebEndpoint = "CONSOLE_WEB_ENDPOINT"

	// CallbackURL ...
	CallbackURL = "CALLBACK_URL"

	// Registry
	RegistryLocation = "REGISTRY_LOCATION"
	RegistryUsername = "REGISTRY_USERNAME"
	RegistryPassword = "REGISTRY_PASSWORD"

	WorkerImage = "WORKER_IMAGE"

	// Github
	GithubClient = "GITHUB_CLIENT"
	GithubSecret = "GITHUB_SECRET"

	//Gitlab
	GitlabURL    = "GITLAB_URL"
	GitlabClient = "GITLAB_CLIENT"
	GitlabSecret = "GITLAB_SECRET"

	// Resource
	LimitMemory   = "LIMIT_MEMORY"
	LimitCPU      = "LIMIT_CPU"
	RequestMemory = "REQUEST_MEMORY"
	RequestCPU    = "REQUEST_CPU"

	WorkingDir = "/root/code"

	// EventID for worker to get the event.
	EventID = "EVENT_ID"

	DockerHost = "DOCKER_HOST"

	// WorkerTimeout ...
	WorkerTimeout = 2 * time.Hour

	// DinDParameter represents the cyclone worker DinD starting parameter, e.g. --bip=192.168.1.5/24
	DinDParameter = "DIND_PARAMETER"
)

// WorkerOptions ...
type WorkerOptions struct {
	// for worker env
	CycloneServer      string
	ConsoleWebEndpoint string

	// Registry
	RegistryLocation string `json:"registryLocation,omitempty"`
	RegistryUsername string `json:"registryUsername,omitempty"`
	RegistryPassword string `json:"registryPassword,omitempty"`

	// github
	GithubClient string
	GithubSecret string

	// gitlab
	GitlabURL    string
	GitlabClient string
	GitlabSecret string

	WorkerImage  string
	EventID      string
	ProjectName  string
	PipelineName string
	DockerHost   string

	Quota Quota

	// DinDParameter represents the cyclone worker DinD starting parameter, e.g. --bip=192.168.1.5/24
	DinDParameter string
}

// AddFlags adds flags to app.Flags
func (opts *WorkerOptions) AddFlags(app *cli.App) {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "cyclone-server",
			Value:       "http://127.0.0.1:7099",
			Usage:       "cyclone server host for worker to connect server",
			EnvVar:      CycloneServer,
			Destination: &opts.CycloneServer,
		},
		cli.StringFlag{
			Name:        "console-web-endpoint",
			Value:       "http://127.0.0.1:3000",
			Usage:       "for worker to deploy to caicloud kubernetes",
			EnvVar:      ConsoleWebEndpoint,
			Destination: &opts.ConsoleWebEndpoint,
		},

		cli.StringFlag{
			Name:        "registry",
			Value:       "cargo.caicloud.io",
			Usage:       "docker registry location for docker push",
			EnvVar:      RegistryLocation,
			Destination: &opts.RegistryLocation,
		},
		cli.StringFlag{
			Name:        "registry-username",
			Usage:       "docker registry username",
			EnvVar:      RegistryUsername,
			Destination: &opts.RegistryUsername,
		},
		cli.StringFlag{
			Name:        "registry-password",
			Usage:       "docker registry password",
			EnvVar:      RegistryPassword,
			Destination: &opts.RegistryPassword,
		},

		// Github
		cli.StringFlag{
			Name:        "github-client",
			Usage:       "github client id",
			EnvVar:      GithubClient,
			Destination: &opts.GithubClient,
		},
		cli.StringFlag{
			Name:        "github-secret",
			Usage:       "github client secret",
			EnvVar:      GithubSecret,
			Destination: &opts.GithubSecret,
		},

		// Gitlab
		cli.StringFlag{
			Name:        "gitlab-url",
			Value:       "https://gitlab.com",
			Usage:       "gitlab url domain",
			EnvVar:      GitlabURL,
			Destination: &opts.GitlabURL,
		},
		cli.StringFlag{
			Name:        "gitlab-client",
			Usage:       "gitlab client id",
			EnvVar:      GitlabClient,
			Destination: &opts.GitlabClient,
		},
		cli.StringFlag{
			Name:        "gitlab-secret",
			Usage:       "gitlab client secret",
			EnvVar:      GitlabSecret,
			Destination: &opts.GitlabSecret,
		},
		cli.StringFlag{
			Name:        "worker-image",
			Value:       "cargo.caicloud.io/caicloud/cyclone-worker",
			Usage:       "basic worker image",
			EnvVar:      WorkerImage,
			Destination: &opts.WorkerImage,
		},
		cli.StringFlag{
			Name:        "event-id",
			Usage:       "id of event to handle",
			EnvVar:      EventID,
			Destination: &opts.EventID,
		},
		cli.StringFlag{
			Name:        "docker-host",
			Value:       "unix:///var/run/docker.sock",
			Usage:       "worker used docker host",
			EnvVar:      DockerHost,
			Destination: &opts.DockerHost,
		},
		cli.StringFlag{
			Name:        "dind-parameter",
			Value:       "",
			Usage:       "the cyclone worker DinD starting parameter, e.g. --bip=192.168.1.5/24",
			EnvVar:      DinDParameter,
			Destination: &opts.DinDParameter,
		},
	}
	app.Flags = append(app.Flags, flags...)

	// For quota flags
	if opts.Quota == nil {
		opts.Quota = DefaultQuota.DeepCopy()
	}

	flags = []cli.Flag{
		cli.GenericFlag{
			Name:   "limit-memory",
			Value:  opts.Quota[ResourceLimitsMemory], // default 512Mi
			Usage:  "default limit memory for worker",
			EnvVar: LimitMemory,
		},
		cli.GenericFlag{
			Name:   "limit-cpu",
			Value:  opts.Quota[ResourceLimitsCPU], // 0.5core
			Usage:  "default limit cpu for worker",
			EnvVar: LimitCPU,
		},
		cli.GenericFlag{
			Name:   "request-memory",
			Value:  opts.Quota[ResourceRequestsMemory],
			Usage:  "default request memory for worker",
			EnvVar: RequestMemory,
		},
		cli.GenericFlag{
			Name:   "request-cpu",
			Value:  opts.Quota[ResourceRequestsCPU],
			Usage:  "default request cpu for worker",
			EnvVar: RequestCPU,
		},
	}
	app.Flags = append(app.Flags, flags...)
}
