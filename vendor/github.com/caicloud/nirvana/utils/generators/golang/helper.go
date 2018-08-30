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
	"fmt"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/api"
)

// Type abstracts common ability from type declarations.
type Type interface {
	// Name returns type name.
	Name() string
	// Generate generates type codes.
	Generate() []byte
}

type basicType struct {
	name     string
	comments string
	target   string
}

func (t *basicType) Name() string {
	return t.name
}

func (t *basicType) Generate() []byte {
	buf := api.NewBuffer()
	buf.Write(t.comments)
	buf.Writef("type %s %s\n", t.name, t.target)
	return buf.Bytes()
}

type arrayType struct {
	name     string
	comments string
	elem     string
}

func (t *arrayType) Name() string {
	return t.name
}

func (t *arrayType) Generate() []byte {
	buf := api.NewBuffer()
	buf.Write(t.comments)
	buf.Writef("type %s []%s\n", t.name, t.elem)
	return buf.Bytes()
}

type pointerType struct {
	name     string
	comments string
	elem     string
}

func (t *pointerType) Name() string {
	return t.name
}

func (t *pointerType) Generate() []byte {
	buf := api.NewBuffer()
	buf.Write(t.comments)
	buf.Writef("type %s *%s\n", t.name, t.elem)
	return buf.Bytes()
}

type mapType struct {
	name     string
	comments string
	key      string
	elem     string
}

func (t *mapType) Name() string {
	return t.name
}

func (t *mapType) Generate() []byte {
	buf := api.NewBuffer()
	buf.Write(t.comments)
	buf.Writef("type %s map[%s]%s\n", t.name, t.key, t.elem)
	return buf.Bytes()
}

type structField struct {
	name     string
	comments string
	typ      string
	tag      string
}

type structType struct {
	name     string
	comments string
	fields   []structField
}

func (t *structType) Name() string {
	return t.name
}

func (t *structType) Generate() []byte {
	buf := api.NewBuffer()
	buf.Write(t.comments)
	buf.Writef("type %s struct {\n", t.name)
	for _, field := range t.fields {
		buf.Write(field.comments)
		buf.Writef("%s %s", field.name, field.typ)
		if field.tag != "" {
			buf.Write(" `" + field.tag + "`")
		}
		buf.Writeln()
	}
	buf.Write("}\n")
	return buf.Bytes()
}

type parameterExtension struct {
	Source string
	Name   string
	Key    string
}

type functionParameter struct {
	Source       string
	Name         string
	ProposedName string
	Typ          string
	Extensions   []parameterExtension
}

type functionResult struct {
	Destination  string
	ProposedName string
	Typ          string
	Creator      string
}

type function struct {
	Path       string
	Method     string
	Code       int
	Name       string
	Comments   string
	Parameters []functionParameter
	Results    []functionResult
}

// helper provides methods to help to generate codes.
type helper struct {
	definitions *api.Definitions
	namer       *typeNamer
	rootPkg     string
}

// newHelper creates a generator helper.
func newHelper(rootPkg string, definitions *api.Definitions) (*helper, error) {
	namer, err := newTypeNamer(rootPkg, definitions.Types)
	if err != nil {
		return nil, err
	}
	return &helper{definitions, namer, rootPkg}, nil
}

// Types returns types which is required to generate.
func (h *helper) Types() ([]Type, []string) {
	types := []Type{}
	generatedTypes := []*api.Type{}

	for name, typ := range h.definitions.Types {
		if typ.Kind == reflect.Func || typ.PkgPath == "" ||
			strings.Contains(typ.PkgPath, "/vendor/") ||
			!(typ.PkgPath == h.rootPkg || strings.HasPrefix(typ.PkgPath, h.rootPkg+"/")) {
			// Ignore functions.
			// Ignore unnamed types.
			// Ignore vendored types.
			// Ignore types which are from standard and third-party libraries.
			continue
		}
		target := h.namer.Name(name)
		switch typ.Kind {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String:
			types = append(types, &basicType{
				name:     target,
				comments: h.namer.Comments(name),
				target:   typ.Kind.String(),
			})
		case reflect.Array, reflect.Slice:
			generatedTypes = append(generatedTypes, typ)
			types = append(types, &arrayType{
				name:     target,
				comments: h.namer.Comments(name),
				elem:     h.namer.Name(typ.Elem),
			})
		case reflect.Ptr:
			generatedTypes = append(generatedTypes, typ)
			types = append(types, &pointerType{
				name:     target,
				comments: h.namer.Comments(name),
				elem:     h.namer.Name(typ.Elem),
			})
		case reflect.Map:
			generatedTypes = append(generatedTypes, typ)
			types = append(types, &mapType{
				name:     target,
				comments: h.namer.Comments(name),
				key:      h.namer.Name(typ.Key),
				elem:     h.namer.Name(typ.Elem),
			})
		case reflect.Struct:
			generatedTypes = append(generatedTypes, typ)
			s := &structType{
				name:     target,
				comments: h.namer.Comments(name),
			}
			types = append(types, s)
			for _, field := range typ.Fields {
				c := api.ParseComments(field.Comments)
				// Ignore field options.
				c.CleanOptions()
				sf := structField{
					comments: c.LineComments(),
					typ:      h.namer.Name(field.Type),
					tag:      string(field.Tag),
				}
				if !field.Anonymous {
					sf.name = field.Name
				}
				s.fields = append(s.fields, sf)
			}
		}
	}
	sort.Sort(typeSorter(types))
	return types, h.packages(generatedTypes, true)
}

