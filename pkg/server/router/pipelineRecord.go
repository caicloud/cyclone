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

	httputil "github.com/caicloud/cyclone/pkg/util/http"
	restful "github.com/emicklei/go-restful"
)

// getPipelineRecord handles the request to get a pipeline record.
func (router *router) getPipelineRecord(request *restful.Request, response *restful.Response) {
	pipelineRecordID := request.PathParameter(pipelineRecordPathParameterName)

	pipelineRecord, err := router.pipelineRecordManager.GetPipelineRecord(pipelineRecordID)
	if err != nil {
		httputil.ResponseWithError(response, http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, pipelineRecord)
}

// listPipelineRecords handles the request to list pipeline records.
func (router *router) listPipelineRecords(request *restful.Request, response *restful.Response) {
	queryParams := httputil.QueryParamsFromRequest(request)

	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)

	pipelineRecords, count, err := router.pipelineRecordManager.ListPipelineRecords(projectName, pipelineName, queryParams)
	if err != nil {
		httputil.ResponseWithError(response, http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(pipelineRecords, len(pipelineRecords), count))
}

// deletePipelineRecord handles the request to delete a pipeline record.
func (router *router) deletePipelineRecord(request *restful.Request, response *restful.Response) {
	pipelineRecordID := request.PathParameter(pipelineRecordPathParameterName)

	if err := router.pipelineRecordManager.DeletePipelineRecord(pipelineRecordID); err != nil {
		httputil.ResponseWithError(response, http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusNoContent, nil)
}
