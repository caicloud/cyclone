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

	"github.com/caicloud/nirvana"
	nconfig "github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/logger"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/reqlog"
	pversion "github.com/caicloud/nirvana/plugins/version"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/server/apis"
	"github.com/caicloud/cyclone/pkg/server/apis/filters"
	"github.com/caicloud/cyclone/pkg/server/apis/modifiers"
	"github.com/caicloud/cyclone/pkg/server/biz/templates"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/version"
	utilk8s "github.com/caicloud/cyclone/pkg/util/k8s"

	_ "github.com/caicloud/cyclone/pkg/server/biz/scm/bitbucket"
	_ "github.com/caicloud/cyclone/pkg/server/biz/scm/github"
	_ "github.com/caicloud/cyclone/pkg/server/biz/scm/gitlab"
	_ "github.com/caicloud/cyclone/pkg/server/biz/scm/gogs"
	_ "github.com/caicloud/cyclone/pkg/server/biz/scm/svn"
)

// Options contains all options(config) for cyclone server
type Options struct {
	// KubeConfig represents the path of Kube config file
	KubeConfig string

	// ConfigMap that configures for cyclone server, default value is 'cyclone-server-config'
	ConfigMap string
}

// NewOptions returns a new Options
func NewOptions() *Options {
	return &Options{}
}

// AddFlags adds flags to APIServerOptions.
func (opts *Options) AddFlags() {
	flag.StringVar(&opts.KubeConfig, "kubeconfig", "", "Kube config file path")
	flag.StringVar(&opts.ConfigMap, "configmap", "cyclone-server-config", "ConfigMap that configures for cyclone server")

	flag.Parse()
}

func initialize(opts *Options) {
	// Init k8s client
	log.Info("kube config:", opts.KubeConfig)
	client, err := utilk8s.GetClient(opts.KubeConfig)
	if err != nil {
		log.Fatalf("Create k8s client error: %v", err)
	}

	// Load configuration from ConfigMap.
	cm, err := client.CoreV1().ConfigMaps(common.GetSystemNamespace()).Get(opts.ConfigMap, meta_v1.GetOptions{})
	if err != nil {
		log.Fatalf("Get ConfigMap %s error: %s", opts.ConfigMap, err)
	}
	if err = config.LoadConfig(cm); err != nil {
		log.Fatalf("Load config from ConfigMap %s error: %s", opts.ConfigMap, err)
	}

	handler.Init(client)
	log.Info("Init handlers succeed.")

	if config.Config.InitDefaultTenant {
		err = v1alpha1.CreateDefaultTenant()
		if err != nil {
			log.Fatalf("Create default cyclone tenant error %v", err)
		}
	} else {
		log.Info("init_default_tenant is false, skip create default tenant")
	}

	if config.Config.CreateBuiltinTemplates {
		templates.InitStageTemplates(client, common.GetSystemNamespace(), "")
	} else {
		log.Info("create_builtin_templates is false, skip create built-in stage templates")
	}
}

func main() {
	// Print Cyclone ascii art logo
	log.Infoln(common.CycloneLogo)

	opts := NewOptions()
	opts.AddFlags()

	initialize(opts)

	// Create nirvana command.
	cmd := nconfig.NewNamedNirvanaCommand("cyclone-server", &nconfig.Option{
		IP:   config.Config.CycloneServerHost,
		Port: config.Config.CycloneServerPort,
	})

	// add flags
	cmd.Add(&opts.KubeConfig, "kubeconfig", "", "Kube config file path")
	cmd.Add(&opts.ConfigMap, "configmap", "", "ConfigMap that configures for cyclone server")

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
