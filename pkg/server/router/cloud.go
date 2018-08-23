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

	"github.com/emicklei/go-restful"
	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

// createCloud handles the request to create a cloud.
func (router *router) createCloud(request *restful.Request, response *restful.Response) {
	cloud := &api.Cloud{}
	if err := httputil.ReadEntityFromRequest(request, cloud); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	createdCloud, err := router.cloudManager.CreateCloud(cloud)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, createdCloud)
}

// listClouds handles the request to list all clouds.
func (router *router) listClouds(request *restful.Request, response *restful.Response) {
	clouds, err := router.cloudManager.ListClouds()
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(clouds, len(clouds)))
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

// listWorkers handles the request to list all workers.
func (router *router) listWorkers(request *restful.Request, response *restful.Response) {
	cloudName := request.PathParameter(cloudPathParameterName)
	namespace := request.QueryParameter(namespaceQueryParameterName)

	workers, err := router.cloudManager.ListWorkers(cloudName, namespace)
	if err != nil {
		log.Errorf("list worker error:%v", err)
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(workers, len(workers)))
}
