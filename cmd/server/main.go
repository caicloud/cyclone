/*
Copyright 2018 caicloud authors. All rights reserved.

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
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/profiling"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1/descriptor"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/k8s"
)

// APIServerOptions contains all options(config) for api server
type APIServerOptions struct {
	KubeHost   string
	KubeConfig string

	CyclonePort int
	CycloneAddr string
}

// NewAPIServerOptions returns a new APIServerOptions
func NewAPIServerOptions() *APIServerOptions {
	return &APIServerOptions{
		CyclonePort: 7099,
	}
}

// AddFlags adds flags to APIServerOptions.
func (opts *APIServerOptions) AddFlags() {

	flag.StringVar(&opts.KubeHost, config.EnvKubeHost, config.KubeHost, "Kube host address")
	flag.StringVar(&opts.KubeConfig, config.EnvKubeConfig, config.KubeConfig, "Kube config file path")

	flag.IntVar(&opts.CyclonePort, config.EnvCycloneAdminPort, 7099, "The port for the cyclone server to serve on.")
	flag.StringVar(&opts.CycloneAddr, config.EnvCycloneAdminAddr, "0.0.0.0", "The IP address for the cyclone server to serve on.")

	flag.Parse()
}

func initialize(opts *APIServerOptions, closing chan struct{}) {
	// Init k8s client
	client, err := k8s.GetClient(opts.KubeHost, opts.KubeConfig)
	if err != nil {
		log.Fatalf("Create k8s client error: %v", err)
	}
	k8s.Client = client
	log.Info("Init k8s client success")

	return
}

func gracefulShutdown(closing chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Infof("capture system signal %s, to close \"closing\" channel", <-signals)
	close(closing)
}

func main() {
	//// Log to standard error instead of files.
	//flag.Set("logtostderr", "true")

	// Flushes all pending log I/O.
	//defer glog.Flush()

	opts := NewAPIServerOptions()
	opts.AddFlags()

	closing := make(chan struct{})

	initialize(opts, closing)

	go gracefulShutdown(closing)

	log.Infof("cyclone starts listening on %s:%v", opts.CycloneAddr, opts.CyclonePort)

	config := nirvana.NewDefaultConfig()
	nirvana.IP(opts.CycloneAddr)(config)
	nirvana.Port(uint16(opts.CyclonePort))(config)
	config.Configure(
		metrics.Path("/metrics"),
		profiling.Path("/debug/pprof/"),
		profiling.Contention(true),
	)

	config.Configure(nirvana.Descriptor(descriptor.Descriptor()))

	log.Infof("API service listening on %s:%d", opts.CycloneAddr, opts.CyclonePort)
	if err := nirvana.NewServer(config).Serve(); err != nil {
		log.Fatal(err)
	}

	log.Info("Cyclone server stopped")
}
