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

package cloud

import (
	"time"

	cli "gopkg.in/urfave/cli.v1"
)

// env
const (

	// WorkerEventID is special
	WorkerEventID = "WORKER_EVENTID"

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

	LogServer = "LOG_SERVER"

	// Resource

	LimitMemory = "LIMIT_MEMORY"
	LimitCPU    = "LIMIT_CPU"

	WorkingDir = "/root/code"
)

const (
	// WorkerTimeout ...
	WorkerTimeout = 2 * time.Hour
)

// -----------------------------------------------------------------------

// WorkerEnvs ...
type WorkerEnvs struct {
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

	LogServer string

	WorkerImage string
}

// NewWorkerEnvs creates a new WorkerEnvs
func NewWorkerEnvs() *WorkerEnvs {
	return &WorkerEnvs{
		WorkerImage: "cargo.caicloud.io/caicloud/cyclone-worker:v1221",
	}
}

// AddFlags adds flags for a specific APIServer to the specified cli.app
func (env *WorkerEnvs) AddFlags(app *cli.App) {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "cyclone-server",
			Value:       "http://127.0.0.1:7099",
			Usage:       "cyclone server host for worker to connect server",
			EnvVar:      CycloneServer,
			Destination: &env.CycloneServer,
		},
		cli.StringFlag{
			Name:        "console-web-endpoint",
			Value:       "http://127.0.0.1:3000",
			Usage:       "for worker to deploy to caicloud kubernetes",
			EnvVar:      ConsoleWebEndpoint,
			Destination: &env.ConsoleWebEndpoint,
		},

		cli.StringFlag{
			Name:        "registry",
			Value:       "cargo.caicloud.io",
			Usage:       "docker registry location for docker push",
			EnvVar:      RegistryLocation,
			Destination: &env.RegistryLocation,
		},
		cli.StringFlag{
			Name:        "registry-username",
			Usage:       "docker registry username",
			EnvVar:      RegistryUsername,
			Destination: &env.RegistryUsername,
		},
		cli.StringFlag{
			Name:        "registry-password",
			Usage:       "docker registry password",
			EnvVar:      RegistryPassword,
			Destination: &env.RegistryPassword,
		},

		// Github
		cli.StringFlag{
			Name:        "github-client",
			Usage:       "github client id",
			EnvVar:      GithubClient,
			Destination: &env.GithubClient,
		},
		cli.StringFlag{
			Name:        "github-secret",
			Usage:       "github client secret",
			EnvVar:      GithubSecret,
			Destination: &env.GithubSecret,
		},
		// Gitlab
		cli.StringFlag{
			Name:        "gitlab-url",
			Value:       "https://gitlab.com",
			Usage:       "gitlab url domain",
			EnvVar:      GitlabURL,
			Destination: &env.GitlabURL,
		},
		cli.StringFlag{
			Name:        "gitlab-client",
			Usage:       "gitlab client id",
			EnvVar:      GitlabClient,
			Destination: &env.GitlabClient,
		},
		cli.StringFlag{
			Name:        "gitlab-secret",
			Usage:       "gitlab client secret",
			EnvVar:      GitlabSecret,
			Destination: &env.GitlabSecret,
		},
		cli.StringFlag{
			Name:        "log-server",
			Value:       "ws://127.0.0.1:8000/ws",
			Usage:       "cyclone log server websocket host",
			EnvVar:      LogServer,
			Destination: &env.LogServer,
		},

		cli.StringFlag{
			Name:        "worker-image",
			Value:       "cargo.caicloud.io/caicloud/cyclone-worker",
			Usage:       "basic worker image",
			EnvVar:      WorkerImage,
			Destination: &env.WorkerImage,
		},
	}
	app.Flags = append(app.Flags, flags...)
}

// ---------------------------------------------------------------------------

// WorkerOptions contains the options for workers creation
type WorkerOptions struct {
	WorkerEnvs *WorkerEnvs

	Quota Quota

	// Namespace represents the k8s namespace where to create worker, only works for k8s cloud provider.
	Namespace string

	// CacheVolume represents the volume to cache dependency for worker.
	CacheVolume string

	// MountPath represents the mount path for the cache volume.
	MountPath string
}

// NewWorkerOptions creates a new WorkerOptions with default value
func NewWorkerOptions() *WorkerOptions {
	return &WorkerOptions{
		WorkerEnvs: NewWorkerEnvs(),
		Quota:      DefaultQuota.DeepCopy(),
	}
}

// DeepCopy returns a deep-copy of the WorkerOptions value.  Note that the method
// receiver is a value, so we can mutate it in-place and return it.
func (opts WorkerOptions) DeepCopy() WorkerOptions {
	opts.Quota = opts.Quota.DeepCopy()
	return opts
}

// AddFlags adds flags for a specific APIServer to the specified cli.app
func (opts *WorkerOptions) AddFlags(app *cli.App) {

	// add env flags
	opts.WorkerEnvs.AddFlags(app)

	// add quota flags
	if opts.Quota == nil {
		opts.Quota = DefaultQuota.DeepCopy()
	}

	flags := []cli.Flag{
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
	}
	app.Flags = append(app.Flags, flags...)
}

// ---------------------------------------------------------------------------

// WorkerInfo ...
type WorkerInfo struct {
	CloudName string `json:"cloudName,omitempty" bson:"cloudName,omitempty"`
	CloudKind string `json:"cloudKind,omitempty" bson:"cloudKind,omitempty"`

	Name       string    `json:"name,omitempty" bson:"name,omitempty"`
	CreateTime time.Time `json:"createTime,omitempty" bson:"createTime,omitempty"`
	DueTime    time.Time `json:"dueTime,omitempty" bson:"dueTime,omitempty"`

	// for k8s
	PodName   string `json:"podName,omitempty" bson:"podName,omitempty"`
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty"`

	// for docker
	ContainerID string `json:"containerID,omitempty" bson:"containerID,omitempty"`
}
