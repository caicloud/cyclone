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

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/server/apis"
	"github.com/caicloud/cyclone/pkg/server/apis/filters"
	"github.com/caicloud/cyclone/pkg/server/apis/modifiers"
	"github.com/caicloud/cyclone/pkg/server/config"
	hcommon "github.com/caicloud/cyclone/pkg/server/handler/common"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/version"

	"github.com/caicloud/nirvana"
	nconfig "github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/logger"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/reqlog"
	pversion "github.com/caicloud/nirvana/plugins/version"
)

// APIServerOptions contains all options(config) for api server
type APIServerOptions struct {
	KubeHost   string
	KubeConfig string

	CyclonePort int
	CycloneAddr string

	Loglevel string

	// StorageClass is used to create pvc for default tenant
	StorageClass string
}

// NewAPIServerOptions returns a new APIServerOptions
func NewAPIServerOptions() *APIServerOptions {
	return &APIServerOptions{
		CycloneAddr: "",
		CyclonePort: 7099,
	}
}

// AddFlags adds flags to APIServerOptions.
func (opts *APIServerOptions) AddFlags() {
	flag.StringVar(&opts.KubeHost, config.FlagKubeHost, config.KubeHost, "Kube host address")
	flag.StringVar(&opts.KubeConfig, config.FlagKubeConfig, config.KubeConfig, "Kube config file path")
	flag.IntVar(&opts.CyclonePort, config.FlagCycloneServerPort, config.CycloneServerPort, "The port for the cyclone server to serve on")
	flag.StringVar(&opts.CycloneAddr, config.FlagCycloneServerHost, config.CycloneServerHost, "The IP address for the cyclone server to serve on")
	flag.StringVar(&opts.Loglevel, config.FlagLogLevel, config.LogLevel, "Log level")
	flag.StringVar(&opts.StorageClass, config.FlagStrorageClass, config.StorageClass, "StorageClass is used to create pvc for default tenant")

	flag.Parse()
}

func initialize(opts *APIServerOptions) {
	// Init k8s client
	log.Info("kube config:", opts.KubeConfig)
	client, err := common.GetClient(opts.KubeHost, opts.KubeConfig)
	if err != nil {
		log.Fatalf("Create k8s client error: %v", err)
	}

	hcommon.InitHandlers(client)
	log.Info("Init k8s client success")

	err = v1alpha1.CreateDefaultTenant()
	if err != nil {
		log.Fatalf("Create default tenant cyclone error %v", err)
	}
	return
}

func main() {
	opts := NewAPIServerOptions()
	opts.AddFlags()

	initialize(opts)

	// Print nirvana banner.
	log.Infoln(nirvana.Logo, nirvana.Banner)

	// Create nirvana command.
	cmd := nconfig.NewNamedNirvanaCommand("cyclone-server", &nconfig.Option{
		IP:   opts.CycloneAddr,
		Port: uint16(opts.CyclonePort),
	})

	// add flags
	cmd.Add(&opts.KubeHost, config.FlagKubeHost, "", "Kube host address")
	cmd.Add(&opts.KubeConfig, config.FlagKubeConfig, "", "Kube config file path")
	cmd.Add(&opts.CyclonePort, config.FlagCycloneServerPort, "", "The port for the cyclone server to serve on")
	cmd.Add(&opts.CycloneAddr, config.FlagCycloneServerHost, "", "The IP address for the cyclone server to serve on")
	cmd.Add(&opts.Loglevel, config.FlagLogLevel, "", "Log level")
	cmd.Add(&opts.StorageClass, config.FlagStrorageClass, "", "StorageClass is used to create pvc for default tenant")

	// Create plugin options.
	metricsOption := metrics.NewDefaultOption() // Metrics plugin.
	loggerOption := logger.NewDefaultOption()   // Logger plugin.
	reqlogOption := reqlog.NewDefaultOption()   // Request log plugin.
	versionOption := pversion.NewOption(        // Version plugin.
		"server",
		version.Version,
		version.Commit,
		version.Package,
	)

	// Enable plugins.
	cmd.EnablePlugin(metricsOption, loggerOption, reqlogOption, versionOption)

	// Create server config.
	serverConfig := nirvana.NewConfig()

	// Configure APIs. These configurations may be changed by plugins.
	serverConfig.Configure(
		nirvana.Logger(log.DefaultLogger()), // Will be changed by logger plugin.
		nirvana.Filter(filters.Filters()...),
		nirvana.Modifier(modifiers.Modifiers()...),
		nirvana.Descriptor(apis.Descriptor()),
	)

	// Set nirvana command hooks.
	cmd.SetHook(&nconfig.NirvanaCommandHookFunc{
		PreServeFunc: func(config *nirvana.Config, server nirvana.Server) error {
			// Output project information.
			config.Logger().Infof("Package:%s Version:%s Commit:%s", version.Package, version.Version, version.Commit)
			return nil
		},
	})

	log.Infof("Cyclone service listening on %s:%d", opts.CycloneAddr, opts.CyclonePort)

	// Start with server config.
	if err := cmd.ExecuteWithConfig(serverConfig); err != nil {
		serverConfig.Logger().Fatal(err)
	}

	log.Info("Cyclone server stopped")
}
