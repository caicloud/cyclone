package websocket

import (
	"net/http"
	"time"

	socket "github.com/gorilla/websocket"
)

const (
	// WriteWait defines time allowed to write a message to the peer.
	WriteWait = 3 * time.Second

	// PongWait defines time allowed to read the next pong message from the peer.
	PongWait = 30 * time.Second

	// PingPeriod defines send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10
)

// Upgrader ...
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