// packages generates a list of imported packages with aliases.
func (h *helper) packages(types []*api.Type, extended bool) []string {
	pkgMap := map[string]bool{}
	for _, typ := range types {
		pkgs := h.pkgs(typ, extended)
		for _, pkg := range pkgs {
			pkgMap[pkg] = true
		}
	}
	results := []string{}
	for pkg := range pkgMap {
		alias := h.namer.Alias(pkg)
		results = append(results, fmt.Sprintf(`%s "%s"`, alias, pkg))
	}
	return results
}

// pkgs generates a list of imported packages without aliases.
func (h *helper) pkgs(typ *api.Type, extended bool) []string {
	switch typ.Kind {
	case reflect.Array, reflect.Slice, reflect.Ptr:
		return h.pkgs(h.definitions.Types[typ.Elem], extended)
	case reflect.Map:
		pkgs := h.pkgs(h.definitions.Types[typ.Key], extended)
		return append(pkgs, h.pkgs(h.definitions.Types[typ.Elem], extended)...)
	}

	if typ.PkgPath != "" {
		index := strings.LastIndex(typ.PkgPath, "/vendor/")
		if index >= 0 ||
			!(typ.PkgPath == h.rootPkg || strings.HasPrefix(typ.PkgPath, h.rootPkg+"/")) {
			pkg := typ.PkgPath
			if index > 0 {
				pkg = pkg[index+len("/vendor/"):]
			}
			return []string{pkg}
		}
		if extended && typ.Kind == reflect.Struct {
			pkgs := []string{}
			for _, field := range typ.Fields {
				pkgs = append(pkgs, h.pkgs(h.definitions.Types[field.Type], extended)...)
			}
			return pkgs
		}
	}
	return []string{}
}

// Functions returns functions which is required to generate.
func (h *helper) Functions() ([]function, []string) {
	functionNames := map[string]int{}
	functions := []function{}
	types := []*api.Type{}
	for path, defs := range h.definitions.Definitions {
		for _, def := range defs {
			fn := function{
				Path:   path,
				Method: def.HTTPMethod,
				Code:   def.HTTPCode,
			}
			// The priority of summary is higher than original function name.
			if def.Summary != "" {
				// Remove invalid chars and regard as function name.
				fn.Name = nameReplacer.ReplaceAllString(def.Summary, "")
			}

			if fn.Name == "" {
				// If original function is public and there is no summary,
				// original function name is selected.
				name := strings.TrimSpace(h.namer.Name(def.Function))
				if name != "" && name[0] >= 'A' && name[0] <= 'Z' {
					fn.Name = name
				}
			}

			if fn.Name == "" {
				// Anonymous function.
				fn.Name = "AnonymousAPI"
			}

			count := functionNames[fn.Name]
			functionNames[fn.Name]++
			if count > 0 {
				fn.Name += strconv.Itoa(count)
			}

			comments := fmt.Sprintf("%s ", fn.Name)
			if def.Description != "" {
				comments += fmt.Sprintf("description:\n%s", def.Description)
			} else {
				comments += "does not have any description."
			}
			fn.Comments = api.ParseComments(comments).LineComments()
			sigNames := h.namer.nameContainer()

			// If there is no specified consumer, defaults to application/json.
			firstNonEmptyConsume := definition.MIMEJSON
			for _, consume := range def.Consumes {
				if consume != "" {
					firstNonEmptyConsume = consume
					break
				}
			}

			for _, param := range def.Parameters {
				if param.Source == definition.Prefab {
					// Ignore prefabs.
					continue
				}
				p := functionParameter{
					Source:       string(param.Source),
					Name:         param.Name,
					ProposedName: sigNames.proposeName(param.Name, param.Type),
					Typ:          h.namer.Name(param.Type),
				}
				types = append(types, h.definitions.Types[param.Type])
				if param.Source == definition.Body {
					// Use first consumer as the name of body parameter.
					p.Name = firstNonEmptyConsume
				}
				if param.Source == definition.Auto {
					// Generate field extensions for auto struct.
					h.enumFields(param.Type, "",
						func(key string, tag string, field api.StructField) {
							source, name, _, err := service.ParseAutoParameterTag(tag)
							if err != nil {
								// Ignore invalid source tag.
								return
							}
							extension := parameterExtension{
								Source: string(source),
								Name:   name,
								Key:    key,
							}
							if source == definition.Body {
								// Use first consumer as the name of body parameter.
								extension.Name = firstNonEmptyConsume
							}
							p.Extensions = append(p.Extensions, extension)
						})
				}
				fn.Parameters = append(fn.Parameters, p)
			}
			for _, result := range def.Results {
				if result.Destination == definition.Error {
					// Ignore errors
					continue
				}
				r := functionResult{
					Destination:  string(result.Destination),
					ProposedName: sigNames.proposeName("", result.Type),
					Typ:          h.namer.Name(result.Type),
				}
				typ := h.definitions.Types[result.Type]
				types = append(types, typ)
				if typ.Kind == reflect.Ptr {
					r.Creator = fmt.Sprintf("new(%s)", h.namer.Name(typ.Elem))
				}
				fn.Results = append(fn.Results, r)
			}
			functions = append(functions, fn)
		}
	}
	sort.Sort(functionSorter(functions))
	return functions, h.packages(types, false)
}

