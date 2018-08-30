/*
Copyright 2017 Caicloud Authors

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

package profiling

import (
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
	rpprof "runtime/pprof"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

func init() {
	nirvana.RegisterConfigInstaller(&profilingInstaller{})
}

// ExternalConfigName is the external config name of profiling.
const ExternalConfigName = "profiling"

// config is profiling config.
type config struct {
	path string
}

type profilingInstaller struct{}

// Name is the external config name.
func (i *profilingInstaller) Name() string {
	return ExternalConfigName
}

// Install installs config to builder.
func (i *profilingInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		if err = builder.AddDescriptor(descriptor(c.path)); err != nil {
			return
		}
	})
	return err
}

// Uninstall uninstalls stuffs after server terminating.
func (i *profilingInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	return nil
}

// Disable returns a configurer to disable profiling.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}

func wrapper(c *nirvana.Config, f func(c *config)) {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{
			path: "/debug/pprof/",
		}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// Path returns a configurer to set metrics path.
// Default path is /debug/pprof.
// Then these path is used:
//   /debug/pprof/profile
//   /debug/pprof/symbol
//   /debug/pprof/trace
func Path(path string) nirvana.Configurer {
	if path == "" {
		path = "/debug/pprof/"
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.path = path
		})
		return nil
	}
}

// descriptor creates descriptor for profiling.
func descriptor(path string) definition.Descriptor {
	return definition.Descriptor{
		Path:     path,
		Consumes: []string{definition.MIMEAll},
		Produces: []string{definition.MIMEAll},
		Definitions: []definition.Definition{{
			Method:   definition.Get,
			Function: service.WrapHTTPHandlerFunc(index),
		}},
		Children: []definition.Descriptor{
			{
				Path: "/profile",
				Definitions: []definition.Definition{{
					Method:   definition.Get,
					Function: service.WrapHTTPHandlerFunc(pprof.Profile),
				}},
			},
			{
				Path: "/symbol",
				Definitions: []definition.Definition{{
					Method:   definition.Get,
					Function: service.WrapHTTPHandlerFunc(pprof.Symbol),
				}},
			},
			{
				Path: "/trace",
				Definitions: []definition.Definition{{
					Method:   definition.Get,
					Function: service.WrapHTTPHandlerFunc(pprof.Trace),
				}},
			},
		},
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	profiles := rpprof.Profiles()
	if err := indexTmpl.Execute(w, profiles); err != nil {
		log.Print(err)
	}
}

// indexTmpl is modified from http/pprof/pprof.go, adding 'pprof/' prefix for all href content
// go pprof http index page served the path '/debug/pprof/' which will be redirected to '/debug/pprof'
// by nirvana
var indexTmpl = template.Must(template.New("index").Parse(`<html>
	<head>
	<title>Profiling</title>
	</head>
	<body>
	Profiling<br>
	<br>
	profiles:<br>
	<table>
	{{range .}}
	<tr><td align=right>{{.Count}}<td><a href="./{{.Name}}?debug=1">{{.Name}}</a>
	{{end}}
	</table>
	<br>
	<a href=",/goroutine?debug=2">full goroutine stack dump</a><br>
	</body>
	</html>
	`))

// Option contains basic configurations of profiling.
type Option struct {
	// Path is profiling path.
	Path string `desc:"Profiling path"`
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		Path: "/debug/pprof/",
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
	)
	return nil
}
