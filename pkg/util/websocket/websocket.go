/*
Copyright 2018 caicloud authors. All rights reserved.

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
	"net/http"
	"time"

	socket "github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 3 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10
)

var Upgrader = socket.Upgrader{
	//disable origin check
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// FilterHeader filters the headers for upgrading the HTTP server connection to the WebSocket protocol.
func FilterHeader(header http.Header) http.Header {
	newHeader := http.Header{}
	for k, vs := range header {
		switch {
		case k == "Upgrade" ||
			k == "Connection" ||
			k == "Sec-Websocket-Key" ||
			k == "Sec-Websocket-Version" ||
			k == "Sec-Websocket-Extensions":
		default:
			newHeader[k] = vs
		}
	}
	return newHeader
}
