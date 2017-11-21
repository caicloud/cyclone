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
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/event"
	"github.com/caicloud/cyclone/kafka"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
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
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// getVersionLog finds an version log from versionID.
//
// GET: /api/v0.1/:uid/versions/:versionID/logs
//
// RESPONSE: (VersionLogGetResponse)
//  {
//    "logs": (string) log
//    "error_msg": (string) set IFF the request fails.
//  }
func getVersionLog(request *restful.Request, response *restful.Response) {
	versionID := request.PathParameter("version_id")
	userID := request.PathParameter("user_id")

	var getResponse api.VersionLogGetResponse

	ds := store.NewStore()
	defer ds.Close()
	result, err := ds.FindVersionLogByVersionID(versionID)
	if err != nil {
		message := fmt.Sprintf("Unable to find version log by versionID %v", versionID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		getResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusNotFound, getResponse)
		return
	} else {
		getResponse.Logs = result.Logs
	}

	response.WriteEntity(getResponse)
}

// createVersionLog creates an version log.
//
// POST: /api/v0.1/:uid/versions/:versionID/logs
//
// PAYLOAD (Version):
//   {
//     "logs": (string) a short description of the logs
//     "version_id": (string) id with the version
//   }
//
// RESPONSE: (VersionLogGetResponse)
//  {
//    "error_msg": (string) set IFF the request fails.
//  }
func createVersionLog(request *restful.Request, response *restful.Response) {
	versionID := request.PathParameter("version_id")
	userID := request.PathParameter("user_id")

	versionLog := api.VersionLog{}
	err := request.ReadEntity(&versionLog)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	var createResponse api.VersionLogCreateResponse

	ds := store.NewStore()
	defer ds.Close()
	result, err := ds.NewVersionLogDocument(&versionLog)
	if err != nil {
		message := fmt.Sprintf("Unable to create version log by versionID %v", versionID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		createResponse.ErrorMessage = message
	} else {
		createResponse.LogID = result
	}

	response.WriteEntity(createResponse)
}

func getVersionLogStream(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter("project")
	pipelineName := request.PathParameter("pipeline")
	recordID := request.PathParameter("recordID")

	pipeline, _, err := findServiceAndVersion(recordID)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		message := fmt.Sprintf("Unable to find pipeline record: %v", recordID)
		log.ErrorWithFields(message, log.Fields{"project": projectName, "pipeline": pipelineName, "error": err})
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	}

	ws, err := upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	writerLogStream(ws, pipeline.ServiceID, recordID, pipeline.UserID)
}

func writerLogStream(ws *socket.Conn, pipelineID string, recordID string, userID string) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()

	// Read the message from standard input.
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
			logFragment = msgFragment[6:]
		} else {
			logFragment = msgFragment
		}
	}
	return []byte(logFragment)
}
