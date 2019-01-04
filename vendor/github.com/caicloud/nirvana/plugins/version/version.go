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

package version

import (
	"context"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

func init() {
	nirvana.RegisterConfigInstaller(&versionInstaller{})
}

// ExternalConfigName is the external config name of version.
const ExternalConfigName = "version"

// config is version config.
type config struct {
	path        string
	name        string
	version     string
	hash        string
	description string
}

type version struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Hash        string `json:"hash,omitempty"`
	Description string `json:"description,omitempty"`
}

type versionInstaller struct{}

// Name is the external config name.
func (i *versionInstaller) Name() string {
	return ExternalConfigName
}

// Install installs stuffs before server starting.
func (i *versionInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		v := &version{
			Name:        c.name,
			Version:     c.version,
			Hash:        c.hash,
			Description: c.description,
		}
		err = builder.AddDescriptor(definition.Descriptor{
			Path:     c.path,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEAll},
			Definitions: []definition.Definition{{
				Method:  definition.Get,
				Results: definition.DataErrorResults(""),
				Function: func(ctx context.Context) (*version, error) {
					return v, nil
				},
			}},
		})
	})
	return err
}

// Uninstall uninstalls stuffs after server terminating.
func (i *versionInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	return nil
}

// Disable returns a configurer to disable version.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}

// Path returns a configurer to set version path.
func Path(path string) nirvana.Configurer {
	if path == "" {
		path = "/version"
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.path = path
		})
		return nil
	}
}

// Name Configurer sets project name.
func Name(name string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.name = name
		})
		return nil
	}
}

// Description Configurer sets project description.
func Description(description string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.description = description
		})
		return nil
	}
}

// Version Configurer sets version number.
func Version(version string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.version = version
		})
		return nil
	}
}

// Hash Configurer sets code hash.
func Hash(hash string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.hash = hash
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
			path: "/version",
		}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// Option contains basic configurations of version.
type Option struct {
	Path        string `desc:"Version path"`
	name        string
	version     string
	hash        string
	description string
}

// NewOption creates default option.
func NewOption(name string, version string, hash string, description string) *Option {
	return &Option{
		name:        name,
		version:     version,
		hash:        hash,
		description: description,
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
		Name(p.name),
		Version(p.version),
		Hash(p.hash),
		Description(p.description),
	)
	return nil
}
