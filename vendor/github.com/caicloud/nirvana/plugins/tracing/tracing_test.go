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

package tracing

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
)

var test = definition.Descriptor{
	Path:        "/",
	Description: "trace example",
	Definitions: []definition.Definition{
		{
			Method: definition.Get,
			Function: func(ctx context.Context) (string, error) {
				return "success", nil
			},
			Consumes: []string{definition.MIMEText},
			Produces: []string{definition.MIMEText},
			Results:  definition.DataErrorResults("results"),
		},
	},
}

func TestMiddleware(t *testing.T) {
	config := nirvana.NewDefaultConfig().
		Configure(
			DefaultTracer("example", "127.0.0.1:6831"),
			AddHook(&DefaultHook{}),
			nirvana.Descriptor(test),
		)

	build, cleaner, err := nirvana.NewServer(config).Builder()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = cleaner(); err != nil {
			t.Fatal(err)
		}
	}()
	service, err := build.Build()
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(service)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "success" {
		t.Fatalf(`response string expected "success" but got "%s"`, string(b))
	}
}
