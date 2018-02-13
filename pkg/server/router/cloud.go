/*
Copyright 2017 caicloud authors. All rights reserved.

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

package router

import (
	"net/http"

	"github.com/caicloud/cyclone/cloud"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	restful "github.com/emicklei/go-restful"
)

// createCloud handles the request to create a cloud.
func (router *router) createCloud(request *restful.Request, response *restful.Response) {
	cloudOpt := &cloud.Options{}
	if err := httputil.ReadEntityFromRequest(request, cloudOpt); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	createdCloud, err := router.cloudManager.CreateCloud(cloudOpt)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, createdCloud)
}

// listClouds handles the request to list all clouds.
func (router *router) listClouds(request *restful.Request, response *restful.Response) {
	clouds := router.cloudManager.ListClouds()

	response.WriteHeaderAndEntity(http.StatusOK, clouds)
}

// deleteCloud handles the request to delete the cloud.
func (router *router) deleteCloud(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(cloudPathParameterName)

	if err := router.cloudManager.DeleteCloud(name); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusNoContent, nil)
}

// pingCloud handles the request to ping a cloud to check its health.
func (router *router) pingCloud(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(cloudPathParameterName)

	resp := make(map[string]string)
	err := router.cloudManager.PingCloud(name)
	if err != nil {
		resp["status"] = err.Error()
	} else {
		resp["status"] = "ok"
	}

	response.WriteHeaderAndEntity(http.StatusOK, resp)
}
