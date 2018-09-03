/*
Copyright 2018 Caicloud Authors

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

package healthcheck

import (
	"context"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

func init() {
	nirvana.RegisterConfigInstaller(&healthcheckInstaller{})
}

// HealthChecker checks if current server is healthy.
type HealthChecker func(ctx context.Context) error

func defaultHealthChecker(ctx context.Context) error {
	return nil
}

// ExternalConfigName is the external config name of health check.
const ExternalConfigName = "healthcheck"

// config is healthcheck config.
type config struct {
	path    string
	checker HealthChecker
}

type healthcheckInstaller struct{}

// Name is the external config name.
func (i *healthcheckInstaller) Name() string {
	return ExternalConfigName
}

// Install installs stuffs before server starting.
func (i *healthcheckInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		err = builder.AddDescriptor(definition.Descriptor{
			Path:     c.path,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEAll},
			Definitions: []definition.Definition{{
				Method:   definition.Get,
				Results:  []definition.Result{definition.ErrorResult()},
				Function: c.checker,
			}},
		})
	})
	return err
}

// Uninstall uninstalls stuffs after server terminating.
func (i *healthcheckInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	return nil
}

// Disable returns a configurer to disable healthcheck.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}

// Path returns a configurer to set health check path.
func Path(path string) nirvana.Configurer {
	if path == "" {
		path = "/healthz"
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.path = path
		})
		return nil
	}
}

// Checker returns a configurer to set health checker.
func Checker(checker HealthChecker) nirvana.Configurer {
	if checker == nil {
		checker = defaultHealthChecker
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.checker = checker
		})
		return nil
	}
}

func wrapper(c *nirvana.Config, f func(c *config)) {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{
			path:    "/healthz",
			checker: defaultHealthChecker,
		}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// Option contains basic configurations of healthcheck.
type Option struct {
	Path    string `desc:"Health check path"`
	checker HealthChecker
}

// NewOption creates default option.
func NewOption(checker HealthChecker) *Option {
	return &Option{
		checker: checker,
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	cfg.Configure(
		Path(p.Path),
		Checker(p.checker),
	)
	return nil
}
