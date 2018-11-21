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
	"time"

	"github.com/caicloud/nirvana/plugins/tracing/utils"
	"github.com/opentracing/opentracing-go/ext"
)

var config = `{"db":{"username":"admin","password":"123456"}}`

var configCache map[string]map[string]string

func init() {
	configCache = make(map[string]map[string]string)
	if err := json.Unmarshal([]byte(config), &configCache); err != nil {
		panic(err)
	}
}

func GetConfig(ctx context.Context, config string) (map[string]string, error) {
	span, _ := utils.StartSpanFromContext(ctx, "getConfig")
	defer span.Finish()

	ext.SpanKindRPCServer.Set(span)
	span.SetTag("config", config)
	time.Sleep(time.Second)

	data, ok := configCache[config]
	if !ok {
		return nil, fmt.Errorf("config %s not found", config)
	}

	return data, nil
}
