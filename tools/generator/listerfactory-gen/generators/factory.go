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

	clientgentypes "github.com/caicloud/cyclone/tools/generator/client-gen/types"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"

	"github.com/golang/glog"
)

// factoryGenerator produces a file of listers for a given GroupVersion and
// type.
type factoryGenerator struct {
	generator.DefaultGen
	outputPackage             string
	imports                   namer.ImportTracker
	groupVersions             map[string]clientgentypes.GroupVersions
	gvGoNames                 map[string]string
	clientSetPackage          string
	internalInterfacesPackage string
	filtered                  bool
}

var _ generator.Generator = &factoryGenerator{}

func (g *factoryGenerator) Filter(c *generator.Context, t *types.Type) bool {
	if !g.filtered {
		g.filtered = true
		return true
	}
	return false
}

func (g *factoryGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *factoryGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	return
}

func (g *factoryGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "{{", "}}")

	glog.V(5).Infof("processing type %v", t)

	gvInterfaces := make(map[string]*types.Type)
	gvNewFuncs := make(map[string]*types.Type)
	gvNewFromFuncs := make(map[string]*types.Type)
	for groupPkgName := range g.groupVersions {
		gvInterfaces[groupPkgName] = c.Universe.Type(types.Name{Package: path.Join(g.outputPackage, groupPkgName), Name: "Interface"})
		gvNewFuncs[groupPkgName] = c.Universe.Function(types.Name{Package: path.Join(g.outputPackage, groupPkgName), Name: "New"})
		gvNewFromFuncs[groupPkgName] = c.Universe.Function(types.Name{Package: path.Join(g.outputPackage, groupPkgName), Name: "NewFrom"})
	}
	m := map[string]interface{}{
		"cacheSharedIndexInformer":       c.Universe.Type(cacheSharedIndexInformer),
		"groupVersions":                  g.groupVersions,
		"gvInterfaces":                   gvInterfaces,
		"gvNewFuncs":                     gvNewFuncs,
		"gvNewFromFuncs":                 gvNewFromFuncs,
		"gvGoNames":                      g.gvGoNames,
		"interfacesTweakListOptionsFunc": c.Universe.Type(types.Name{Package: g.internalInterfacesPackage, Name: "TweakListOptionsFunc"}),
		"sharedInformerFactoryInterface": c.Universe.Type(sharedInformerFactoryInterface),
		"clientSetInterface":             c.Universe.Type(clientsetInterface),
		"reflectType":                    c.Universe.Type(reflectType),
		"runtimeObject":                  c.Universe.Type(runtimeObject),
		"schemaGroupVersionResource":     c.Universe.Type(schemaGroupVersionResource),
		"syncMutex":                      c.Universe.Type(syncMutex),
		"timeDuration":                   c.Universe.Type(timeDuration),
		"namespaceAll":                   c.Universe.Type(metav1NamespaceAll),
	}

	sw.Do(clientListerFactoryStruct, m)
	sw.Do(clientListerFactoryInterface, m)

	return sw.Error()
}

var clientListerFactoryStruct = `
type clientListerFactory struct {
	client {{.clientSetInterface|raw}}
	tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}
}

// NewListerFactoryFromClient constructs a new instance of clientListerFactory
func NewListerFactoryFromClient(client {{.clientSetInterface|raw}}) ListerFactory {
  return NewFilteredListerFactoryFromClient(client, nil)
}

// NewFilteredListerFactoryFromClient constructs a new instance of clientListerFactory.
// Listers obtained via this ClientListerFactory will be subject to the same filters
// as specified here.
func NewFilteredListerFactoryFromClient(client {{.clientSetInterface|raw}}, tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}) ListerFactory {
  return &clientListerFactory{
    client:           client,
	tweakListOptions: tweakListOptions,
  }
}

type informerListerFactory struct {
	factory  {{.sharedInformerFactoryInterface|raw}}
}

func NewListerFactoryFromInformer(factory {{.sharedInformerFactoryInterface|raw}}) ListerFactory {
	return &informerListerFactory {
		factory: factory,
	}
}

`

var clientListerFactoryInterface = `
// ListerFactory provides listers for resources in all known
// API group versions.
type ListerFactory interface {
	{{$gvInterfaces := .gvInterfaces}}
	{{$gvGoNames := .gvGoNames}}
	{{range $groupName, $group := .groupVersions}}{{index $gvGoNames $groupName}}() {{index $gvInterfaces $groupName|raw}}
	{{end}}
}

{{$gvNewFuncs := .gvNewFuncs}}
{{$gvNewFromFuncs := .gvNewFromFuncs}}
{{$gvGoNames := .gvGoNames}}
{{range $groupPkgName, $group := .groupVersions}}
func (f *clientListerFactory) {{index $gvGoNames $groupPkgName}}() {{index $gvInterfaces $groupPkgName|raw}} {
  return {{index $gvNewFuncs $groupPkgName|raw}}(f.client, f.tweakListOptions)
}
{{end}}

{{range $groupPkgName, $group := .groupVersions}}
func (f *informerListerFactory) {{index $gvGoNames $groupPkgName}}() {{index $gvInterfaces $groupPkgName|raw}} {
	return {{index $gvNewFromFuncs $groupPkgName|raw}}(f.factory)
}
{{end}}



`
