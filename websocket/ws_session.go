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
	"sync"
	"time"

	"github.com/caicloud/cyclone/pkg/log"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"
)

// WSSession is the type for web socket session.
type WSSession struct {
	wcConnection *websocket.Conn
	// receive last vaild message from web client seconds
	nLastActive int64
	sSessionID  string
	//sync log for tapic enable flags
	sync.RWMutex
	// map for tapic enable flag
	mapTopicEnable map[string]bool
}

//GetSessionID get the session id
func (wss *WSSession) GetSessionID() string {
	return wss.sSessionID
}

//OnClosed handle something when the session close
//remove session from sessionlist
//close websocket link
//clear member var
func (wss *WSSession) OnClosed() {
	if wss.wcConnection != nil {
		GetSessionList().removeSession(wss.sSessionID)
		wss.wcConnection.Close()
		wss.sSessionID = ""
		wss.wcConnection = nil
		wss.ClearTopicEnable()
	}
}

// OnStart handles something when the session start
func (wss *WSSession) OnStart(sSessionID string) {
}

//OnReceive handle something when the session receive message
//analysis the message
func (wss *WSSession) OnReceive(iData IDataPacket) bool {
	log.Debugf("Session(%s) recv: %s", wss.sSessionID, string(iData.GetData()))
	return AnalysisMessage(iData.(*DataPacket))
}

// UpdateActiveTime updates active time of this session.
func (wss *WSSession) UpdateActiveTime() {
	wss.nLastActive = time.Now().Unix()
}

//Send sends data packet to web client
func (wss *WSSession) Send(iData IDataPacket) {
	if wss.wcConnection != nil {
		log.Debugf("WS Send to client(%s): %s", wss.GetSessionID(), iData.GetData())
		err := websocket.Message.Send(wss.wcConnection, string(iData.GetData()))
		if err != nil {
			log.Error(err.Error())
		}
	}
}

//SessionTimeoverCheck check timeout if unreceive vaild message from web cliet
func (wss *WSSession) SessionTimeoverCheck() bool {
	nIdleTimeOut := GetConfig().IdleSessionTimeOut
	nCurTime := time.Now().Unix()
	if nCurTime-wss.nLastActive > nIdleTimeOut {
		log.Infof("Close Session(%s),Idle TimeOut: %d",
			wss.sSessionID, nCurTime-wss.nLastActive)
		return true
	}
	return false
}

//SetTopicEnable set sepcial topic enable flag
func (wss *WSSession) SetTopicEnable(sTopic string, bEnable bool) {
	wss.Lock()
	defer wss.Unlock()
	log.Info("set topic ", sTopic, " ", bEnable)
	wss.mapTopicEnable[sTopic] = bEnable
}

//GetTopicEnable get sepcial topic enable flag
func (wss *WSSession) GetTopicEnable(sTopic string) bool {
	wss.RLock()
	defer wss.RUnlock()
	bEnable, bFound := wss.mapTopicEnable[sTopic]
	if bFound {
		return bEnable
	}
	return false
}

//ClearTopicEnable clear the map of tapic enable flags
func (wss *WSSession) ClearTopicEnable() {
	wss.Lock()
	defer wss.Unlock()
	for sTopic := range wss.mapTopicEnable {
		wss.mapTopicEnable[sTopic] = false
	}
}

//CreateWSSession create new session and incert it to sessionlist
func CreateWSSession(wcConn *websocket.Conn) (wssSession *WSSession, err error) {
	uuid := uuid.NewV4()

	wssSession = &WSSession{
		wcConnection:   wcConn,
		nLastActive:    time.Now().Unix(),
		sSessionID:     uuid.String(),
		mapTopicEnable: make(map[string]bool),
	}
	GetSessionList().addOnlineSession(wssSession)
	return wssSession, err
}
