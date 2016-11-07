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
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/caicloud/cyclone/pkg/log"
	"golang.org/x/net/websocket"
)

//StartServer start the websocket server
func StartServer() (err error) {
	scServerConfig := GetConfig()
	log.Infof("Start Websocket Server at Port:%d", scServerConfig.Port)

	hsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", scServerConfig.Port),
		Handler: websocket.Handler(webMessageHandle),
	}
	return hsServer.ListenAndServe()
}

//webMessageHandle handle the message receive from web client
func webMessageHandle(wsConn *websocket.Conn) {
	wssSession, err := CreateWSSession(wsConn)
	if err != nil {
		wsConn.Close()
		log.Error(err.Error())
		return
	}
	defer wssSession.OnClosed()

	nIdleCheckInterval := GetConfig().IdleCheckInterval
	for {
		//timeout
		err = wsConn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(nIdleCheckInterval)))
		if err != nil {
			log.Error(err.Error())
			return
		}

		//receive msg
		var byrMesssage []byte
		err = websocket.Message.Receive(wsConn, &byrMesssage)
		if err == nil {
			if receiveData(wssSession, byrMesssage) {
				wssSession.UpdateActiveTime()
			}
			continue
		}

		//error and timeout handler
		e, ok := err.(net.Error)
		if !ok || !e.Timeout() {
			log.Error(err.Error())
			return
		} else {
			if wssSession.SessionTimeoverCheck() {
				return
			}
		}
	}
}

//receiveData post the received data to sessions
func receiveData(isSession ISession, byrMesssage []byte) bool {
	dp := &DataPacket{}
	dp.SetData(byrMesssage)
	dp.SetLength(len(byrMesssage))
	sSessionID := isSession.GetSessionID()
	dp.SetReceiveFrom(sSessionID)
	return isSession.OnReceive(dp)
}
