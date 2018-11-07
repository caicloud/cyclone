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
	for groupPkgName := range g.groupVersions {
		gvInterfaces[groupPkgName] = c.Universe.Type(types.Name{Package: path.Join(g.outputPackage, groupPkgName), Name: "Interface"})
		gvNewFuncs[groupPkgName] = c.Universe.Function(types.Name{Package: path.Join(g.outputPackage, groupPkgName), Name: "New"})
	}
	m := map[string]interface{}{
		"cacheSharedIndexInformer":                  c.Universe.Type(cacheSharedIndexInformer),
		"groupVersions":                             g.groupVersions,
		"gvInterfaces":                              gvInterfaces,
		"gvNewFuncs":                                gvNewFuncs,
		"gvGoNames":                                 g.gvGoNames,
		"interfacesNewInformerFunc":                 c.Universe.Type(types.Name{Package: g.internalInterfacesPackage, Name: "NewInformerFunc"}),
		"interfacesTweakListOptionsFunc":            c.Universe.Type(types.Name{Package: g.internalInterfacesPackage, Name: "TweakListOptionsFunc"}),
		"informerFactoryInterface":                  c.Universe.Type(types.Name{Package: g.internalInterfacesPackage, Name: "SharedInformerFactory"}),
		"clientSetInterface":                        c.Universe.Type(types.Name{Package: g.clientSetPackage, Name: "Interface"}),
		"reflectType":                               c.Universe.Type(reflectType),
		"runtimeObject":                             c.Universe.Type(runtimeObject),
		"schemaGroupVersionResource":                c.Universe.Type(schemaGroupVersionResource),
		"syncMutex":                                 c.Universe.Type(syncMutex),
		"timeDuration":                              c.Universe.Type(timeDuration),
		"namespaceAll":                              c.Universe.Type(metav1NamespaceAll),
		"object":                                    c.Universe.Type(metav1Object),
		"informersShardInformerFactory":             c.Universe.Type(types.Name{Package: "k8s.io/client-go/informers", Name: "SharedInformerFactory"}),
		"informersNewFilteredSharedInformerFactory": c.Universe.Function(types.Name{Package: "k8s.io/client-go/informers", Name: "NewFilteredSharedInformerFactory"}),
	}

	sw.Do(sharedInformerFactoryStruct, m)
	sw.Do(sharedInformerFactoryInterface, m)

	return sw.Error()
}

var sharedInformerFactoryStruct = `
// SharedInformerOption defines the functional option type for SharedInformerFactory.
type SharedInformerOption func(*sharedInformerFactory) *sharedInformerFactory

type sharedInformerFactory struct {
	{{.informersShardInformerFactory|raw}}
	namespace string
	tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}
}

// WithTweakListOptions sets a custom filter on all listers of the configured SharedInformerFactory.
func WithTweakListOptions(tweakListOptions internalinterfaces.TweakListOptionsFunc) SharedInformerOption {
	return func(factory *sharedInformerFactory) *sharedInformerFactory {
		factory.tweakListOptions = tweakListOptions
		return factory
	}
}

// WithNamespace limits the SharedInformerFactory to the specified namespace.
func WithNamespace(namespace string) SharedInformerOption {
	return func(factory *sharedInformerFactory) *sharedInformerFactory {
		factory.namespace = namespace
		return factory
	}
}

// NewSharedInformerFactory constructs a new instance of sharedInformerFactory for all namespaces.
func NewSharedInformerFactory(client {{.clientSetInterface|raw}}, defaultResync {{.timeDuration|raw}}) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync)
}

// NewFilteredSharedInformerFactory constructs a new instance of sharedInformerFactory.
// Listers obtained via this SharedInformerFactory will be subject to the same filters
// as specified here.
// Deprecated: Please use NewSharedInformerFactoryWithOptions instead
func NewFilteredSharedInformerFactory(client {{.clientSetInterface|raw}}, defaultResync {{.timeDuration|raw}}, namespace string, tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}) SharedInformerFactory {
	return NewSharedInformerFactoryWithOptions(client, defaultResync, WithNamespace(namespace), WithTweakListOptions(tweakListOptions)) 
}

// NewSharedInformerFactoryWithOptions constructs a new instance of a SharedInformerFactory with additional options.
func NewSharedInformerFactoryWithOptions(client {{.clientSetInterface|raw}}, defaultResync {{.timeDuration|raw}}, options ...SharedInformerOption) SharedInformerFactory {
	factory := &sharedInformerFactory{
		namespace:        v1.NamespaceAll,
	}
	
	// Apply all options
	for _, opt := range options {
		factory = opt(factory)
	}

	factory.SharedInformerFactory = {{.informersNewFilteredSharedInformerFactory|raw}}(client, defaultResync, factory.namespace, factory.tweakListOptions)
	return factory
}

`

var sharedInformerFactoryInterface = `
// SharedInformerFactory provides shared informers for resources in all known
// API group versions.
type SharedInformerFactory interface {
	{{.informersShardInformerFactory|raw}}

	{{$gvInterfaces := .gvInterfaces}}
	{{$gvGoNames := .gvGoNames}}
	{{range $groupName, $group := .groupVersions}}{{index $gvGoNames $groupName}}() {{index $gvInterfaces $groupName|raw}}
	{{end}}
}

{{$gvNewFuncs := .gvNewFuncs}}
{{$gvGoNames := .gvGoNames}}
{{range $groupPkgName, $group := .groupVersions}}
func (f *sharedInformerFactory) {{index $gvGoNames $groupPkgName}}() {{index $gvInterfaces $groupPkgName|raw}} {
  return {{index $gvNewFuncs $groupPkgName|raw}}(f, f.namespace, f.tweakListOptions)
}
{{end}}
`
