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

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/cmd/nirvana/buildutils"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/generators/swagger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const mimeHTML = "text/html"

func init() {
	err := service.RegisterProducer(service.NewSimpleSerializer("text/html"))
	if err != nil {
		log.Fatalln(err)
	}
}

func newAPICommand() *cobra.Command {
	options := &apiOptions{}
	cmd := &cobra.Command{
		Use:   "api /path/to/apis",
		Short: "Generate API documents for your project",
		Long:  options.Manuals(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Validate(cmd, args); err != nil {
				log.Fatalln(err)
			}
			if err := options.Run(cmd, args); err != nil {
				log.Fatalln(err)
			}
		},
	}
	options.Install(cmd.PersistentFlags())
	return cmd
}

type apiOptions struct {
	Serve  string
	Output string
}

func (o *apiOptions) Install(flags *pflag.FlagSet) {
	flags.StringVar(&o.Serve, "serve", "127.0.0.1:8080", "Start a server to host api docs")
	flags.StringVar(&o.Output, "output", "", "Directory to output api specifications")
}

func (o *apiOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *apiOptions) Run(cmd *cobra.Command, args []string) error {
	if len(args) <= 0 {
		defaultAPIsPath := "pkg/apis"
		args = append(args, defaultAPIsPath)
		log.Infof("No packages are specified, defaults to %s", defaultAPIsPath)
	}

	config, definitions, err := buildutils.Build(args...)
	if err != nil {
		return err
	}

	log.Infof("Project root directory is %s", config.Root)

	generator := swagger.NewDefaultGenerator(config, definitions)
	swaggers, err := generator.Generate()
	if err != nil {
		return err
	}

	files := map[string][]byte{}
	for _, s := range swaggers {
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}

		files[s.Info.Version] = data
	}

	if o.Output != "" {
		if err = o.write(files); err != nil {
			return err
		}
	}

	if o.Serve != "" {
		err = o.serve(files)
	}
	return err
}

func (o *apiOptions) write(apis map[string][]byte) error {
	dir := o.Output
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0775); err != nil {
		return err
	}
	for version, data := range apis {
		file := filepath.Join(dir, o.pathForVersion(version))
		if err := ioutil.WriteFile(file, data, 0664); err != nil {
			return err
		}
	}
	log.Infof("Generated openapi schemes to %s", dir)
	return nil
}

func (o *apiOptions) serve(apis map[string][]byte) error {
	hosts := strings.Split(o.Serve, ":")
	ip := strings.TrimSpace(hosts[0])
	if ip == "" {
		ip = "127.0.0.1"
	}
	port := uint16(8080)
	if len(hosts) >= 2 {
		p := strings.TrimSpace(hosts[1])
		if p != "" {
			pt, err := strconv.Atoi(p)
			if err != nil {
				return err
			}
			port = uint16(pt)
		}
	}
	log.SetDefaultLogger(log.NewStdLogger(0))
	cfg := nirvana.NewDefaultConfig()
	versions := []string{}
	for v, data := range apis {
		versions = append(versions, v)
		cfg.Configure(nirvana.Descriptor(
			o.descriptorForData(o.pathForVersion(v), data, definition.MIMEJSON),
		))
	}
	data, err := o.indexData(versions)
	if err != nil {
		return err
	}
	cfg.Configure(
		nirvana.Descriptor(o.descriptorForData("/", data, mimeHTML)),
		nirvana.IP(ip),
		nirvana.Port(port),
	)
	log.Infof("Listening on %s:%d. Please open your browser to view api docs", cfg.IP(), cfg.Port())
	return nirvana.NewServer(cfg).Serve()
}

func (o *apiOptions) descriptorForData(path string, data []byte, ct string) definition.Descriptor {
	return definition.Descriptor{
		Path: path,
		Definitions: []definition.Definition{
			{
				Method:   definition.Get,
				Consumes: []string{definition.MIMENone},
				Produces: []string{ct},
				Function: func(context.Context) ([]byte, error) {
					return data, nil
				},
				Parameters: []definition.Parameter{},
				Results:    definition.DataErrorResults(""),
			},
		},
	}
}

func (o *apiOptions) pathForVersion(version string) string {
	return fmt.Sprintf("/api.%s.json", version)
}

func (o *apiOptions) indexData(versions []string) ([]byte, error) {
	index := `
<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" >
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@3.25.0/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@3.25.0/favicon-16x16.png" sizes="16x16" />
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }
      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }
      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"> </script>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-standalone-preset.js"> </script>
    <script>
	  // list of APIS
      var apis = [
        {{ range $i, $v := . }}
        {
            name: '{{ $v.Name }}',
            url: '{{ $v.Path }}'
        },
		{{ end }}
      ];
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        urls: apis,
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      // End Swagger UI call region
      window.ui = ui
    }
  </script>
  </body>
</html>
`
	tmpl, err := template.New("index.html").Parse(index)
	if err != nil {
		return nil, err
	}
	data := make([]struct {
		Name string
		Path string
	}, len(versions))
	for i, v := range versions {
		data[i].Name = v
		data[i].Path = o.pathForVersion(v)
	}
	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (o *apiOptions) Manuals() string {
	return ""
}
