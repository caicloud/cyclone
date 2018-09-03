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

package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"path"
	"strings"
	"text/template"

	"github.com/caicloud/nirvana/utils/api"
	"github.com/caicloud/nirvana/utils/generators/utils"
	"github.com/caicloud/nirvana/utils/project"
)

// Generator is for generating golang client.
type Generator struct {
	config  *project.Config
	apis    *api.Definitions
	rest    string
	pkg     string
	rootPkg string
}

// NewGenerator creates a golang client generator.
func NewGenerator(
	config *project.Config,
	apis *api.Definitions,
	rest string,
	pkg string,
	rootPkg string,
) *Generator {
	return &Generator{
		config:  config,
		apis:    apis,
		rest:    rest,
		pkg:     pkg,
		rootPkg: rootPkg,
	}
}

// Generate generate files
func (g *Generator) Generate() (map[string][]byte, error) {
	definitions, err := utils.SplitDefinitions(g.apis, g.config)
	if err != nil {
		return nil, err
	}
	codes := make(map[string][]byte)
	versions := make([]string, 0, len(definitions))
	for version, defs := range definitions {
		versions = append(versions, version)
		helper, err := newHelper(g.rootPkg, defs)
		if err != nil {
			return nil, err
		}
		types, imports := helper.Types()
		typeCodes, err := g.typeCodes(version, types, imports)
		if err != nil {
			return nil, err
		}
		functions, imports := helper.Functions()
		functionCodes, err := g.functionCodes(version, functions, imports)
		if err != nil {
			return nil, err
		}
		codes[version+"/types"] = typeCodes
		codes[version+"/client"] = functionCodes
	}
	client, err := g.aggregationClientCode(versions)
	if err != nil {
		return nil, err
	}
	codes["client"] = client
	return codes, nil
}

func (g *Generator) typeCodes(version string, types []Type, imports []string) ([]byte, error) {
	data := bytes.NewBufferString(fmt.Sprintf("package %s\n", version))
	writeln := func(str string) {
		_, err := fmt.Fprintln(data, str)
		// Ignore this error.
		_ = err
	}

	writeln("import (")
	for _, pkg := range imports {
		writeln(pkg)
	}
	writeln(")")
	for _, typ := range types {
		writeln("")
		writeln(string(typ.Generate()))
	}
	return format.Source(data.Bytes())
}

func (g *Generator) functionCodes(version string, functions []function, imports []string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	template, err := template.New("codes").Parse(`
package {{ .Version }}

import (
	context "context"

	{{- range .Imports }}
	{{.}}
	{{- end }}

	rest "{{ .Rest }}"
)

// Interface describes {{ .Version }} client.
type Interface interface {
{{- range .Functions }}
{{ .Comments -}}
	{{ .Name }}(ctx context.Context{{ range .Parameters }},{{ .ProposedName }} {{ .Typ }}{{- end }}) (
	{{- range .Results }}{{ .ProposedName }} {{ .Typ }}, {{ end }}err error)
{{- end }}
}

// Client for version {{ .Version }}.
type Client struct {
	rest *rest.Client
}

// NewClient creates a new client.
func NewClient(cfg *rest.Config) (*Client, error) {
	client, err := rest.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

// MustNewClient creates a new client or panic if an error occurs.
func MustNewClient(cfg *rest.Config) *Client {
	client, err := NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return client
}

{{ range .Functions }}
{{ .Comments -}}
func (c *Client) {{ .Name }}(ctx context.Context{{ range .Parameters }},{{ .ProposedName }} {{ .Typ }}{{- end }}) (
	{{- range .Results }}{{ .ProposedName }} {{ .Typ }}, {{ end }}err error) {
	{{- range .Results }}
	{{- if ne .Creator "" }}
	{{ .ProposedName }} = {{ .Creator }}
    {{- end }}
    {{- end }}
	err = c.rest.Request("{{ .Method }}", {{ .Code }}, "{{ .Path }}").
	{{ range .Parameters }}
	{{ $param := .ProposedName }}
	{{ if not .Extensions }}
	{{ .Source }}("{{ .Name }}", {{ $param }}).
	{{ end }}
	{{ range .Extensions }}
	{{ .Source }}("{{ .Name }}", {{ $param }}.{{ .Key }}).
	{{ end }}
    {{ end }}

	{{ range .Results }}
	{{- if ne .Creator "" }}
	{{ .Destination }}({{ .ProposedName }}).
    {{- else }}
	{{ .Destination }}(&{{ .ProposedName }}).
    {{- end }}
    {{ end }}
	Do(ctx)
	return 
}
{{ end }}
		`)
	if err != nil {
		return nil, err
	}
	err = template.Execute(buf, map[string]interface{}{
		"Version":   version,
		"Rest":      g.rest,
		"Functions": functions,
		"Imports":   imports,
	})
	if err != nil {
		return nil, err
	}
	return format.Source(buf.Bytes())
}

type versionedPackage struct {
	Version  string
	Path     string
	Function string
}

func (g *Generator) aggregationClientCode(versions []string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	template, err := template.New("codes").Parse(`
package {{ .PackageName }}

import (
	{{ range .Pakcages }}
	{{ .Version }} "{{ .Path }}"
	{{ end }}

	rest "{{ .Rest }}"
)

// Interface describes a versioned client.
type Interface interface {
{{- range .Pakcages }}
	// {{ .Function }} returns {{ .Version }} client.
	{{ .Function }}() {{ .Version }}.Interface
{{- end }}
}

// Client contains versioned clients.
type Client struct {
	{{ range .Pakcages }}
	{{ .Version }} *{{ .Version }}.Client
	{{ end }}
}

// NewClient creates a new client.
func NewClient(cfg *rest.Config) (Interface, error) {
	c := &Client{}
	var err error
	{{ range .Pakcages }}
	c.{{ .Version }}, err =  {{ .Version }}.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	{{ end }}
	return c, nil
}

// MustNewClient creates a new client or panic if an error occurs.
func MustNewClient(cfg *rest.Config) Interface {
	return &Client{
	{{- range .Pakcages }}
	{{ .Version }}: {{ .Version }}.MustNewClient(cfg),
	{{- end }}
	}
}


{{ range .Pakcages }}
// {{ .Function }} returns a versioned client.
func (c *Client) {{ .Function }}() {{ .Version }}.Interface {
	return c.{{ .Version }}
}
{{ end }}
		`)
	if err != nil {
		return nil, err
	}
	packages := []versionedPackage{}
	for _, version := range versions {
		packages = append(packages, versionedPackage{
			Version:  version,
			Path:     path.Join(g.pkg, version),
			Function: strings.Title(version),
		})
	}
	err = template.Execute(buf, map[string]interface{}{
		"PackageName": path.Base(g.pkg),
		"Pakcages":    packages,
		"Rest":        g.rest,
	})
	if err != nil {
		return nil, err
	}
	return format.Source(buf.Bytes())
}