func (h *helper) enumFields(name api.TypeName, key string, fn func(key string, source string, field api.StructField)) {
	typ := h.definitions.Types[name]
	if typ.Kind == reflect.Struct {
		if key != "" {
			key += "."
		}
		for _, field := range typ.Fields {
			source := field.Tag.Get("source")
			fieldKey := key + field.Name
			if source != "" {
				fn(fieldKey, source, field)
			} else {
				fieldType := h.definitions.Types[field.Type]
				if fieldType.Kind == reflect.Struct {
					h.enumFields(field.Type, key, fn)
				}
			}
		}
	}
}

type nameContainer struct {
	names map[string]int
	namer *typeNamer
}

var nameReplacer = regexp.MustCompile(`[^a-zA-Z0-9]`)

func (n *nameContainer) proposeName(name string, typ api.TypeName) string {
	if name == "" {
		name = n.deconstruct(typ)
	}
	name = nameReplacer.ReplaceAllString(name, "")
	if name == "" {
		name = "temp"
	}
	if name[0] >= 'A' && name[0] <= 'Z' {
		name = string(name[0]|0x20) + name[1:]
	}
	index := n.names[name]
	if index > 0 {
		name += strconv.Itoa(index)
	}
	n.names[name]++
	return name
}

func (n *nameContainer) deconstruct(name api.TypeName) string {
	typ := n.namer.types[name]
	switch typ.Kind {
	case reflect.Ptr:
		return n.deconstruct(typ.Elem)
	case reflect.Array, reflect.Slice, reflect.Map:
		result := n.deconstruct(typ.Elem)
		// Unsafe to convert result to plural form.
		switch {
		case strings.HasSuffix(result, "es"):
		case strings.HasSuffix(result, "y"):
			result = result[:len(result)-1] + "ies"
		case strings.HasSuffix(result, "s"):
			result = result[:len(result)-1] + "es"
		default:
			result += "s"
		}
		return result
	default:
		return n.namer.Name(name)
	}
}

type typeNamer struct {
	types    map[api.TypeName]*api.Type
	names    map[api.TypeName]string
	aliases  map[string]int
	packages map[string]string
	comments map[api.TypeName]string
	rootPkg  string
}

func newTypeNamer(rootPkg string, types map[api.TypeName]*api.Type) (*typeNamer, error) {
	n := &typeNamer{
		types:    types,
		names:    make(map[api.TypeName]string),
		aliases:  make(map[string]int),
		packages: make(map[string]string),
		comments: make(map[api.TypeName]string),
		rootPkg:  rootPkg,
	}
	for tn := range n.types {
		if _, err := n.parse(tn); err != nil {
			return nil, err
		}
	}
	return n, nil
}

