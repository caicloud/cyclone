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

package websocket

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/caicloud/cyclone/kafka"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/satori/go.uuid"
)

const (
	DOCKER_IMAGE_LOG_FLAG = "layer"
)

//AnalysisMessage analysis message receive from the web client.
func AnalysisMessage(dp *DataPacket) bool {
	sReceiveFrom := dp.GetReceiveFrom()
	jsonPacket := dp.GetData()

	defer unmarshalRecover(sReceiveFrom)
	var mapData map[string]interface{}

	if err := json.Unmarshal(jsonPacket, &mapData); err != nil {
		panic(err)
	}

	sAction := mapData["action"].(string)
	MsgHandler, bFoundAction := MsgHandlerMap[sAction]
	if bFoundAction {
		MsgHandler(sReceiveFrom, jsonPacket)
		return true
	}
	log.Infof("ws server recv unknow packet:%s", sAction)
	return false
}

//unmarshalRecover recover the panic of unmarshal message
func unmarshalRecover(sReceiveFrom string) {
	err := recover()
	if nil != err {
		log.Errorf("packet data unmarshal recover: %v", err)
		onCloseSession(sReceiveFrom)
	}
}

//onCloseSession handle something when the session close
func onCloseSession(sSessionID string) {
	isSession := GetSessionList().GetSession(sSessionID)
	if nil == isSession {
		return
	}
	isSession.OnClosed()
}

// MsgHandler is the type for message handler.
type MsgHandler func(sReceiveFrom string, jsonPacket []byte)

//MsgHandlerMap is the map of web client message handler
var MsgHandlerMap = map[string]MsgHandler{
	"watch_log":       watchLogHandler,
	"heart_beat":      heartBeatHandler,
	"worker_push_log": workerPushLogHandler,
}

//watchLogHandler handle the watch log message
func watchLogHandler(sReceiveFrom string, jsonPacket []byte) {
	//Handle watch_log data
	log.Infof("receive watch_log packet")

	pWatchLog := &WatchLogPacket{}
	if err := json.Unmarshal(jsonPacket, &pWatchLog); err != nil {
		panic(err)
	}

	sTopic := CreateTopicName(pWatchLog.Api, pWatchLog.UserId,
		pWatchLog.ServiceId, pWatchLog.VersionId)
	wss := GetSessionList().GetSession(sReceiveFrom).(*WSSession)
	if "start" == pWatchLog.Operation {
		wss.SetTopicEnable(sTopic, true)
		go PushTopic(wss, pWatchLog)
	} else if "stop" == pWatchLog.Operation {
		wss.SetTopicEnable(sTopic, false)
	}

	byrResponse := PacketResponse(pWatchLog.Action, pWatchLog.Id,
		Error_Code_Successful)
	dpPacket := &DataPacket{
		byrFrame:  byrResponse,
		nFrameLen: len(byrResponse),
		sSendTo:   wss.sSessionID,
	}
	wss.Send(dpPacket)
}

//heartBeatHandler handle heart beat message
func heartBeatHandler(sReceiveFrom string, jsonPacket []byte) {
	//Handle heart_beat data
	log.Infof("receive heart_beat packet")
}

//pushLogHandler handle the watch log message
func workerPushLogHandler(sReceiveFrom string, jsonPacket []byte) {
	//Handle watch_log data

	workerPushLog := &WorkerPushLogPacket{}
	if err := json.Unmarshal(jsonPacket, &workerPushLog); err != nil {
		panic(err)
	}

	log.Debugf("Worker log (%s): %s", workerPushLog.Topic, workerPushLog.Log)
	kafka.Produce(workerPushLog.Topic, []byte(workerPushLog.Log))
}

//convertUUID convert - to _ in UUID
func convertUUID(sUUID string) string {
	sConverted := strings.Replace(sUUID, "-", "_", -1)
	return sConverted
}

//CreateTopicName creat topic name as: api__userid__serviceid__versionid
func CreateTopicName(sAPI string, sUserID string, sServiceID string,
	sVersionID string) string {
	sConvertedAPI := convertUUID(sAPI)
	sConvertedUserID := convertUUID(sUserID)
	sConvertedServiceID := convertUUID(sServiceID)
	sConvertedVersionID := convertUUID(sVersionID)
	return fmt.Sprintf("%s__%s__%s__%s", sConvertedAPI, sConvertedUserID,
		sConvertedServiceID, sConvertedVersionID)
}

//PushTopic push log from special topic to web client
func PushTopic(wss *WSSession, pWatchLog *WatchLogPacket) {
	sTopic := CreateTopicName(pWatchLog.Api, pWatchLog.UserId,
		pWatchLog.ServiceId, pWatchLog.VersionId)
	log.Infof("start push %s to %s", sTopic, wss.GetSessionID())

	consumer, err := kafka.NewConsumer(sTopic)
	if nil != err {
		log.Error(err.Error())
		return
	}

	for {
		if nil == wss {
			break
		}

		if false == wss.GetTopicEnable(sTopic) {
			break
		}

		msg, errConsume := consumer.Consume()
		if nil != errConsume {
			if errConsume != kafka.ErrNoData {
				log.Infof("Can't consume %s topic message: %s", sTopic)
				break
			} else {
				continue
			}
		}

		str := string(msg.Value)
		array := strings.Split(str, "\n")
		for _, arr := range array {
			if arr != "\r" && arr != "" {
				if isDockerImageOperationLog(arr) {
					// In order to achieve overlapping the log according to the same layer id,
					// so extracted the layer id from the log into the ID section
					// in the websockect package, then the UI received the webpacket can overlap
					// the log according to the ID.
					tmpss := strings.Split(arr, ":")
					tmps := strings.Split(tmpss[0], " ")
					tmp := tmps[1]
					byrLog := PacketPushLog(pWatchLog.Api, pWatchLog.UserId,
						pWatchLog.ServiceId, pWatchLog.VersionId, arr[6:], tmp)
					dpPacket := &DataPacket{
						byrFrame:  byrLog,
						nFrameLen: len(byrLog),
						sSendTo:   wss.sSessionID,
					}
					wss.Send(dpPacket)
				} else {
					number := uuid.NewV4().String()
					byrLog := PacketPushLog(pWatchLog.Api, pWatchLog.UserId,
						pWatchLog.ServiceId, pWatchLog.VersionId, arr,
						number)
					dpPacket := &DataPacket{
						byrFrame:  byrLog,
						nFrameLen: len(byrLog),
						sSendTo:   wss.sSessionID,
					}
					wss.Send(dpPacket)
				}
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
	log.Infof("stop push %s to %s", sTopic, wss.GetSessionID())
}

// isDockerImageOperationLog check the log whether is the log of pulling or pushing docker image.
func isDockerImageOperationLog(log string) bool {
	return strings.HasPrefix(log, DOCKER_IMAGE_LOG_FLAG)
}
