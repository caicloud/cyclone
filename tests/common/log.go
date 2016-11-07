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

	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/websocket"
	"github.com/satori/go.uuid"
	gwebsocket "golang.org/x/net/websocket"
)

var (
	ErrUnfoundLog = fmt.Errorf("unfound log")
)

const (
	ApiCreateVersion  = "create-version"
	LogServerOrigin   = "http://localhost/"
	LogServerUrl      = "ws://localhost:8000/ws"
	StartOperation    = "start"
	StopOperation     = "stop"
	TimeOutPushLog    = 5
	TimeOutResponse   = 10
	ReadMsgBufferSize = 32768
)

func DialLogServer() (ws *gwebsocket.Conn, err error) {
	return gwebsocket.Dial(LogServerUrl, "", LogServerOrigin)
}

func SendMsgToLogServer(ws *gwebsocket.Conn, msg []byte) error {
	if _, err := ws.Write(msg); err != nil {
		return err
	}
	return nil
}

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

func WatchLog(ws *gwebsocket.Conn, apiName string, userID string,
	serviceID string, versionID string) (err error) {
	//start watch log
	msg := websocket.PacketWatchLog(apiName, userID, serviceID,
		versionID, StartOperation, uuid.NewV4().String())
	if err = SendMsgToLogServer(ws, msg); err != nil {
		return err
	}

	var rec = make([]byte, ReadMsgBufferSize)
	//start watch log response
	if rec, err = ReadMsgFromLogServer(ws, TimeOutResponse); err != nil {
		return err
	}

	var mapData map[string]interface{}
	if mapData, err = analysisMsg(rec); err != nil {
		return err
	}
	if "reponse" != mapData["action"].(string) && "successful" != mapData["error_msg"] {
		log.Errorf("Receive unexcept data: %s", string(rec))
		return fmt.Errorf("start watch log error")
	}

	//watch log
	bHasReceiveLog := false
	for {
		if rec, err = ReadMsgFromLogServer(ws, TimeOutPushLog); err != nil {
			break
		}

		if mapData, err = analysisMsg(rec); err != nil {
			return err
		}

		if "push_log" == mapData["action"].(string) {
			bHasReceiveLog = true
		}
	}
	if false == bHasReceiveLog {
		return ErrUnfoundLog
	}

	//stop watch log
	msg = websocket.PacketWatchLog(apiName, userID, serviceID,
		versionID, StopOperation, uuid.NewV4().String())
	if err = SendMsgToLogServer(ws, msg); err != nil {
		return err
	}

	if rec, err = ReadMsgFromLogServer(ws, TimeOutResponse); err != nil {
		return err
	}

	if mapData, err = analysisMsg(rec); err != nil {
		return err
	}
	if "reponse" != mapData["action"].(string) && "successful" != mapData["error_msg"] {
		return fmt.Errorf("stop watch log error")
	}
	return nil
}

func StopWatchLog(ws *gwebsocket.Conn, apiName string, userID string,
	serviceID string, versionID string) {
	msg := websocket.PacketWatchLog(apiName, userID, serviceID,
		versionID, StopOperation, uuid.NewV4().String())
	SendMsgToLogServer(ws, msg)
}
