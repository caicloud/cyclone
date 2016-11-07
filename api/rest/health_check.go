/*
Copyright 2016 caicloud authors. All rights reserved.

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

package rest

import (
	"net/http"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/kafka"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
)

// healthCheck checks health.
//
// GET: /api/v0.1/healthcheck
//
// RESPONSE: (HealthCheckResponse)
//  {
//    "error_msg": (string) set IFF the request fails.
//  }
func healthCheck(request *restful.Request, response *restful.Response) {
	var healthCheckResponse api.HealthCheckResponse

	ds := store.NewStore()
	defer ds.Close()

	// Check mongo.
	if nil != ds.Ping() {
		healthCheckResponse.ErrorMessage = "mongo disconnect"
		response.WriteHeaderAndEntity(http.StatusNotAcceptable, healthCheckResponse)
		return
	}

	// Check kafka.
	if false == kafka.IsConnected() {
		healthCheckResponse.ErrorMessage = "kafka disconnect"
		response.WriteHeaderAndEntity(http.StatusNotAcceptable, healthCheckResponse)
		return
	}

	healthCheckResponse.ErrorMessage = "ok"
	response.WriteHeaderAndEntity(http.StatusOK, healthCheckResponse)
}
