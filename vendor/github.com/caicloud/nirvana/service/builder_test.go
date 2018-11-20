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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
)

type responseWriter struct {
	code   int
	header http.Header
	buf    *bytes.Buffer
}

func newRW() *responseWriter {
	return &responseWriter{0, http.Header{}, bytes.NewBuffer(nil)}
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) Write(d []byte) (int, error) {
	return r.buf.Write(d)
}

func (r *responseWriter) WriteHeader(code int) {
	r.code = code
}

var desc = definition.Descriptor{
	Path:        "/api/v1/",
	Definitions: []definition.Definition{},
	Consumes:    []string{"application/json"},
	Produces:    []string{"application/json"},
	Children: []definition.Descriptor{
		{
			Path: "/{target1}/{target2}",
			Definitions: []definition.Definition{
				{
					Method:   definition.Create,
					Function: Handle,
					Parameters: []definition.Parameter{
						{
							Source: definition.Header,
							Name:   "User-Agent",
						},
						{
							Source: definition.Query,
							Name:   "target1",
						},
						{
							Source:  definition.Query,
							Name:    "target2",
							Default: false,
						},
						{
							Source: definition.Body,
							Name:   "app",
						},
					},
					Results: []definition.Result{
						{Destination: definition.Data},
						{Destination: definition.Error},
					},
				},
			},
		},
	},
}

type Application struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Target    string `json:"target"`
	Target1   int    `json:"target2"`
	Target2   bool   `json:"target1"`
}

func Handle(ctx context.Context, userAgent string, target1 int, target2 bool, app *Application) (*Application, error) {
	path := HTTPContextFrom(ctx).RoutePath()
	if path != "/api/v1/{target1}/{target2}" {
		return nil, fmt.Errorf("http abstract path is not correct: %s", path)
	}
	app.Target = userAgent
	app.Target1 = target1
	app.Target2 = target2
	return app, nil
}

func TestServer(t *testing.T) {
	u, _ := url.Parse("/api/v1/1222/false?target1=1&target2=false")
	data := []byte(`{
	"name": "asdasd",
	"namespace": "system"
}`)

	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"Accept":       []string{"application/json"},
			"User-Agent":   []string{"nothing"},
		},
		ContentLength: int64(len(data)),
	}
	builder := NewBuilder()
	builder.SetModifier(FirstContextParameter())
	builder.AddFilter(RedirectTrailingSlash(), FillLeadingSlash(), ParseRequestForm())
	err := builder.AddDescriptor(desc)
	if err != nil {
		t.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.Background())
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	resp := newRW()
	s.ServeHTTP(resp, req)
	t.Log(resp.code)
	t.Log(resp.header)
	t.Logf("%s", resp.buf.Bytes())
	if resp.code != 201 {
		t.Fatalf("Response code should be 201, but got: %d", resp.code)
	}
	if resp.header == nil || resp.header.Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type should be application/json, but got: %s", resp.header.Get("Content-Type"))
	}
	result := resp.buf.String()
	target := `{"name":"asdasd","namespace":"system","target":"nothing","target2":1,"target1":false}` + "\n"
	if result != target {
		t.Fatalf("Response does not match: %s", result)
	}
}

