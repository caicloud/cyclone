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

package common

import (
	"encoding/json"
	"fmt"
	"time"

	gwebsocket "golang.org/x/net/websocket"
)

var (
	// ErrUnfoundLog is the error type for "log not found."
	ErrUnfoundLog = fmt.Errorf("log not found")
)

const (
	// APICreateVersion is the API for version creation.
	APICreateVersion = "create-version"
	// LogServerOrigin is the host of log server.
	LogServerOrigin = "http://localhost/"
	// LogServerURL is the URL of log server.
	LogServerURL = "ws://localhost:8000/ws"
	// StartOperation is the operation name for starting.
	StartOperation = "start"
	// StopOperation is the operation name for stoping.
	StopOperation = "stop"
	// TimeOutPushLog is the timeout for log pusing.
	TimeOutPushLog = 5
	// TimeOutResponse is the timeout for response.
	TimeOutResponse = 10
	// ReadMsgBufferSize is the default size of buffer for reading messages.
	ReadMsgBufferSize = 32768
)

// DialLogServer dials LogServerURL for connection.
func DialLogServer() (ws *gwebsocket.Conn, err error) {
	return gwebsocket.Dial(LogServerURL, "", LogServerOrigin)
}

// SendMsgToLogServer sends messages to log server directly.
func SendMsgToLogServer(ws *gwebsocket.Conn, msg []byte) error {
	if _, err := ws.Write(msg); err != nil {
		return err
	}
	return nil
}

// ReadMsgFromLogServer reads messages from log server.
func ReadMsgFromLogServer(ws *gwebsocket.Conn, timeout int) ([]byte, error) {
	var msg = make([]byte, ReadMsgBufferSize)
	var n int
	var err error

	ws.SetReadDeadline(time.Now().Add(time.Second * time.Duration(timeout)))
	if n, err = ws.Read(msg); err != nil {
		return nil, err
	}

	return msg[:n], nil
}

func analysisMsg(msg []byte) (mapData map[string]interface{}, err error) {
	if err = json.Unmarshal(msg, &mapData); err != nil {
		return nil, err
	}
	return mapData, nil
}
