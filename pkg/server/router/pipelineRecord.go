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
	"strings"
	"time"

	"github.com/zoumo/logdog"

	oldapi "github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/event"
	"github.com/caicloud/cyclone/kafka"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"

	httputil "github.com/caicloud/cyclone/pkg/util/http"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/websocket"
	"github.com/emicklei/go-restful"
	socket "github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 3 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = socket.Upgrader{
	//disable origin check
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

//getPipelineRecordLogStream get real-time log of pipeline record refering to recordID
func getPipelineRecordLogStream(request *restful.Request, response *restful.Response) {
	recordID := request.PathParameter("recordID")

	pipeline, _, err := store.FindServiceAndVersion(recordID)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to find pipeline record %s for err: %s", recordID, err.Error()))
		httputil.ResponseWithError(response, httperror.ErrorContentNotFound.Format(recordID))
		return
	}

	//upgrade HTTP rest API --> socket connection
	ws, err := upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to upgrade websocket for err: %s", err.Error()))
		httputil.ResponseWithError(response, httperror.ErrorUnknownInternal.Format(err.Error()))
		return
	}

	writerLogStream(ws, pipeline.ServiceID, recordID, pipeline.UserID)
}

//writerLogStream write logfragment received from chan logstream to websocket connection
func writerLogStream(ws *socket.Conn, pipelineID string, recordID string, userID string) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()

	// load log fragment from kafka --> logstream
	logstream := make(chan []byte, 10)
	go getLogStreamFromKafka(logstream, pipelineID, recordID, userID)

	for {
		select {
		case logFragment, ok := <-logstream:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				ws.WriteMessage(socket.CloseMessage, []byte{})
				return
			}

			if err := ws.WriteMessage(socket.TextMessage, []byte(logFragment)); err != nil {
				log.Error(err.Error())
				return
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(socket.PingMessage, []byte{}); err != nil {
				log.Error(err.Error())
				return
			}
		}
	}
}

//getLogStreamFromKafka loads msg from kafka
func getLogStreamFromKafka(logstream chan []byte, pipelineID string, recordID string, userID string) {
	sTopic := websocket.CreateTopicName(string(event.CreateVersionOps), userID, pipelineID, recordID)

	consumer, err := kafka.NewConsumer(sTopic)
	if nil != err {
		log.Error(err.Error())
		return
	}

	for {
		msg, errConsume := consumer.Consume()
		if errConsume != nil {
			if errConsume != kafka.ErrNoData {
				log.Infof("Can't consume %s topic message: %s", sTopic)
				break
			} else {
				continue
			}
		}

		processKafkaMsg(logstream, string(msg.Value))

		time.Sleep(time.Millisecond * 100)
	}
}

//processKafkaMsg converts kafka msg to log fragment and sends them to chan logstream
func processKafkaMsg(logstream chan []byte, msg string) {
	msgFragments := strings.Split(msg, "\n")

	var logFragment []byte
	for _, msgFragment := range msgFragments {
		logFragment = parseLogFragment(msgFragment)
		if len(logFragment) == 0 {
			continue
		}
		logstream <- logFragment
	}
}

func parseLogFragment(msgFragment string) []byte {
	var logFragment string
	if msgFragment != "\r" && msgFragment != "" {
		if websocket.IsDockerImageOperationLog(msgFragment) {
			//omite control characters for folding
			logFragment = msgFragment[6:]
		} else {
			logFragment = msgFragment
		}
	}
	return []byte(logFragment)
}
