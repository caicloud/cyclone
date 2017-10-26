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
	"fmt"
	"net/http"

	"github.com/zoumo/logdog"

	oldapi "github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/event"
	"github.com/caicloud/cyclone/pkg/api"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	restful "github.com/emicklei/go-restful"
)

// createPipelineRecord handles the request to perform pipeline, which will create a pipeline record.
func (router *router) createPipelineRecord(request *restful.Request, response *restful.Response) {
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

// getPipelineRecord handles the request to get a pipeline record.
func (router *router) getPipelineRecord(request *restful.Request, response *restful.Response) {
	pipelineRecordID := request.PathParameter(pipelineRecordPathParameterName)

	pipelineRecord, err := router.pipelineRecordManager.GetPipelineRecord(pipelineRecordID)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, pipelineRecord)
}

// listPipelineRecords handles the request to list pipeline records.
func (router *router) listPipelineRecords(request *restful.Request, response *restful.Response) {
	queryParams, err := httputil.QueryParamsFromRequest(request)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)

	pipelineRecords, count, err := router.pipelineRecordManager.ListPipelineRecords(projectName, pipelineName, queryParams)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(pipelineRecords, count))
}

// deletePipelineRecord handles the request to delete a pipeline record.
func (router *router) deletePipelineRecord(request *restful.Request, response *restful.Response) {
	pipelineRecordID := request.PathParameter(pipelineRecordPathParameterName)

	if err := router.pipelineRecordManager.DeletePipelineRecord(pipelineRecordID); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusNoContent, nil)
}

// updatePipelineRecordStatus handles the request to update a pipeline record status, only support to set 'Aborted'
// status for running pipeline records.
func (router *router) updatePipelineRecordStatus(request *restful.Request, response *restful.Response) {
	pipelineRecordID := request.PathParameter(pipelineRecordPathParameterName)

	pipelineRecord, err := router.pipelineRecordManager.GetPipelineRecord(pipelineRecordID)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	if pipelineRecord.Status != api.Running {
		logdog.Warnf("The pipeline record %s is not running, can not be aborted, will do no action", pipelineRecord.Name)
		response.WriteHeaderAndEntity(http.StatusOK, pipelineRecord)
		return
	}

	type RecordStatus struct {
		Status api.Status `json:"status"`
	}

	recordStatus := &RecordStatus{}
	if err := httputil.ReadEntityFromRequest(request, recordStatus); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	if recordStatus.Status != api.Aborted {
		err = httperror.ErrorValidationFailed.Format("status", "only support Aborted")
		httputil.ResponseWithError(response, err)
		return
	}

	e, err := event.LoadEventFromEtcd(oldapi.EventID(pipelineRecord.ID))
	if err != nil {
		err := fmt.Errorf("Unable to find event by versonID %v", pipelineRecord.ID)
		logdog.Error(err)
		httputil.ResponseWithError(response, err)
		return
	}

	if e.Status == oldapi.EventStatusRunning {
		e.Status = oldapi.EventStatusCancel
		event.SaveEventToEtcd(e)

		pipelineRecord.Status = api.Aborted
	}

	response.WriteHeaderAndEntity(http.StatusOK, pipelineRecord)
}

// getPipelineRecordLogs handles the request to get pipeline record logs, only supports finished pipeline records.
func (router *router) getPipelineRecordLogs(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)
	pipelineRecordID := request.PathParameter(pipelineRecordPathParameterName)
	download, err := httputil.DownloadQueryParamsFromRequest(request)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	logs, err := router.pipelineRecordManager.GetPipelineRecordLogs(pipelineRecordID)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.AddHeader(restful.HEADER_ContentType, "text/plain")
	if download {
		logFileName := fmt.Sprintf("%s-%s-%s-log.txt", projectName, pipelineName, pipelineRecordID)
		response.AddHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", logFileName))
	}

	response.Write([]byte(logs))
}