var childrenDesc = definition.Descriptor{
	Path:        "",
	Definitions: []definition.Definition{},
	Consumes:    []string{definition.MIMEJSON},
	Produces:    []string{definition.MIMEJSON},
	Children: []definition.Descriptor{
		{
			Path: "/api/v1",
			Children: []definition.Descriptor{
				{
					Path: "",
					Definitions: []definition.Definition{
						{
							Method: definition.Get,
							Function: func(ctx context.Context) string {
								return ""
							},
							Results: []definition.Result{
								{Destination: definition.Data},
							},
						},
					},
					Children: []definition.Descriptor{
						{
							Path: "/ping",
							Definitions: []definition.Definition{
								{
									Method:   definition.Get,
									Function: echoHandle,
									Results: []definition.Result{
										{Destination: definition.Data},
										{Destination: definition.Error},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

type echoResult struct {
	Message string `json:"message"`
}

func echoHandle(ctx context.Context) (*echoResult, error) {
	return &echoResult{
		Message: "pong",
	}, nil
}

func TestChildrenPath(t *testing.T) {
	u, _ := url.Parse("/api/v1/ping")

	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			"Content-Type": []string{definition.MIMEJSON},
			"Accept":       []string{definition.MIMEJSON},
		},
	}
	builder := NewBuilder()
	builder.SetModifier(FirstContextParameter())
	err := builder.AddDescriptor(childrenDesc)
	if err != nil {
		t.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.Background())
	resp := newRW()
	s.ServeHTTP(resp, req)

	var result echoResult
	if err := json.NewDecoder(resp.buf).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if resp.code != 200 && result.Message != "pong" {
		t.Fatalf("Response code should be 200, but got: %d", resp.code)
	}
}

var homeDesc = definition.Descriptor{
	Path:        "",
	Definitions: []definition.Definition{},
	Consumes:    []string{definition.MIMEJSON},
	Produces:    []string{definition.MIMEJSON},
	Children: []definition.Descriptor{
		{
			Path: "/",
			Definitions: []definition.Definition{
				{
					Method:   definition.Get,
					Function: homeHandle,
					Results: []definition.Result{
						{Destination: definition.Data},
						{Destination: definition.Error},
					},
				},
			},
		},
	},
}

type homeResult struct {
	Message string `json:"message"`
}

func homeHandle(ctx context.Context) (*homeResult, error) {
	return &homeResult{
		Message: "home",
	}, nil
}

func TestHomePath(t *testing.T) {
	u, _ := url.Parse("/")

	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			"Content-Type": []string{definition.MIMEJSON},
			"Accept":       []string{definition.MIMEJSON},
		},
	}
	builder := NewBuilder()
	builder.SetModifier(FirstContextParameter())
	err := builder.AddDescriptor(homeDesc)
	if err != nil {
		t.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.Background())
	resp := newRW()
	s.ServeHTTP(resp, req)

	var result homeResult
	if err := json.NewDecoder(resp.buf).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if resp.code != 200 && result.Message != "home" {
		t.Fatalf("Response code should be 200, but got: %d", resp.code)
	}
}

type User struct {
	Username string `source:"Header,Username"`
	Password string `source:"Header,Password"`
}

var defaultParamsDesc = definition.Descriptor{
	Definitions: []definition.Definition{},
	Consumes:    []string{definition.MIMEJSON},
	Produces:    []string{definition.MIMEJSON},
	Children: []definition.Descriptor{
		{
			Path: "/default",
			Definitions: []definition.Definition{
				{
					Method:   definition.Get,
					Function: defaultParamsHandler,
					Parameters: []definition.Parameter{
						{
							Source:  definition.Query,
							Name:    "q1",
							Default: "q1",
						},
						{
							Source: definition.Query,
							Name:   "q2",
						},
						{
							Source: definition.Header,
							Name:   "X-Tenant",
						},
						{
							Source: definition.Header,
							Name:   "X-Tenant2",
						},
						{
							Source: definition.Auto,
							Name:   "user",
						},
					},
					Results: []definition.Result{
						{Destination: definition.Data},
						{Destination: definition.Error},
					},
				},
			},
		},
	},
}

func defaultParamsHandler(ctx context.Context, q1, q2, tenant, tenant2 string, u *User) (string, error) {
	if q1 == "q1" && q2 == "" && tenant == "" && tenant2 == "tenant2" && u.Username == "name" && u.Password == "pwd" {
		return "match", nil
	}
	return "", errors.NotFound.Error("not match params")

}

func TestDefaultParams(t *testing.T) {
	u, _ := url.Parse("/default")

	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			"Content-Type": []string{definition.MIMEJSON},
			"Accept":       []string{definition.MIMEJSON},
			"X-Tenant2":    []string{"tenant2"},
			"Username":     []string{"name"},
			"Password":     []string{"pwd"},
		},
	}
	builder := NewBuilder()
	builder.SetModifier(FirstContextParameter())
	err := builder.AddDescriptor(defaultParamsDesc)
	if err != nil {
		t.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.Background())
	resp := newRW()
	s.ServeHTTP(resp, req)

	if resp.code != 200 && resp.buf.String() != "match" {
		t.Fatalf("Response code should be 200, but got: %d", resp.code)
	}
}

var errDesc = definition.Descriptor{
	Path:        "/api/v1/",
	Definitions: []definition.Definition{},
	Consumes:    []string{definition.MIMEAll},
	Produces:    []string{definition.MIMEJSON},
	Children: []definition.Descriptor{
		{
			Path: "/{err}",
			Definitions: []definition.Definition{
				{
					Method:        definition.Get,
					ErrorProduces: []string{definition.MIMEXML},
					Function: func(err bool) (*Application, error) {
						if err {
							return nil, errors.NotFound.Error("error for test ${test}", err)
						}
						return &Application{
							Name: "test",
						}, nil
					},
					Parameters: []definition.Parameter{
						{
							Source: definition.Path,
							Name:   "err",
						},
					},
					Results: definition.DataErrorResults(""),
				},
			},
		},
	},
}

func TestErrorProduces(t *testing.T) {
	builder := NewBuilder()
	err := builder.AddDescriptor(errDesc)
	if err != nil {
		t.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{
		Method: "GET",
		Header: http.Header{
			"Accept": []string{"application/json, application/xml"},
		},
	}
	req = req.WithContext(context.Background())

	u, _ := url.Parse("/api/v1/true")
	req.URL = u
	resp := newRW()
	s.ServeHTTP(resp, req)

	if resp.code != 404 {
		t.Fatalf("Response code should be 404, but got: %d", resp.code)
	}
	desired := `<message><Message>error for test true</Message><Data><test>true</test></Data></message>`
	if resp.buf.String() != desired {
		t.Fatalf("Response data is not desired: %s", resp.buf.String())
	}

	u, _ = url.Parse("/api/v1/false")
	req.URL = u
	resp = newRW()
	s.ServeHTTP(resp, req)

	if resp.code != 200 {
		t.Fatalf("Response code should be 200, but got: %d", resp.code)
	}
	desired = `{"name":"test","namespace":"","target":"","target2":0,"target1":false}` + "\n"
	if resp.buf.String() != desired {
		t.Fatalf("Response data is not desired: %s", resp.buf.String())
	}
}

func BenchmarkServer(b *testing.B) {
	u, _ := url.Parse("/api/v1/1222/false?target1=1&target2=false")
	data := []byte(`{
	"name": "asdasd",
	"namespace": "system"
}`)

	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"Accept":       []string{"application/json"},
			"User-Agent":   []string{"nothing"},
		},
		ContentLength: int64(len(data)),
	}
	builder := NewBuilder()
	builder.SetModifier(FirstContextParameter())
	builder.AddFilter(RedirectTrailingSlash(), FillLeadingSlash(), ParseRequestForm())
	err := builder.AddDescriptor(desc)
	if err != nil {
		b.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		b.Fatal(err)
	}
	req = req.WithContext(context.Background())
	resp := newRW()
	for n := 0; n < b.N; n++ {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		s.ServeHTTP(resp, req)
		resp.buf = bytes.NewBuffer(resp.buf.Bytes())
	}
}
