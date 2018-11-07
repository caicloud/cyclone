/*
Copyright 2016 The Kubernetes Authors.

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

package generators

import (
	"io"
	"path"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"

	"github.com/caicloud/cyclone/tools/generator/client-gen/generators/util"
	clientgentypes "github.com/caicloud/cyclone/tools/generator/client-gen/types"
)

// versionInterfaceGenerator generates the per-version interface file.
type versionInterfaceGenerator struct {
	generator.DefaultGen
	outputPackage             string
	imports                   namer.ImportTracker
	types                     []*types.Type
	groupPkgName              string
	groupGoName               string
	groupVersion              clientgentypes.GroupVersion
	clientSetPackage          string
	filtered                  bool
	internalInterfacesPackage string
}

var _ generator.Generator = &versionInterfaceGenerator{}

func (g *versionInterfaceGenerator) Filter(c *generator.Context, t *types.Type) bool {
	if !g.filtered {
		g.filtered = true
		return true
	}
	return false
}

func (g *versionInterfaceGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *versionInterfaceGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	return
}

func (g *versionInterfaceGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "{{", "}}")

	gvListerInterfaces := make([]*types.Type, len(g.types))

	for i, t := range g.types {
		gvListerInterfaces[i] = c.Universe.Type(types.Name{Package: path.Join("k8s.io/client-go/listers", g.groupPkgName, g.groupVersion.Version.String()), Name: t.Name.Name + "Lister"})

	}
	m := map[string]interface{}{
		"interfacesTweakListOptionsFunc": c.Universe.Type(types.Name{Package: g.internalInterfacesPackage, Name: "TweakListOptionsFunc"}),
		"types":                          g.types,
		"clientSetInterface":             c.Universe.Type(clientsetInterface),
		"sharedInformerFactoryInterface": c.Universe.Type(sharedInformerFactoryInterface),
		"gvListerInterfaces":             gvListerInterfaces,
		"group":                          namer.IC(g.groupGoName),
		"version":                        namer.IC(g.groupVersion.Version.String()),
	}

	sw.Do(versionTemplate, m)
	for i, typeDef := range g.types {
		tags, err := util.ParseClientGenTags(typeDef.SecondClosestCommentLines)
		if err != nil {
			return err
		}
		m["namespaced"] = !tags.NonNamespaced
		m["type"] = typeDef
		m["gvListerInterface"] = gvListerInterfaces[i]
		sw.Do(versionFuncTemplate, m)
	}

	return sw.Error()
}

var versionTemplate = `
// Interface provides access to all the listers in this group version.
type Interface interface {
	{{- $gvListerInterfaces := .gvListerInterfaces -}}
	{{range $index, $typ := .types -}}
		// {{$typ|publicPlural}} returns a {{$typ|public}}Lister
		{{$typ|publicPlural}}() {{index $gvListerInterfaces $index|raw}}
	{{end}}
}

type version struct {
	client {{.clientSetInterface|raw}}
	tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}
}

type infromerVersion struct {
	factory  {{.sharedInformerFactoryInterface|raw}}
}

// New returns a new Interface.
func New(client {{.clientSetInterface|raw}}, tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}) Interface {
	return &version{client: client, tweakListOptions: tweakListOptions}
}

// NewFrom returns a new Interface.
func NewFrom(factory {{.sharedInformerFactoryInterface|raw}}) Interface {
	return &infromerVersion{factory: factory}
}
`

var versionFuncTemplate = `
// {{.type|publicPlural}} returns a {{.type|public}}Lister.
func (v *version) {{.type|publicPlural}}() {{.gvListerInterface|raw}} {
	return &{{.type|private}}Lister{client: v.client, tweakListOptions: v.tweakListOptions}
}

// {{.type|publicPlural}} returns a {{.type|public}}Lister.
func (v *infromerVersion) {{.type|publicPlural}}() {{.gvListerInterface|raw}} {
	return v.factory.{{.group}}().{{.version}}().{{.type|publicPlural}}().Lister()
}
`
