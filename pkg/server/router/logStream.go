package router

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/caicloud/cyclone/api/rest"
	"github.com/caicloud/cyclone/event"
	"github.com/caicloud/cyclone/kafka"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/websocket"
	"github.com/emicklei/go-restful"
	socket "github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 3 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var upgrader = socket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func getVersionLogStream(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter("project")
	pipelineName := request.PathParameter("pipeline")
	recordID := request.PathParameter("recordID")

	pipeline, _, err := rest.FindServiceAndVersion(recordID)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		message := fmt.Sprintf("Unable to find pipeline record: %v", recordID)
		log.ErrorWithFields(message, log.Fields{"project": projectName, "pipeline": pipelineName, "error": err})
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	}

	ws, err := upgrader.Upgrade(response.ResponseWriter, request.Request, nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	writerLogStream(ws, pipeline.ServiceID, recordID, pipeline.UserID)
}

func writerLogStream(ws *socket.Conn, pipelineID string, recordID string, userID string) {
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()

	// Read the message from standard input.
	logstream := make(chan []byte, 10)
	go getLogStreamFromKafka(logstream, pipelineID, recordID, userID)

	for {
		select {
		case logFragment, ok := <-logstream:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				ws.WriteMessage(socket.CloseMessage, []byte{})
				return
			}

			if err := ws.WriteMessage(socket.TextMessage, []byte(logFragment)); err != nil {
				log.Error(err.Error())
				return
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(socket.PingMessage, []byte{}); err != nil {
				log.Error(err.Error())
				return
			}
		}
	}
}

func getLogStreamFromKafka(logstream chan []byte, pipelineID string, recordID string, userID string) {
	sTopic := websocket.CreateTopicName(string(event.CreateVersionOps), userID, pipelineID, recordID)

	consumer, err := kafka.NewConsumer(sTopic)
	if nil != err {
		log.Error(err.Error())
		return
	}

	for {
		msg, errConsume := consumer.Consume()
		if errConsume != nil {
			if errConsume != kafka.ErrNoData {
				log.Infof("Can't consume %s topic message: %s", sTopic)
				break
			} else {
				continue
			}
		}

		processKafkaMsg(logstream, string(msg.Value))

		time.Sleep(time.Millisecond * 100)
	}
}

func processKafkaMsg(logstream chan []byte, msg string) {
	msgFragments := strings.Split(msg, "\n")

	var logFragment []byte
	for _, msgFragment := range msgFragments {
		logFragment = parseLogFragment(msgFragment)
		if len(logFragment) == 0 {
			continue
		}
		logstream <- logFragment
	}
}

func parseLogFragment(msgFragment string) []byte {
	var logFragment string
	if msgFragment != "\r" && msgFragment != "" {
		if websocket.IsDockerImageOperationLog(msgFragment) {
			logFragment = msgFragment[6:]
		} else {
			logFragment = msgFragment
		}
	}
	return []byte(logFragment)
}
