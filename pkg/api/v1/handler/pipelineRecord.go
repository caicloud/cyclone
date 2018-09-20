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

package handler

import (
	"context"
	"fmt"
	"reflect"
	"time"

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/event"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
)

const (
	// Stop signal
	stopSignal = 1

	// Interval of loading logfragment
	loadInterval = 100 * time.Millisecond
)

// CreatePipelineRecord handles the request to perform pipeline, which will create a pipeline record.
func CreatePipelineRecord(ctx context.Context, projectName, pipelineName, user string) (*api.PipelineRecord, error) {
	pipeline, err := pipelineManager.GetPipeline(projectName, pipelineName, 0, 0, 0)
	if err != nil {
		return nil, err
	}

	performParams := &api.PipelinePerformParams{}
	contextutil.GetJsonPayload(ctx, performParams)
	if err != nil {
		return nil, err
	}

	pipelineRecord := &api.PipelineRecord{
		Name:          performParams.Name,
		PipelineID:    pipeline.ID,
		PerformParams: performParams,
		Trigger:       user,
		Status:        api.Pending,
	}
	createdPipelineRecord, err := pipelineRecordManager.CreatePipelineRecord(pipelineRecord)
	if err != nil {
		return nil, err
	}

	return createdPipelineRecord, nil
}

// GetPipelineRecord handles the request to get a pipeline record.
func GetPipelineRecord(ctx context.Context, pipelineRecordID string) (*api.PipelineRecord, error) {
	pipelineRecord, err := pipelineRecordManager.GetPipelineRecord(pipelineRecordID)
	if err != nil {
		return nil, err
	}

	return pipelineRecord, nil
}

// ListPipelineRecords handles the request to list pipeline records.
func ListPipelineRecords(ctx context.Context, projectName, pipelineName string) (api.ListResponse, error) {
	queryParams, err := httputil.QueryParamsFromContext(ctx)
	if err != nil {
		return api.ListResponse{}, err
	}

	pipelineRecords, count, err := pipelineRecordManager.ListPipelineRecords(projectName, pipelineName, queryParams)
	if err != nil {
		return api.ListResponse{}, err
	}

	return httputil.ResponseWithList(pipelineRecords, count), nil
}

// DeletePipelineRecord handles the request to delete a pipeline record.
func DeletePipelineRecord(ctx context.Context, pipelineRecordID string) error {
	if err := pipelineRecordManager.DeletePipelineRecord(pipelineRecordID); err != nil {
		return err
	}

	return nil
}