func (n *typeNamer) nameContainer() *nameContainer {
	return &nameContainer{
		names: map[string]int{},
		namer: n,
	}
}

func (n *typeNamer) parse(tn api.TypeName) (string, error) {
	name, ok := n.names[tn]
	if ok {
		return name, nil
	}
	typ, ok := n.types[tn]
	if !ok {
		return "", fmt.Errorf("no type with name %s", tn)
	}
	comments := ""
	if typ.Kind == reflect.Interface ||
		typ.PkgPath != "" && (strings.Contains(typ.PkgPath, "/vendor/") ||
			!(typ.PkgPath == n.rootPkg || strings.HasPrefix(typ.PkgPath, n.rootPkg+"/"))) {
		// Interfaces, Builtin types, Third-party types, Standard types are imported rather than copied.
		alias := n.packages[typ.PkgPath]
		if alias == "" && typ.PkgPath != "" {
			alias = path.Base(typ.PkgPath)
			index := n.aliases[alias]
			n.aliases[alias]++
			if index > 0 {
				alias += strconv.Itoa(index)
			}
			n.packages[typ.PkgPath] = alias
		}
		name = typ.Name
		if alias != "" {
			name = fmt.Sprintf("%s.%s", alias, name)
		}
	} else {
		switch typ.Kind {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.String:
			if typ.PkgPath != "" {
				name, comments = n.reconcileNameAndComments(typ.Name, typ.Comments)
			} else {
				name = typ.Kind.String()
			}
		case reflect.Array, reflect.Slice:
			elemName, err := n.parse(typ.Elem)
			if err != nil {
				return "", err
			}
			if typ.PkgPath != "" {
				name, comments = n.reconcileNameAndComments(typ.Name, typ.Comments)
			} else {
				name = fmt.Sprintf("[]%s", elemName)
			}
		case reflect.Ptr:
			elemName, err := n.parse(typ.Elem)
			if err != nil {
				return "", err
			}
			if typ.PkgPath != "" {
				name, comments = n.reconcileNameAndComments(typ.Name, typ.Comments)
			} else {
				name = fmt.Sprintf("*%s", elemName)
			}
		case reflect.Map:
			keyName, err := n.parse(typ.Key)
			if err != nil {
				return "", err
			}
			elemName, err := n.parse(typ.Elem)
			if err != nil {
				return "", err
			}
			if typ.PkgPath != "" {
				name, comments = n.reconcileNameAndComments(typ.Name, typ.Comments)
			} else {
				name = fmt.Sprintf("map[%s]%s", keyName, elemName)
			}
		case reflect.Struct, reflect.Func:
			name, comments = n.reconcileNameAndComments(typ.Name, typ.Comments)
		default:
			return "", fmt.Errorf("can't generate a name for type %s", tn)
		}
	}
	n.names[tn] = name
	n.comments[tn] = comments
	return name, nil
}

func (n *typeNamer) reconcileNameAndComments(origin, comments string) (string, string) {
	c := api.ParseComments(comments)
	aliases := c.Option(api.CommentsOptionAlias)
	if len(aliases) > 0 {
		alias := aliases[0]
		if alias != "" && alias != origin {
			c.Rename(origin, alias)
			c.CleanOptions()
			c.AddOption(api.CommentsOptionOrigin, origin)
			origin = alias
		}
	}
	return origin, c.LineComments()
}

func (n *typeNamer) Name(tn api.TypeName) string {
	name, ok := n.names[tn]
	if !ok {
		panic(fmt.Errorf("can't find type %s", tn))
	}
	return name
}

func (n *typeNamer) Package(tn api.TypeName) (alias string, pkg string) {
	typ, ok := n.types[tn]
	if !ok {
		panic(fmt.Errorf("can't find type %s", tn))
	}
	alias, ok = n.packages[typ.PkgPath]
	if !ok {
		return "", ""
	}
	return alias, typ.PkgPath
}

func (n *typeNamer) Alias(pkg string) string {
	return n.packages[pkg]
}

func (n *typeNamer) Comments(tn api.TypeName) string {
	comments, ok := n.comments[tn]
	if !ok {
		panic(fmt.Errorf("can't find type %s", tn))
	}
	return comments
}

type typeSorter []Type

// Len is the number of elements in the collection.
func (s typeSorter) Len() int {
	return len(s)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s typeSorter) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

// Swap swaps the elements with indexes i and j.
func (s typeSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type functionSorter []function

// Len is the number of elements in the collection.
func (s functionSorter) Len() int {
	return len(s)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s functionSorter) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// Swap swaps the elements with indexes i and j.
func (s functionSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
