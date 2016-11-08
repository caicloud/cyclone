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
)

// WatchLogPacket is the type for watch_log packet.
type WatchLogPacket struct {
	Action    string `json:"action"`
	Api       string `json:"api"`
	UserId    string `json:"user_id"`
	ServiceId string `json:"service_id"`
	VersionId string `json:"version_id"`
	Operation string `json:"operation"`
	Id        string `json:"id"`
}

// WorkerPushLogPacket is the type for worker_push_log packet.
type WorkerPushLogPacket struct {
	Action string `json:"action"`
	Topic  string `json:"topic"`
	Log    string `json:"log"`
}

// PushLogPacket is the type for push_log packet.
type PushLogPacket struct {
	Action    string `json:"action"`
	Api       string `json:"api"`
	UserId    string `json:"user_id"`
	ServiceId string `json:"service_id"`
	VersionId string `json:"version_id"`
	Log       string `json:"log"`
	Id        string `json:"id"`
}

// HeartBeatPacket is the type for heart_beat packet.
type HeartBeatPacket struct {
	Action string `json:"action"`
	Id     string `json:"id"`
}

// ResponsePacket is the type for response packet.
type ResponsePacket struct {
	Action    string `json:"action"`
	Response  string `json:"response"`
	IdAck     string `json:"id_ack"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

const (
	Error_Code_Successful = 0
	Error_Code_Failure    = 4001
)

// ErrorMsgMap is the map from status code to error message.
var ErrorMsgMap = map[int]string{
	0:    "successful",
	4001: "failure",
}

//PacketWatchLog packet the data frame of watch log
func PacketWatchLog(SAPI string, sUserID string, sServiceID string,
	sVersion string, sOperation string, sID string) []byte {
	structData := &WatchLogPacket{
		Action:    "watch_log",
		Api:       SAPI,
		UserId:    sUserID,
		ServiceId: sServiceID,
		VersionId: sVersion,
		Operation: sOperation,
		Id:        sID}
	jsonData, _ := json.Marshal(structData)
	return jsonData
}

//PacketPushLog packet the data frame of push log
func PacketPushLog(SAPI string, sUserID string, sServiceID string,
	sVersion string, sLog string, sID string) []byte {
	structData := &PushLogPacket{
		Action:    "push_log",
		Api:       SAPI,
		UserId:    sUserID,
		ServiceId: sServiceID,
		VersionId: sVersion,
		Log:       sLog,
		Id:        sID}
	jsonData, _ := json.Marshal(structData)
	return jsonData
}

//PacketHeartBeat packet the data frame of heart beat
func PacketHeartBeat(sID string) []byte {
	structData := &HeartBeatPacket{
		Action: "heart_beat",
		Id:     sID}
	jsonData, _ := json.Marshal(structData)
	return jsonData
}

//PacketResponse packet the data frame of response
func PacketResponse(sResponse string, sIDAck string, nErrorCode int) []byte {
	structData := &ResponsePacket{
		Action:    "response",
		Response:  sResponse,
		IdAck:     sIDAck,
		ErrorCode: nErrorCode,
		ErrorMsg:  ErrorMsgMap[nErrorCode]}
	jsonData, _ := json.Marshal(structData)
	return jsonData
}
