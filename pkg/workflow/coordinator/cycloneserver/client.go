package cycloneserver

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
)

const (
	cycloneAPIVersion = "/apis/v1alpha1"

	apiPathForLogStream = "/workflowruns/%s/streamlogs"
)

// Client ...
type Client interface {
	PushLogStream(ns, workflowrun, stage, container string, reader io.Reader, close chan struct{}) error
}

type client struct {
	baseURL string
	client  *http.Client
}

// NewClient ...
func NewClient(cycloneServer string) Client {
	baseURL := strings.TrimRight(cycloneServer, "/")
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}

	return &client{
		baseURL: baseURL,
		client:  http.DefaultClient,
	}
}

// PushLogStream ...
func (c *client) PushLogStream(ns, workflowrun, stage, container string, reader io.Reader, close chan struct{}) error {
	path := fmt.Sprintf(apiPathForLogStream, workflowrun)
	host := strings.TrimPrefix(c.baseURL, "http://")
	host = strings.TrimPrefix(host, "https://")
	requestURL := url.URL{
		Host:     host,
		Path:     cycloneAPIVersion + path,
		RawQuery: fmt.Sprintf("namespace=%s&stage=%s&container=%s", ns, stage, container),
		Scheme:   "ws",
	}

	log.Infof("Path: %s", requestURL.String())

	header := http.Header{
		"Connection":            []string{"Upgrade"},
		"Upgrade":               []string{"websocket"},
		"Host":                  []string{host},
		"Sec-Websocket-Key":     []string{"SGVsbG8sIHdvcmxkIQ=="},
		"Sec-Websocket-Version": []string{"13"},
	}
	filteredHeader := websocketutil.FilterHeader(header)

	ws, _, err := websocket.DefaultDialer.Dial(requestURL.String(), filteredHeader)
	if err != nil {
		log.Errorf("Fail to new the WebSocket connection as %s", err.Error())
		return err
	}
	defer ws.Close()

	return watchLogs(ws, reader, close)
}

func watchLogs(ws *websocket.Conn, reader io.Reader, close chan struct{}) error {
	//logFile, err := os.Open(filePath)
	//if err != nil {
	//	log.Error(err.Error())
	//	return err
	//}
	//defer logFile.Close()

	buf := bufio.NewReader(reader)
	ticker := time.NewTicker(10 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			line, errRead := buf.ReadBytes('\n')
			if errRead != nil {
				if errRead == io.EOF {
					continue
				}
				log.Errorf("watch log file errs: %v", errRead)
				err := ws.WriteMessage(websocket.CloseMessage, []byte(errRead.Error()))
				if err != nil {
					log.Warningf("write close message error:%v", err)
				}
				return errRead
			}
			err := ws.WriteMessage(websocket.TextMessage, line)
			if err != nil {
				log.Warningf("write text message error:%v", err)
			}
		case <-close:
			log.Info("Close the watch of log file")
			return nil
		}
	}
}
