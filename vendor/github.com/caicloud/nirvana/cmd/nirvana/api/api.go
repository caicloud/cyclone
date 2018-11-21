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
	"github.com/caicloud/nirvana/cmd/nirvana/utils"
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
	builder := utils.NewAPIBuilder(o.findAllChildernPaths(args...)...)
	definitions, err := builder.Build()
	if err != nil {
		return err
	}
	file := ""
	for _, path := range args {
		file, err = o.findProjectFile(path)
		if err == nil {
			break
		}
	}
	config := &swagger.Config{}
	if file == "" {
		log.Warning("can't find nirvana.yaml, use empty config as instead")
	} else {
		config, err = swagger.LoadConfig(file)
		if err != nil {
			return err
		}
	}
	generator := swagger.NewDefaultGenerator(config, definitions)
	swaggers, err := generator.Generate()
	if err != nil {
		return err
	}

	files := map[string][]byte{}
	for _, s := range swaggers {
		data, err := json.Marshal(s)
		if err != nil {
			return err
		}
		files[s.Info.Version] = data
	}

	if o.Output != "" {
		err = o.write(files)
	}

	if o.Serve != "" {
		err = o.serve(files)
	}
	return err
}

// findAllChildernPaths walkthroughs all child directories but ignore vendors.
func (o *apiOptions) findAllChildernPaths(paths ...string) []string {
	walked := map[string]bool{}
	goDir := map[string]bool{}
	for _, path := range paths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				if info.Name() == "vendor" || walked[path] {
					return filepath.SkipDir
				}
				walked[path] = true
				return nil
			}
			if strings.HasSuffix(path, ".go") {
				dir := filepath.Dir(path)
				goDir[dir] = true
			}
			return nil
		})
		_ = err
	}
	results := []string{}
	for path := range goDir {
		results = append(results, path)
	}
	return results
}

// findProjectFile find the path of nirvana.yaml.
// It will find the path itself and its parents recursively.
func (o *apiOptions) findProjectFile(path string) (string, error) {
	goPath, absPath, err := utils.GoPath(path)
	if err != nil {
		return "", err
	}
	fileName := "nirvana.yaml"
	for len(absPath) > len(goPath) {
		path = filepath.Join(absPath, fileName)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path, nil
		}
		absPath = filepath.Dir(absPath)
	}
	return "", fmt.Errorf("can't find nirvana.yaml")
}

func (o *apiOptions) write(apis map[string][]byte) error {
	dir := o.Output
	for version, data := range apis {
		file := filepath.Join(dir, o.pathForVersion(version))
		if err := ioutil.WriteFile(file, data, 0664); err != nil {
			return err
		}
	}
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
{{ $total := len .}}
{{ $multipleVersions := gt $total 1 }}
<!DOCTYPE html>
<html>
  <head>
    <title>ReDoc Demo: Multiple apis</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
      body {
        margin: 0;
        {{ if $multipleVersions -}}
        padding: 50px 0 0 0;
        {{ end -}}
      }
      {{ if $multipleVersions -}}
      .topnav {
        position: fixed;
        top: 0px;
        width: 100%;
        height: 50px;
        box-sizing: border-box;
        z-index: 10;
        display: flex;
        -webkit-box-align: center;
        align-items: center;
        font-family: Lato;
        background: white;
        border-bottom: 1px solid rgb(204, 204, 204);
        padding: 5px;
      }
      .topnav-item {
        float: left;
        display: block;
        font-size: 15px;
        line-height: 1.5;
        height: 34px;
        text-align: center;
        padding-top: 15px;
        padding-left: 25px;
        padding-right: 25px;
        margin-right: 1px;
        color: #555555;
        background-color: #fafafa;
        cursor: pointer;
      }
      .topnav-img {
        height: 40px;
        width: 124px;
        display: inline-block;
        margin-right: 90px;
      }
      {{ end -}}
    </style>
  </head>
  <body>
    {{ if $multipleVersions -}}
    <!-- Top navigation placeholder -->
    <nav class="topnav">
      <img class="topnav-img" src="https://github.com/Rebilly/ReDoc/raw/master/docs/images/redoc-logo.png">
      <ul id="links_container">
      </ul>
    </nav>
	{{ end }}
    <redoc scroll-y-offset="body > nav"></redoc>

    <script src="https://rebilly.github.io/ReDoc/releases/v1.x.x/redoc.min.js"> </script>
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
      // initially render first API
      Redoc.init(apis[0].url, {
          suppressWarnings: true
      });
      {{ if $multipleVersions -}}
      function onClick() {
        var url = this.getAttribute('data-link');
        Redoc.init(url, {
          suppressWarnings: true
        });
      }
      // dynamically building navigation items
      var $list = document.getElementById('links_container');
      apis.forEach(function(api) {
        var $listitem = document.createElement('li');
        $listitem.setAttribute('data-link', api.url);
        $listitem.setAttribute('class', "topnav-item");
        $listitem.innerText = api.name;
        $listitem.addEventListener('click', onClick);
        $list.appendChild($listitem);
      });
	  {{ end }}
    </script>
  </body>
</html>
`
	tmpl, err := template.New("index.html").Parse(index)
	if err != nil {
		return nil, err
	}
	data := []struct {
		Name string
		Path string
	}{}
	for _, v := range versions {
		path := o.pathForVersion(v)
		data = append(data, struct {
			Name string
			Path string
		}{v, path})
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
