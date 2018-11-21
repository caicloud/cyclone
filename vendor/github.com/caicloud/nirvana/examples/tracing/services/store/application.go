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

package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/caicloud/nirvana/examples/tracing/services/models"
	tracingHTTP "github.com/caicloud/nirvana/plugins/tracing/clients/http"
	"github.com/caicloud/nirvana/plugins/tracing/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tlog "github.com/opentracing/opentracing-go/log"
)

var cache []*models.Application

var httpClient = &http.Client{Transport: &tracingHTTP.Transport{
	EnableHTTPtrace: true,
	OperationName:   "getConfig",
}}

func CreateApplication(ctx context.Context, namespace, name string, configName string) (*models.Application, error) {
	var span opentracing.Span
	span, ctx = utils.StartSpanFromContext(ctx, "createApplication")
	defer span.Finish()

	// 1. check the application is exists
	exists := applicationExists(ctx, namespace, name)
	if exists {
		err := fmt.Errorf("the application exists. namespace:%s, name:%s", namespace, name)
		ext.Error.Set(span, true)
		span.LogFields(
			tlog.Error(err),
		)
		return nil, err
	}

	// 2. load config
	data, err := loadConfig(ctx, configName)
	if err != nil {
		return nil, err
	}

	app := &models.Application{
		Namespace:  namespace,
		Name:       name,
		ConfigData: data,
	}

	// 3. save application to DB
	saveApplication(ctx, app)

	return app, nil
}

func applicationExists(ctx context.Context, namespace, name string) bool {
	span, _ := utils.StartSpanFromContext(ctx, "SearchApplication")
	defer span.Finish()

	span.SetTag("namespace", namespace)
	span.SetTag("name", name)

	for _, app := range cache {
		if app.Namespace == namespace && app.Name == name {
			return true
		}
	}
	return false
}

func saveApplication(ctx context.Context, app *models.Application) {
	span, _ := utils.StartSpanFromContext(ctx, "saveApplicationToCache")
	defer span.Finish()

	app.CreationTime = time.Now()
	time.Sleep(time.Second / 2)
	cache = append(cache, app)
}

func loadConfig(ctx context.Context, config string) (map[string]string, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:8081/config?config=%s", config), nil)
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var b []byte
		b, err = ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("status code should be 200, but got: %d\n%s", resp.StatusCode, string(b))
		return nil, err
	}

	var data map[string]string
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
