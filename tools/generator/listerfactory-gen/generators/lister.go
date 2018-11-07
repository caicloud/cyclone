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

	"github.com/golang/glog"
)

// listerGenerator produces a file of listers for a given GroupVersion and
// type.
type listerGenerator struct {
	generator.DefaultGen
	outputPackage             string
	groupPkgName              string
	groupVersion              clientgentypes.GroupVersion
	groupGoName               string
	typeToGenerate            *types.Type
	imports                   namer.ImportTracker
	clientSetPackage          string
	listersPackage            string
	internalInterfacesPackage string
	extraCtx                  *generator.Context
}

var _ generator.Generator = &listerGenerator{}

func (g *listerGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.typeToGenerate
}

func (g *listerGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *listerGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	return
}

func (g *listerGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "{{", "}}")

	glog.V(5).Infof("processing type %v", t)

	tags, err := util.ParseClientGenTags(append(t.SecondClosestCommentLines, t.CommentLines...))
	if err != nil {
		return err
	}

	listerPath := path.Join("k8s.io/client-go/listers", g.groupPkgName, g.groupVersion.Version.String())
	gvListerInterface := g.extraCtx.Universe.Type(types.Name{Package: listerPath, Name: g.typeToGenerate.Name.Name + "Lister"})
	gvListerExpansionInterface := g.extraCtx.Universe.Type(types.Name{Package: listerPath, Name: g.typeToGenerate.Name.Name + "ListerExpansion"})
	var gvNamespaceListerInterface *types.Type
	var gvNamespaceListerExpansionInterface *types.Type
	if !tags.NonNamespaced {
		gvNamespaceListerInterface = c.Universe.Type(types.Name{Package: listerPath, Name: g.typeToGenerate.Name.Name + "NamespaceLister"})
		gvNamespaceListerExpansionInterface = c.Universe.Type(types.Name{Package: listerPath, Name: g.typeToGenerate.Name.Name + "NamespaceListerExpansion"})
	}
	m := map[string]interface{}{
		"clientSetInterface":                  c.Universe.Type(clientsetInterface),
		"interfacesTweakListOptionsFunc":      c.Universe.Type(types.Name{Package: g.internalInterfacesPackage, Name: "TweakListOptionsFunc"}),
		"namespaceAll":                        c.Universe.Type(metav1NamespaceAll),
		"namespaced":                          !tags.NonNamespaced,
		"type":                                t,
		"v1ListOptions":                       c.Universe.Type(v1ListOptions),
		"v1GetOptions":                        c.Universe.Type(v1GetOptions),
		"group":                               namer.IC(g.groupGoName),
		"version":                             namer.IC(g.groupVersion.Version.String()),
		"gvListerInterface":                   gvListerInterface,
		"gvListerExpansionInterface":          gvListerExpansionInterface,
		"gvNamespaceListerInterface":          gvNamespaceListerInterface,
		"gvNamespaceListerExpansionInterface": gvNamespaceListerExpansionInterface,
	}

	generateExpansionInterface := func(methods map[string]*types.Type) {
		for name, method := range methods {
			m["methodName"] = name
			sw.Do("\n", m)
			sw.Do("func (s *{{.type|private}}Lister) {{.methodName}}(", m)
			for i, p := range method.Signature.Parameters {
				if i > 0 {
					sw.Do(",", m)
				}
				m["parameter"] = p
				sw.Do("{{.parameter|raw}}", m)
				delete(m, "parameter")
			}
			sw.Do(")(", m)
			for j, r := range method.Signature.Results {
				if j > 0 {
					sw.Do(",", m)
				}
				m["result"] = r
				if r.Kind == types.Interface {
					sw.Do("{{.result.Name}}", m)
				} else {
					sw.Do("{{.result|raw}}", m)
				}
				delete(m, "result")
			}
			sw.Do("){\n", m)
			sw.Do("return ", m)
			for j := range method.Signature.Results {
				if j > 0 {
					sw.Do(",", m)
				}
				sw.Do("nil", m)
			}
			sw.Do("\n", m)
			sw.Do("}\n", m)
		}
	}

	sw.Do(typeListerInterface, m)

	if !tags.NonNamespaced {
		sw.Do(namespaceListerInterface, m)
	}

	sw.Do(typeListerStruct, m)
	sw.Do(typeListerConstructor, m)
	sw.Do(typeLister_List, m)

	// generate funcs for ExpansionInterface
	generateExpansionInterface(gvListerExpansionInterface.Methods)

	if tags.NonNamespaced {
		sw.Do(typeLister_NonNamespacedGet, m)
		return sw.Error()
	}

	sw.Do(typeLister_NamespaceLister, m)
	sw.Do(namespaceListerStruct, m)
	sw.Do(namespaceLister_List, m)
	sw.Do(namespaceLister_Get, m)
	generateExpansionInterface(gvNamespaceListerExpansionInterface.Methods)

	return sw.Error()
}

