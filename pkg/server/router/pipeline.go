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

	"github.com/caicloud/cyclone/pkg/api"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	"github.com/emicklei/go-restful"
)

// createPipeline handles the request to create a pipeline.
func (router *router) createPipeline(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	project, err := router.projectManager.GetProject(projectName)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	pipeline := &api.Pipeline{}
	if err := httputil.ReadEntityFromRequest(request, pipeline); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	pipeline.ProjectID = project.ID
	createdPipeline, err := router.pipelineManager.CreatePipeline(projectName, pipeline)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, createdPipeline)
}

// getPipeline handles the request to get a pipeline.
func (router *router) getPipeline(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)

	pipeline, err := router.pipelineManager.GetPipeline(projectName, pipelineName)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, pipeline)
}

// listPipelines handles the request to list pipelines.
func (router *router) listPipelines(request *restful.Request, response *restful.Response) {
	queryParams, err := httputil.QueryParamsFromRequest(request)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	projectName := request.PathParameter(projectPathParameterName)

	pipelines, count, err := router.pipelineManager.ListPipelines(projectName, queryParams)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(pipelines, count))
}

// updatePipeline handles the request to update a pipeline.
func (router *router) updatePipeline(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)

	pipeline := &api.Pipeline{}
	if err := httputil.ReadEntityFromRequest(request, pipeline); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	updatedPipeline, err := router.pipelineManager.UpdatePipeline(projectName, pipelineName, pipeline)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, updatedPipeline)
}

// deletePipeline handles the request to delete a pipeline.
func (router *router) deletePipeline(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)

	if err := router.pipelineManager.DeletePipeline(projectName, pipelineName); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusNoContent, nil)
}

// performPipeline handles the request to perform a pipeline.
func (router *router) performPipeline(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)

	performParams := &api.PipelinePerformParams{}
	if err := httputil.ReadEntityFromRequest(request, performParams); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	if err := router.pipelineManager.PerformPipeline(projectName, pipelineName, performParams); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, nil)
}
