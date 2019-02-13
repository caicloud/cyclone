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
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/version"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/server/biz/tenants"
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
	// KubeHost is Kube host address
	KubeHost string
	// KubeConfig represents the path of Kube config file
	KubeConfig string

	// ConfigMap that configures for cyclone server
	ConfigMap string
	// Namespace that cyclone server will run in
	Namespace string
}

// NewAPIServerOptions returns a new APIServerOptions
func NewAPIServerOptions() *APIServerOptions {
	return &APIServerOptions{}
}

// AddFlags adds flags to APIServerOptions.
func (opts *APIServerOptions) AddFlags() {
	flag.StringVar(&opts.KubeHost, "kubehost", "", "Kube host address")
	flag.StringVar(&opts.KubeConfig, "kubeconfig", "", "Kube config file path")
	flag.StringVar(&opts.ConfigMap, "configmap", "cyclone-server-config", "ConfigMap that configures for cyclone server")
	flag.StringVar(&opts.Namespace, "namespace", "default", "Namespace that cyclone server will run in")

	flag.Parse()
}

func initialize(opts *APIServerOptions) {
	// Init k8s client
	log.Info("kube config:", opts.KubeConfig)
	client, err := common.GetClient(opts.KubeHost, opts.KubeConfig)
	if err != nil {
		log.Fatalf("Create k8s client error: %v", err)
	}

	// Load configuration from ConfigMap.
	cm, err := client.CoreV1().ConfigMaps(opts.Namespace).Get(opts.ConfigMap, meta_v1.GetOptions{})
	if err != nil {
		log.Fatalf("Get ConfigMap %s error: %s", opts.ConfigMap, err)
	}
	if err = config.LoadConfig(cm); err != nil {
		log.Fatalf("Load config from ConfigMap %s error: %s", opts.ConfigMap, err)
	}

	handler.InitHandlers(client)
	log.Info("Init k8s client success")

	err = v1alpha1.CreateAdminTenant()
	if err != nil {
		log.Fatalf("Create default tenant cyclone error %v", err)
	}
	tenants.InitStageTemplates("")
}

func main() {
	// Print Cyclone ascii art logo
	log.Infoln(common.CycloneLogo)

	opts := NewAPIServerOptions()
	opts.AddFlags()

	initialize(opts)

	// Create nirvana command.
	cmd := nconfig.NewNamedNirvanaCommand("cyclone-server", &nconfig.Option{
		IP:   config.Config.CycloneServerHost,
		Port: config.Config.CycloneServerPort,
	})

	// add flags
	cmd.Add(&opts.KubeHost, "kubehost", "", "Kube host address")
	cmd.Add(&opts.KubeConfig, "kubeconfig", "", "Kube config file path")
	cmd.Add(&opts.ConfigMap, "configmap", "", "ConfigMap that configures for cyclone server")
	cmd.Add(&opts.Namespace, "namespace", "", "Namespace that cyclone server will run in")

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

	log.Infof("Cyclone service listening on %s:%d", config.Config.CycloneServerHost, config.Config.CycloneServerPort)

	// Start with server config.
	if err := cmd.ExecuteWithConfig(serverConfig); err != nil {
		serverConfig.Logger().Fatal(err)
	}

	log.Info("Cyclone server stopped")
}