var typeListerInterface = `
var _ {{.gvListerInterface|raw}} = &{{.type|private}}Lister{}
`

var namespaceListerInterface = `
var _ {{.gvNamespaceListerInterface|raw}} = &{{.type|private}}NamespaceLister{}
`

var typeListerStruct = `
// {{.type|private}}Lister implements the {{.type|public}}Lister interface.
type {{.type|private}}Lister struct {
	client {{.clientSetInterface|raw}}
	tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}
}
`

var typeListerConstructor = `
// New{{.type|public}}Lister returns a new {{.type|public}}Lister.
func New{{.type|public}}Lister(client {{.clientSetInterface|raw}}) {{.gvListerInterface|raw}} {
	return NewFiltered{{.type|public}}Lister(client, nil)
}

func NewFiltered{{.type|public}}Lister(client {{.clientSetInterface|raw}}, tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}) {{.gvListerInterface|raw}} {
	return &{{.type|private}}Lister{
		client: client,
		tweakListOptions: tweakListOptions,
	}
}
`

var typeLister_List = `
// List lists all {{.type|publicPlural}} in the indexer.
func (s *{{.type|private}}Lister) List(selector labels.Selector) (ret []*{{.type|raw}}, err error) {
	listopt := {{.v1ListOptions|raw}}{
		LabelSelector: selector.String(),
	}
	if s.tweakListOptions != nil {
		s.tweakListOptions(&listopt)
	}
	list, err := s.client.{{.group}}{{.version}}().{{.type|publicPlural}}({{if .namespaced}}{{.namespaceAll|raw}}{{end}}).List(listopt)
	if err != nil {
		return nil, err
	}
	for i := range list.Items{
		ret = append(ret, &list.Items[i])
	}
	return ret, nil
}
`

var typeLister_NamespaceLister = `
// {{.type|publicPlural}} returns an object that can list and get {{.type|publicPlural}}.
func (s *{{.type|private}}Lister) {{.type|publicPlural}}(namespace string) {{.gvNamespaceListerInterface|raw}} {
	return {{.type|private}}NamespaceLister{client: s.client, tweakListOptions: s.tweakListOptions, namespace: namespace}
}
`

var typeLister_NonNamespacedGet = `
// Get retrieves the {{.type|public}} from the index for a given name.
func (s *{{.type|private}}Lister) Get(name string) (*{{.type|raw}}, error) {
	return s.client.{{.group}}{{.version}}().{{.type|publicPlural}}().Get(name, {{.v1GetOptions|raw}}{})
}
`
var namespaceListerStruct = `
// {{.type|private}}NamespaceLister implements the {{.type|public}}NamespaceLister
// interface.
type {{.type|private}}NamespaceLister struct {
	client {{.clientSetInterface|raw}}
	namespace string
	tweakListOptions {{.interfacesTweakListOptionsFunc|raw}}
}
`

var namespaceLister_List = `
// List lists all {{.type|publicPlural}} in the indexer for a given namespace.
func (s {{.type|private}}NamespaceLister) List(selector labels.Selector) (ret []*{{.type|raw}}, err error) {
	listopt := {{.v1ListOptions|raw}}{
		LabelSelector: selector.String(),
	}
	if s.tweakListOptions != nil {
		s.tweakListOptions(&listopt)
	}
	list, err := s.client.{{.group}}{{.version}}().{{.type|publicPlural}}(s.namespace).List(listopt)
	if err != nil {
		return nil, err
	}
	for i := range list.Items{
		ret = append(ret, &list.Items[i])
	}
	return ret, nil
}
`

var namespaceLister_Get = `
// Get retrieves the {{.type|public}} from the indexer for a given namespace and name.
func (s {{.type|private}}NamespaceLister) Get(name string) (*{{.type|raw}}, error) {
	return s.client.{{.group}}{{.version}}().{{.type|publicPlural}}(s.namespace).Get(name, {{.v1GetOptions|raw}}{})
}
`
