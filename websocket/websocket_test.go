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
	"testing"
	"time"

	"github.com/satori/go.uuid"
	wslib "golang.org/x/net/websocket"
)

// TestStartServer test start websocket server and dail it.
func TestStartServer(t *testing.T) {
	// Load websocket server config.
	err := LoadServerConfig()
	if err != nil {
		t.Errorf("Load websocket server config error: %v.", err)
	}

	// Start websocket server.
	go func() {
		err := StartServer()
		if err != nil {
			t.Errorf("Start websocket server error: %v.", err)
		}
	}()
	time.Sleep(5 * time.Millisecond)
}

// dialTestServer dail to local websocket test server, return websocket handler or err.
func dialTestServer() (*wslib.Conn, error) {
	origin := "http://127.0.0.1/"
	url := "ws://127.0.0.1:8000/ws"
	return wslib.Dial(url, "", origin)
}

// TestWebMessageHandle test websocket server handle message
func TestWebMessageHandle(t *testing.T) {
	// Dail to server.
	ws, err := dialTestServer()
	if err != nil {
		t.Errorf("Dail websocket server error: %v.", err)
	}

	go webMessageHandle(ws)
	ws.Write([]byte("{\"action\":\"watch_log\"}"))
	ws.Write([]byte("test message"))

	// Close the connection.
	err = ws.Close()
	if err != nil {
		t.Errorf("Close connection to websocket server error: %v.", err)
	}
}

// TestSessionList test session list functions
func TestSessionList(t *testing.T) {
	// Get a initial session list.
	sessionList := GetSessionList()
	if nil == sessionList {
		t.Error("Get session list err")
	}

	// New a websocket session and add to session list.
	wssSession := &WSSession{
		nLastActive: time.Now().Unix(),
		sSessionID:  uuid.NewV4().String(),
	}
	sessionList.addOnlineSession(wssSession)

	// Get an added session from session list.
	sAdded := sessionList.GetSession(wssSession.sSessionID)
	if sAdded == nil {
		t.Error("Get added session err.")
	}

	// Remove an added session from session list.
	sessionList.removeSession(wssSession.sSessionID)

	// Get an removed session from session list.
	sRemoved := sessionList.GetSession(wssSession.sSessionID)
	if sRemoved != nil {
		t.Error("Get removed session err.")
	}
}

// TestWSSession test websocket session functions
func TestWSSession(t *testing.T) {
	// Dail to server.
	ws, err := dialTestServer()
	if err != nil {
		t.Errorf("Dail websocket server error: %v.", err)
	}

	// Create websocket session.
	wsSession, err := CreateWSSession(ws)
	if err != nil {
		t.Errorf("Create websocket session err: %v.", err)
	}

	// Get websocket session ID.
	if wsSession.GetSessionID() == "" {
		t.Error("Get websocket session ID err.")
	}

	// Websocket session receive data packet.
	dp := generateTestDataPacket()
	if wsSession.OnReceive(dp) == false {
		t.Error("Websocket session receive data packet err.")
	}

	// Test websocket session active flag.
	wsSession.UpdateActiveTime()
	if wsSession.SessionTimeoverCheck() == true {
		t.Error("Test websocket session active flag err.")
	}

	// Test websocket session topic switch.
	wsSession.SetTopicEnable("topic", true)
	if wsSession.GetTopicEnable("topic") != true {
		t.Error("Websocket session set topic switch err.")
	}

	// Close the connection.
	err = ws.Close()
	if err != nil {
		t.Errorf("Close connection to websocket server error: %v.", err)
	}
}

// generateTestDataPacket generate a test data packet.
func generateTestDataPacket() *DataPacket {
	var dp DataPacket

	data := "{\"action\":\"heart_beat\"}"
	dp.SetData([]byte(data))
	dp.SetLength(len(data))
	dp.SetReceiveFrom("receive")
	dp.SetSendTo("send")

	return &dp
}

// TestProtocol test packet frame functions.
func TestProtocol(t *testing.T) {
	var packet []byte

	sApi := "api"
	sUserId := "user"
	sServiceId := "service"
	sVersion := "version"
	sOperation := "operation"
	sId := "id"
	sLog := "log"
	sResponse := "ok"
	nErrorCode := 0

	// Packet watch log frame.
	packet = PacketWatchLog(sApi, sUserId, sServiceId, sVersion, sOperation, sId)
	if string(packet) != `{"action":"watch_log","api":"api","user_id":"user","service_id":"service","version_id":"version","operation":"operation","id":"id"}` {
		t.Error("Packet watch log frame err")
	}

	// Packet heart beat frame.
	packet = PacketHeartBeat(sId)
	if string(packet) != `{"action":"heart_beat","id":"id"}` {
		t.Error("Packet heart beat frame err")
	}

	// Packet push log frame.
	packet = PacketPushLog(sApi, sUserId, sServiceId, sVersion, sLog, sId)
	if string(packet) != `{"action":"push_log","api":"api","user_id":"user","service_id":"service","version_id":"version","log":"log","id":"id"}` {
		t.Error("Packet push log frame err")
	}

	// Packet response frame.
	packet = PacketResponse(sResponse, sId, nErrorCode)
	if string(packet) != `{"action":"response","response":"ok","id_ack":"id","error_code":0,"error_msg":"successful"}` {
		t.Error("Packet response frame err")
	}
}
