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
	"reflect"
	"time"

	"github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/event"
	"github.com/caicloud/cyclone/pkg/log"

	httputil "github.com/caicloud/cyclone/pkg/util/http"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
	"github.com/emicklei/go-restful"
)

const (
	// Stop signal
	stopSignal = 1

	// Interval of loading logfragment
	loadInterval = 100 * time.Millisecond
)

// createPipelineRecord handles the request to perform pipeline, which will create a pipeline record.
func (router *router) createPipelineRecord(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)

	pipeline, err := router.pipelineManager.GetPipeline(projectName, pipelineName)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	performParams := &api.PipelinePerformParams{}
	if err := httputil.ReadEntityFromRequest(request, performParams); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	pipelineRecord := &api.PipelineRecord{
		Name:          performParams.Name,
		PipelineID:    pipeline.ID,
		PerformParams: performParams,
		Trigger:       request.Request.Header.Get(httputil.HeaderUser),
		Status:        api.Pending,
	}
	createdPipelineRecord, err := router.pipelineRecordManager.CreatePipelineRecord(pipelineRecord)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, createdPipelineRecord)
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

	e, err := event.GetEvent(pipelineRecord.ID)
	if err != nil {
		err := fmt.Errorf("Unable to find event by versonID %v", pipelineRecord.ID)
		logdog.Error(err)
		httputil.ResponseWithError(response, err)
		return
	}

	abortPipelineRecord(e.PipelineRecord)
	// event and pipeline record both will be updated in method "UpdateEvent"
	err = event.UpdateEvent(e)
	if err != nil {
		log.Errorf("update event %s error: %v", e.ID, err)
	}

	response.WriteHeaderAndEntity(http.StatusOK, pipelineRecord)
}

func abortPipelineRecord(p *api.PipelineRecord) {

	if p.Status == api.Running {
		p.Status = api.Aborted
	}

	if p.StageStatus != nil {
		stageStatusElem := reflect.ValueOf(p.StageStatus).Elem()

		for i := 0; i < stageStatusElem.NumField(); i++ {
			if !stageStatusElem.Field(i).IsNil() {

				stageElem := stageStatusElem.Field(i).Elem()
				statusValue := stageElem.FieldByName("Status")
				if statusValue.String() == string(api.Running) {
					statusValue.Set(reflect.ValueOf(api.Aborted))
				}
			}

		}

	}

}

// getPipelineRecordLogs handles the request to get pipeline record logs, only supports finished pipeline records.
func (router *router) getPipelineRecordLogs(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	pipelineName := request.PathParameter(pipelinePathParameterName)
	pipelineRecordID := request.PathParameter(pipelineRecordPathParameterName)
	stage := request.QueryParameter(pipelineRecordStageQueryParameterName)
	task := request.QueryParameter(pipelineRecordTaskQueryParameterName)
	download, err := httputil.DownloadQueryParamsFromRequest(request)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	logs, err := router.pipelineRecordManager.GetPipelineRecordLogs(pipelineRecordID, stage, task)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.AddHeader(restful.HEADER_ContentType, "text/plain")
	if download {
		logFileName := fmt.Sprintf("%s-%s-%s-%s-log.txt", projectName, pipelineName, stage, pipelineRecordID)
		response.AddHeader("Content-Disposition", fmt.Sprintf("attachment; filename=%s", logFileName))
	}

	response.Write([]byte(logs))
}

// receivePipelineRecordLogStream receives real-time log of pipeline record.
func (router *router) receivePipelineRecordLogStream(request *restful.Request, response *restful.Response) {
	recordID := request.PathParameter(pipelineRecordPathParameterName)
	stage := request.QueryParameter(pipelineRecordStageQueryParameterName)
	task := request.QueryParameter(pipelineRecordTaskQueryParameterName)

	_, err := router.pipelineRecordManager.GetPipelineRecord(recordID)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to find pipeline record %s for err: %s", recordID, err.Error()))
		httputil.ResponseWithError(response, httperror.ErrorContentNotFound.Format(recordID))
		return
	}

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to upgrade websocket for err: %s", err.Error()))
		httputil.ResponseWithError(response, httperror.ErrorUnknownInternal.Format(err.Error()))
		return
	}
	defer ws.Close()

	if err := router.pipelineRecordManager.ReceivePipelineRecordLogStream(recordID, stage, task, ws); err != nil {
		log.Error(fmt.Sprintf("Fail to receive log stream for pipeline record %s: %s", recordID, err.Error()))
		httputil.ResponseWithError(response, httperror.ErrorUnknownInternal.Format(err.Error()))
		return
	}
}

// getPipelineRecordLogStream gets real-time log of pipeline record refering to recordID
func (router *router) getPipelineRecordLogStream(request *restful.Request, response *restful.Response) {
	recordID := request.PathParameter(pipelineRecordPathParameterName)
	stage := request.QueryParameter(pipelineRecordStageQueryParameterName)
	task := request.QueryParameter(pipelineRecordTaskQueryParameterName)

	_, err := router.pipelineRecordManager.GetPipelineRecord(recordID)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to find pipeline record %s for err: %s", recordID, err.Error()))
		httputil.ResponseWithError(response, httperror.ErrorContentNotFound.Format(recordID))
		return
	}

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to upgrade websocket for err: %s", err.Error()))
		httputil.ResponseWithError(response, httperror.ErrorUnknownInternal.Format(err.Error()))
		return
	}
	defer ws.Close()

	if err := router.pipelineRecordManager.GetPipelineRecordLogStream(recordID, stage, task, ws); err != nil {
		log.Error(fmt.Sprintf("Unable to get logstream for pipeline record %s for err: %s", recordID, err.Error()))
		httputil.ResponseWithError(response, httperror.ErrorUnknownInternal.Format(err.Error()))
		return
	}
}
