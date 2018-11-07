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
	"path/filepath"
	"strings"

	clientgentypes "github.com/caicloud/cyclone/tools/generator/client-gen/types"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// groupInterfaceGenerator generates the per-group interface file.
type groupInterfaceGenerator struct {
	generator.DefaultGen
	outputPackage             string
	imports                   namer.ImportTracker
	groupVersions             clientgentypes.GroupVersions
	clientSetPackage          string
	filtered                  bool
	internalInterfacesPackage string
}

var _ generator.Generator = &groupInterfaceGenerator{}

func (g *groupInterfaceGenerator) Filter(c *generator.Context, t *types.Type) bool {
	if !g.filtered {
		g.filtered = true
		return true
	}
	return false
}

func (g *groupInterfaceGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *groupInterfaceGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	return
}

type versionData struct {
	Name      string
	Interface *types.Type
	New       *types.Type
	NewFrom   *types.Type
}

func (g *groupInterfaceGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")

	versions := make([]versionData, 0, len(g.groupVersions.Versions))
	for _, version := range g.groupVersions.Versions {
		gv := clientgentypes.GroupVersion{Group: g.groupVersions.Group, Version: version.Version}
		versionPackage := filepath.Join(g.outputPackage, strings.ToLower(gv.Version.NonEmpty()))
		iface := c.Universe.Type(types.Name{Package: versionPackage, Name: "Interface"})
		versions = append(versions, versionData{
			Name:      namer.IC(version.Version.NonEmpty()),
			Interface: iface,
			New:       c.Universe.Function(types.Name{Package: versionPackage, Name: "New"}),
			NewFrom:   c.Universe.Function(types.Name{Package: versionPackage, Name: "NewFrom"}),
		})
	}
	m := map[string]interface{}{
		"interfacesTweakListOptionsFunc": c.Universe.Type(types.Name{Package: g.internalInterfacesPackage, Name: "TweakListOptionsFunc"}),
		"versions":                       versions,
		"clientSetInterface":             c.Universe.Type(clientsetInterface),
		"sharedInformerFactoryInterface": c.Universe.Type(sharedInformerFactoryInterface),
	}

	sw.Do(groupTemplate, m)

	return sw.Error()
}

var groupTemplate = `
// Interface provides access to each of this group's versions.
type Interface interface {
	$range .versions -$
		// $.Name$ provides access to listers for resources in $.Name$.
		$.Name$() $.Interface|raw$
	$end$
}

type group struct {
	client $.clientSetInterface|raw$
	tweakListOptions $.interfacesTweakListOptionsFunc|raw$
}

type informerGroup struct {
	factory  $.sharedInformerFactoryInterface|raw$
}

// New returns a new Interface.
func New(client $.clientSetInterface|raw$, tweakListOptions $.interfacesTweakListOptionsFunc|raw$) Interface {
	return &group{client: client, tweakListOptions: tweakListOptions}
}

// NewFrom returns a new Interface
func NewFrom(factory $.sharedInformerFactoryInterface|raw$) Interface {
	return &informerGroup{factory: factory}
}

$range .versions$
// $.Name$ returns a new $.Interface|raw$.
func (g *group) $.Name$() $.Interface|raw$ {
	return $.New|raw$(g.client, g.tweakListOptions)
}

func (g *informerGroup) $.Name$() $.Interface|raw$ {
	return $.NewFrom|raw$(g.factory)
}
$end$
`
