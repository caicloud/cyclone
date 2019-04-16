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
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
)

// Analyzer analyzes go packages.
type Analyzer struct {
	root         string
	files        map[string]*ast.File
	packageFiles map[string][]*ast.File
	packages     map[string]*types.Package
	fileset      *token.FileSet
}

// NewAnalyzer creates a code analyzer.
func NewAnalyzer(root string) *Analyzer {
	return &Analyzer{
		root:         root,
		files:        map[string]*ast.File{},
		packageFiles: map[string][]*ast.File{},
		packages:     map[string]*types.Package{},
		fileset:      token.NewFileSet(),
	}
}

// Import imports a package and all packages it depends on.
func (a *Analyzer) Import(path string) (*types.Package, error) {
	pkg, err := build.Import(path, a.root, 0)
	if err != nil {
		return nil, err
	}
	if parsedPkg, ok := a.packages[pkg.ImportPath]; ok {
		return parsedPkg, nil
	}
	cfg := &types.Config{
		IgnoreFuncBodies: true,
		Importer:         a,
		Error: func(err error) {
		},
	}
	files := make([]*ast.File, len(pkg.GoFiles))
	for i, path := range pkg.GoFiles {
		path = filepath.Join(pkg.Dir, path)
		file, ok := a.files[path]
		if !ok {
			file, err = parser.ParseFile(a.fileset, path, nil, parser.ParseComments)
			if err != nil {
				return nil, err
			}
		}
		a.files[path] = file
		files[i] = file
	}
	a.packageFiles[pkg.ImportPath] = files
	parsedPkg, err := cfg.Check(pkg.ImportPath, a.fileset, files, nil)
	a.packages[path] = parsedPkg
	if pkg.ImportPath != path {
		a.packages[pkg.ImportPath] = parsedPkg
	}
	return parsedPkg, err
}

// PackageComments returns comments above package keyword.
// Import package before calling this method.
func (a *Analyzer) PackageComments(path string) []*ast.CommentGroup {
	files, ok := a.packageFiles[path]
	if !ok {
		return nil
	}
	results := []*ast.CommentGroup{}
	for _, file := range files {
		for _, cg := range file.Comments {
			if cg.End() < file.Package {
				results = append(results, cg)
			}
		}
	}
	return results
}

// Packages returns packages under specified directory (including itself).
// Import package before calling this method.
func (a *Analyzer) Packages(parent string, vendor bool) []string {
	results := []string{}
	for path := range a.packages {
		if !vendor && strings.Index(path, "/vendor/") > 0 {
			continue
		}
		if strings.HasPrefix(path, parent) {
			results = append(results, path)
		}
	}
	return results
}

// FindPackages returns packages which contain target.
// Import package before calling this method.
func (a *Analyzer) FindPackages(target string) []string {
	results := []string{}
	for path := range a.packages {
		if strings.Contains(path, target) {
			results = append(results, path)
		}
	}
	return results
}

// Comments returns immediate comments above pos.
// Import package before calling this method.
func (a *Analyzer) Comments(pos token.Pos) *ast.CommentGroup {
	position := a.fileset.Position(pos)
	file := a.files[position.Filename]
	for _, cg := range file.Comments {
		cgPos := a.fileset.Position(cg.End())
		if cgPos.Line == position.Line-1 {
			return cg
		}
	}
	return nil
}

// ObjectOf returns declaration object of target.
func (a *Analyzer) ObjectOf(pkg, name string) (types.Object, error) {
	p, err := a.Import(pkg)
	// Ignore the error if p is not nil.
	// We need to rewrite analyzer with go/parser rather than go/types.
	if p == nil && err != nil {
		return nil, err
	}
	obj := p.Scope().Lookup(name)
	if obj == nil {
		return nil, fmt.Errorf("can't find declearation of %s.%s", pkg, name)
	}
	return obj, nil
}
