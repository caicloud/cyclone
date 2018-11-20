/*
Copyright 2017 Caicloud Authors

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

package service

import (
	"context"
	"io"
	"mime/multipart"
	"reflect"
	"testing"

	"github.com/caicloud/nirvana/definition"
)

type vc struct{}

func (v *vc) Path(key string) (string, bool) {
	if key == "test" {
		return "path", true
	}
	return "", false
}

func (v *vc) Query(key string) ([]string, bool) {
	if key == "test" {
		return []string{"query"}, true
	}
	return nil, false
}

func (v *vc) Header(key string) ([]string, bool) {
	if key == "test" {
		return []string{"header"}, true
	}
	return nil, true
}

func (v *vc) Form(key string) ([]string, bool) {
	if key == "test" {
		return []string{"form"}, true
	}
	return nil, true
}

type file struct {
	data []byte
	read int
}

func (f *file) Read(p []byte) (n int, err error) {
	if f.read >= len(f.data) {
		return 0, io.EOF
	}
	written := copy(p, f.data[f.read:])
	f.read += written
	if f.read >= len(f.data) {
		err = io.EOF
	}
	return written, err
}
func (f *file) ReadAt(p []byte, off int64) (n int, err error) {
	panic("ReadAt is not implemented")
}
func (f *file) Seek(offset int64, whence int) (int64, error) {
	panic("Seek is not implemented")
}
func (f *file) Close() error {
	return nil
}

func (v *vc) File(key string) (multipart.File, bool) {
	if key == "test" {
		return &file{[]byte("test file"), 0}, true
	}
	return nil, false
}

func (v *vc) Body() (reader io.ReadCloser, contentType string, ok bool) {
	return &file{[]byte(`{"value":"test body"}`), 0}, definition.MIMEJSON, true
}

func TestPathParameterGenerator(t *testing.T) {
	g := &PathParameterGenerator{}
	if g.Source() != definition.Path {
		t.Fatalf("PathParameterGenerator has a wrong source: %s", g.Source())
	}
	if err := g.Validate("test", "default", reflect.TypeOf("")); err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), &vc{}, AllConsumers(), "test", reflect.TypeOf(""))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual("path", result) {
		t.Fatalf("PathParameterGenerator values is not equal: %+v, %+v", "path", result)
	}
}

func TestQueryParameterGenerator(t *testing.T) {
	g := &QueryParameterGenerator{}
	if g.Source() != definition.Query {
		t.Fatalf("QueryParameterGenerator has a wrong source: %s", g.Source())
	}
	if err := g.Validate("test", "default", reflect.TypeOf("")); err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), &vc{}, AllConsumers(), "test", reflect.TypeOf(""))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual("query", result) {
		t.Fatalf("QueryParameterGenerator values is not equal: %+v, %+v", "query", result)
	}
}

func TestHeaderParameterGenerator(t *testing.T) {
	g := &HeaderParameterGenerator{}
	if g.Source() != definition.Header {
		t.Fatalf("HeaderParameterGenerator has a wrong source: %s", g.Source())
	}
	if err := g.Validate("test", "default", reflect.TypeOf("")); err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), &vc{}, AllConsumers(), "test", reflect.TypeOf(""))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual("header", result) {
		t.Fatalf("HeaderParameterGenerator values is not equal: %+v, %+v", "header", result)
	}
}

func TestFormParameterGenerator(t *testing.T) {
	g := &FormParameterGenerator{}
	if g.Source() != definition.Form {
		t.Fatalf("FormParameterGenerator has a wrong source: %s", g.Source())
	}
	if err := g.Validate("test", "default", reflect.TypeOf("")); err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), &vc{}, AllConsumers(), "test", reflect.TypeOf(""))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual("form", result) {
		t.Fatalf("FormParameterGenerator values is not equal: %+v, %+v", "form", result)
	}
}

func TestFileParameterGenerator(t *testing.T) {
	g := &FileParameterGenerator{}
	if g.Source() != definition.File {
		t.Fatalf("FileParameterGenerator has a wrong source: %s", g.Source())
	}
	target := reflect.TypeOf((*io.Reader)(nil)).Elem()
	if err := g.Validate("test", nil, target); err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), &vc{}, AllConsumers(), "test", target)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result.(io.Reader); !ok {
		t.Fatalf("FileParameterGenerator result is not io.Reader: %s", reflect.TypeOf(result))
	}
}

type ts struct {
	Value string `json:"value"`
}

func TestBodyParameterGenerator(t *testing.T) {
	g := &BodyParameterGenerator{}
	if g.Source() != definition.Body {
		t.Fatalf("BodyParameterGenerator has a wrong source: %s", g.Source())
	}

	target := reflect.TypeOf((*io.Reader)(nil)).Elem()
	if err := g.Validate("test", nil, target); err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), &vc{}, AllConsumers(), "test", target)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result.(io.Reader); !ok {
		t.Fatalf("BodyParameterGenerator result is not io.Reader: %s", reflect.TypeOf(result))
	}

	target = reflect.TypeOf(&ts{})
	if err := g.Validate("test", nil, target); err != nil {
		t.Fatal(err)
	}
	result, err = g.Generate(context.Background(), &vc{}, AllConsumers(), "test", target)
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := result.(*ts); !ok || r.Value != "test body" {
		t.Fatalf("BodyParameterGenerator result is not correct: %+v", result)
	}
}

func TestPrefabParameterGenerator(t *testing.T) {
	g := &PrefabParameterGenerator{}
	if g.Source() != definition.Prefab {
		t.Fatalf("PrefabParameterGenerator has a wrong source: %s", g.Source())
	}
	target := reflect.TypeOf((*context.Context)(nil)).Elem()
	if err := g.Validate("context", nil, target); err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), &vc{}, AllConsumers(), "context", target)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result.(context.Context); !ok {
		t.Fatalf("PrefabParameterGenerator result is not io.Reader: %s", reflect.TypeOf(result))
	}
}

type as struct {
	Hello     string          `source:"path, hello, default=world"`
	IsDefault bool            `source:"path,isDefault, default=true,test=10"`
	Age       int             `source:"Path,age"`
	Name      string          `source:"Path,name"`
	Path      string          `source:"Path,test"`
	Query     string          `source:"query,test"`
	Header    string          `source:"header,test"`
	Form      string          `source:"form,test"`
	File      io.Reader       `source:"File,test"`
	Body      *ts             `source:"Body"`
	Context   context.Context `source:"Prefab,context"`
}

func TestAutoParameterGenerator(t *testing.T) {
	g := &AutoParameterGenerator{}
	if g.Source() != definition.Auto {
		t.Fatalf("AutoParameterGenerator has a wrong source: %s", g.Source())
	}
	target := reflect.TypeOf(&as{})
	if err := g.Validate("test", nil, target); err != nil {
		t.Fatal(err)
	}
	result, err := g.Generate(context.Background(), &vc{}, AllConsumers(), "test", target)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", result)
	if r, ok := result.(*as); !ok ||
		r.Hello != "world" ||
		!r.IsDefault ||
		r.Path != "path" ||
		r.Query != "query" ||
		r.Header != "header" ||
		r.Form != "form" ||
		r.File == nil ||
		r.Body == nil ||
		r.Context == nil ||
		r.Body.Value != "test body" {
		t.Fatalf("BodyParameterGenerator result is not correct: %+v", result)
	}
}

func TestInvalidAutoParameter(t *testing.T) {
	g := &AutoParameterGenerator{}
	if g.Source() != definition.Auto {
		t.Fatalf("AutoParameterGenerator has a wrong source: %s", g.Source())
	}
	target := reflect.TypeOf(1)
	err := g.Validate("test", "test", target)
	if err == nil {
		t.Fatal("TestInvalidAutoParameter: Validate should return an error")
	}

	t.Log(err)
}
