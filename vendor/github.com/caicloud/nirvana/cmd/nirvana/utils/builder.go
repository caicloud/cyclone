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

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/caicloud/nirvana/utils/api"
)

var buildTagRegexp = regexp.MustCompile(`^[ \t]*\+nirvana:api[ \t]*=(.*)\n`)

// APIBuilder builds api definitions by specified package.
type APIBuilder struct {
	paths []string
}

// NewAPIBuilder creates an api builder.
func NewAPIBuilder(paths ...string) *APIBuilder {
	return &APIBuilder{
		paths: paths,
	}
}

// Build build api definitions.
func (b *APIBuilder) Build() (*api.Definitions, error) {
	analyzer := api.NewAnalyzer()
	parents := []string{}
	for _, path := range b.paths {
		path, err := PackageForPath(path)
		if err != nil {
			return nil, err
		}
		pkg, err := analyzer.Import(path)
		if pkg == nil && err != nil {
			return nil, err
		}
		parents = append(parents, pkg.Path())
	}

	packages := map[string]bool{}
	for _, parent := range parents {
		for _, pkg := range analyzer.Packages(parent, false) {
			packages[pkg] = true
		}
	}

	const target = "github.com/caicloud/nirvana/definition"
	results := analyzer.FindPackages(target)
	found := false
	prefix := ""
	for _, real := range results {
		components := strings.Split(real, "/vendor/")
		if len(components) > 0 && components[len(components)-1] == target {
			components = components[:len(components)-1]
			found = true
			prefix = path.Join(components...)
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("can't find nirvana package %s", target)
	}

	descriptors := []function{}
	modifiers := []function{}

	var getFunction = func(pkg, name string) (*function, error) {
		f := &function{
			Pkg:  pkg,
			Name: name,
		}
		obj, err := analyzer.ObjectOf(pkg, name)
		if err != nil {
			return nil, err
		}
		key := fmt.Sprintf("%s.%s", pkg, name)
		if !obj.Exported() {
			return nil, fmt.Errorf("%s is not exported", key)
		}
		ft, ok := obj.Type().(*types.Signature)
		if !ok {
			return nil, fmt.Errorf("%s is not a function", key)
		}
		if ft.Params().Len() > 0 {
			return nil, fmt.Errorf("%s should not have parameters", key)
		}
		results := ft.Results()
		if results.Len() != 1 {
			return nil, fmt.Errorf("%s should have one result", key)
		}
		result := results.At(0)
		switch result.Type().(type) {
		case *types.Named:
			f.Array = false
		case *types.Slice:
			f.Array = true
		default:
			return nil, fmt.Errorf("%s should return an object or a slice", key)
		}
		return f, nil
	}

	for pkg := range packages {
		groups := analyzer.PackageComments(pkg)
		for _, group := range groups {
			matches := buildTagRegexp.FindAllStringSubmatch(group.Text(), -1)
			for _, match := range matches {
				if len(match) == 2 {
					tag := reflect.StructTag(match[1])
					descriptorFunc := tag.Get("descriptors")
					if descriptorFunc != "" {
						f, err := getFunction(pkg, descriptorFunc)
						if err != nil {
							return nil, err
						}
						descriptors = append(descriptors, *f)
					}
					modifierFunc := tag.Get("modifiers")
					if modifierFunc != "" {
						f, err := getFunction(pkg, modifierFunc)
						if err != nil {
							return nil, err
						}
						modifiers = append(modifiers, *f)
					}
				}
			}
		}
	}
	return b.runMain(descriptors, modifiers, prefix)
}

type function struct {
	Pkg   string
	Name  string
	Array bool
}

func (b *APIBuilder) runMain(descriptors, modifiers []function, root string) (*api.Definitions, error) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "nirvana")
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(tempDir, 0775); err != nil {
		return nil, err
	}
	data, err := b.file(descriptors, modifiers, root)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(tempDir, "main.go")
	if err := ioutil.WriteFile(path, data, 0664); err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "run", path)
	cmd.Stderr = os.Stderr
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	definitions := &api.Definitions{}
	if err := json.NewDecoder(buf).Decode(definitions); err != nil {
		return nil, err
	}
	return definitions, nil
}

func (b *APIBuilder) file(descriptors, modifiers []function, root string) ([]byte, error) {
	const tpl = `
package main

import (
	"fmt"
	"encoding/json"

	{{ range $i,$m := .modifiers }}
	m{{ $i }} "{{ $m.Pkg }}"
	{{ end }}
	{{ range $i,$d := .descriptors }}
	d{{ $i }} "{{ $d.Pkg }}"
	{{ end }}

	"{{ .root }}github.com/caicloud/nirvana/utils/api"
	"{{ .root }}github.com/caicloud/nirvana/log"
)

func main() {
	container := api.NewContainer()
	{{ range $i,$m := .modifiers }}
	container.AddModifier(m{{ $i }}.{{ $m.Name }}(){{ if $m.Array }}...{{ end }})
	{{ end }}
	{{ range $i,$d := .descriptors }}
	container.AddDescriptor(d{{ $i }}.{{ $d.Name }}(){{ if $d.Array }}...{{ end }})
	{{ end }}
	definitions, err := container.Generate()
	if definitions == nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(definitions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", data)
}
`
	tmpl, err := template.New("main.go").Parse(tpl)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, map[string]interface{}{
		"modifiers":   modifiers,
		"descriptors": descriptors,
		"root":        root,
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
