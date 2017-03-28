package main

import (
	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/worker"
	"gopkg.in/urfave/cli.v1"
)

// WorkerOptions ...
type WorkerOptions struct {
	WorkerEnvs *cloud.WorkerEnvs
	Config     *worker.Config
}

// NewWorkerOptions creates a new ServerOptions
func NewWorkerOptions() *WorkerOptions {
	return &WorkerOptions{
		WorkerEnvs: cloud.NewWorkerEnvs(),
		Config:     worker.NewConfig(),
	}
}

// AddFlags adds flags to app.Flags
func (opts *WorkerOptions) AddFlags(app *cli.App) {
	opts.WorkerEnvs.AddFlags(app)
	opts.Config.AddFlags(app)
}

// NewWorker returns a new APIServer with config
func (opts *WorkerOptions) NewWorker() *worker.Worker {
	s := &worker.Worker{
		Envs:   opts.WorkerEnvs,
		Config: opts.Config,
	}
	return s
}