// UpdatePipelineRecordStatus handles the request to update a pipeline record status, only support to set 'Aborted'
// status for running pipeline records.
func UpdatePipelineRecordStatus(ctx context.Context, pipelineRecordID string) (*api.PipelineRecord, error) {
	pipelineRecord, err := pipelineRecordManager.GetPipelineRecord(pipelineRecordID)
	if err != nil {
		return nil, err
	}

	if pipelineRecord.Status != api.Running {
		log.Infof("The pipeline record %s is not running, can not be aborted, will do no action", pipelineRecord.Name)
		return pipelineRecord, nil
	}

	type RecordStatus struct {
		Status api.Status `json:"status"`
	}

	recordStatus := &RecordStatus{}
	err = contextutil.GetJsonPayload(ctx, recordStatus)
	if err != nil {
		return nil, err
	}

	if recordStatus.Status != api.Aborted {
		err = httperror.ErrorValidationFailed.Format("status", "only support Aborted")
		return nil, err
	}

	e, err := event.GetEvent(pipelineRecord.ID)
	if err != nil {
		err := fmt.Errorf("Unable to find event by versonID %v", pipelineRecord.ID)
		log.Error(err)
		return nil, err
	}

	abortPipelineRecord(e.PipelineRecord)
	// event and pipeline record both will be updated in method "UpdateEvent"
	err = event.UpdateEvent(e)
	if err != nil {
		log.Errorf("update event %s error: %v", e.ID, err)
	}

	return pipelineRecord, nil
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
func GetPipelineRecordLogs(ctx context.Context, projectName, pipelineName, pipelineRecordID, stage, task string, download bool) ([]byte, map[string]string, error) {
	//download, err := httputil.DownloadQueryParamsFromRequest(request)
	//if err != nil {
	//	httputil.ResponseWithError(response, err)
	//	return
	//}

	logs, err := pipelineRecordManager.GetPipelineRecordLogs(pipelineRecordID, stage, task)
	if err != nil {
		return nil, nil, err
	}

	headers := make(map[string]string)
	headers[httputil.HEADER_ContentType] = "text/plain"
	if download {
		logFileName := fmt.Sprintf("%s-%s-%s-%s-log.txt", projectName, pipelineName, stage, pipelineRecordID)
		headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=%s", logFileName)
	}

	return []byte(logs), headers, nil
}

// GetPipelineRecordTestResults handles the request to get pipeline record test results.
func ListPipelineRecordTestResults(ctx context.Context, projectName, pipelineName, pipelineRecordID string) (api.ListResponse, error) {
	results, count, err := pipelineRecordManager.ListPipelineRecordTestResults(pipelineRecordID)
	return httputil.ResponseWithList(results, count), err
}

// GetPipelineRecordTestResult handles the request to get pipeline record test result.
func GetPipelineRecordTestResult(ctx context.Context, projectName, pipelineName, pipelineRecordID, fileName string, download bool) ([]byte, map[string]string, error) {
	result, err := pipelineRecordManager.GetPipelineRecordTestResult(pipelineRecordID, fileName)
	if err != nil {
		return nil, nil, err
	}

	headers := make(map[string]string)
	headers[httputil.HEADER_ContentType] = "text/plain"
	if download {
		resultFileName := fmt.Sprintf("%s-%s-%s-%s", projectName, pipelineName, pipelineRecordID, fileName)
		headers["Content-Disposition"] = fmt.Sprintf("attachment; filename=%s", resultFileName)
	}

	return []byte(result), headers, nil
}

// ReceivePipelineRecordLogStream receives real-time log of pipeline record.
func ReceivePipelineRecordLogStream(ctx context.Context, recordID, stage, task string) error {
	request := contextutil.GetHttpRequest(ctx)
	writer := contextutil.GetHttpResponseWriter(ctx)

	_, err := pipelineRecordManager.GetPipelineRecord(recordID)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to find pipeline record %s for err: %s", recordID, err.Error()))
		return httperror.ErrorContentNotFound.Format(recordID)
	}

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to upgrade websocket for err: %s", err.Error()))
		return httperror.ErrorUnknownInternal.Format(err.Error())
	}
	defer ws.Close()

	if err := pipelineRecordManager.ReceivePipelineRecordLogStream(recordID, stage, task, ws); err != nil {
		log.Error(fmt.Sprintf("Fail to receive log stream for pipeline record %s: %s", recordID, err.Error()))
		return httperror.ErrorUnknownInternal.Format(err.Error())
	}

	return nil
}

// ReceivePipelineRecordTestResult receives test result file from cyclone worker.
func ReceivePipelineRecordTestResult(ctx context.Context, recordID string) error {
	request := contextutil.GetHttpRequest(ctx)
	file, handler, err := request.FormFile("Upload-File")
	if err != nil {
		log.Error(fmt.Sprintf("form file err: %s", err))
		return err
	}
	defer file.Close()

	if err := pipelineRecordManager.ReceivePipelineRecordTestResult(recordID, handler.Filename, file); err != nil {
		log.Error(fmt.Sprintf("Fail to receive log stream for pipeline record %s: %s", recordID, err.Error()))
		return httperror.ErrorUnknownInternal.Format(err.Error())
	}

	return nil

}

// GetPipelineRecordLogStream gets real-time log of pipeline record refering to recordID
func GetPipelineRecordLogStream(ctx context.Context, recordID, stage, task string) error {
	request := contextutil.GetHttpRequest(ctx)
	writer := contextutil.GetHttpResponseWriter(ctx)
	_, err := pipelineRecordManager.GetPipelineRecord(recordID)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to find pipeline record %s for err: %s", recordID, err.Error()))
		return httperror.ErrorContentNotFound.Format(recordID)
	}

	//upgrade HTTP rest API --> socket connection
	ws, err := websocketutil.Upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to upgrade websocket for err: %s", err.Error()))
		return httperror.ErrorUnknownInternal.Format(err.Error())
	}
	defer ws.Close()

	if err := pipelineRecordManager.GetPipelineRecordLogStream(recordID, stage, task, ws); err != nil {
		log.Error(fmt.Sprintf("Unable to get logstream for pipeline record %s for err: %s", recordID, err.Error()))
		return httperror.ErrorUnknownInternal.Format(err.Error())
	}

	return nil
}
