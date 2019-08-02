package websocket

import (
	"bufio"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/gorilla/websocket"
	socket "github.com/gorilla/websocket"
)

const (
	// WriteWait defines time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

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

// SendStream sends stream from reader to a remote websocket
func SendStream(server string, reader io.Reader, close <-chan struct{}) error {
	if !strings.Contains(server, "://") {
		server = "ws://" + server
	}
	requestURL, err := url.Parse(server)
	if err != nil {
		return err
	}
	requestURL.Scheme = "ws"
	log.Info("Request url:", requestURL.String())
	header := http.Header{
		"Host": []string{requestURL.Host},
	}

	ws, _, err := websocket.DefaultDialer.Dial(requestURL.String(), header)
	if err != nil {
		log.Errorf("Fail to new the WebSocket connection as %s", err.Error())
		return err
	}
	defer ws.Close()

	return Send(ws, reader, close)
}

// Send sends stream from reader by websocket
func Send(ws *websocket.Conn, reader io.Reader, close <-chan struct{}) error {
	buf := bufio.NewReader(reader)
	err := Write(ws, buf)
	if err != nil {
		log.Error("websocket writer error:", err)
	}

	return err
}

// ReadBytes reads and returns a single byte.
// If no byte is available, returns an error.
type ReadBytes interface {
	ReadBytes(delim byte) ([]byte, error)
}

// Write writes message from reader to websocket
func Write(ws *websocket.Conn, reader ReadBytes) error {
	pingTicker := time.NewTicker(PingPeriod)
	sendTicker := time.NewTicker(10 * time.Millisecond)
	defer func() {
		log.Info("close ticket and websocket")
		pingTicker.Stop()
		sendTicker.Stop()
		ws.Close()
	}()

	for {
		select {
		case <-pingTicker.C:
			err := ws.SetWriteDeadline(time.Now().Add(WriteWait))
			if err != nil {
				log.Warning("set write deadline error:", err)
			}
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Warning("write ping message error:", err)
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
					return nil
				}
				return err
			}
		case <-sendTicker.C:
			// With buf.ReadBytes, when err is not nil (often io.EOF), line is not guaranteed to be empty,
			// it holds data before the error occurs.
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				log.Warning("folder reader read bytes error:", err)
				err = ws.SetWriteDeadline(time.Now().Add(WriteWait))
				if err != nil {
					log.Warning("set write deadline error:", err)
				}
				err = ws.WriteMessage(websocket.CloseMessage, []byte("Interval error happens, TERMINATE"))
				if err != nil {
					log.Warning("write close message error:", err)
				}
				break
			}

			if len(line) > 0 {
				err = ws.SetWriteDeadline(time.Now().Add(WriteWait))
				if err != nil {
					log.Warning("set write deadline error:", err)
				}
				err = ws.WriteMessage(websocket.TextMessage, line)
				if err != nil {
					log.Warning("write text message error:", err)
					if !websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
						return nil
					}
					return err
				}
			}
		}
	}
}
