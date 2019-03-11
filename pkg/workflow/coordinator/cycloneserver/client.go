package cycloneserver

import (
	"bufio"
	"bytes"
	"encoding/json"
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

// do sends the request to Cyclone and returns an HTTP response.
func (c *client) do(method, relativePath string, bodyObject interface{}) (*http.Response, error) {
	url := c.baseURL + cycloneAPIVersion + relativePath
	log.Infof("Request for Cyclone server: %s %s", method, url)

	var body io.Reader
	if bodyObject != nil {
		bodyBytes, err := json.Marshal(bodyObject)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		log.Errorf(err.Error())
		return nil, err
	}

	return resp, nil
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
				ws.WriteMessage(websocket.CloseMessage, []byte(errRead.Error()))
				return errRead
			}
			ws.WriteMessage(websocket.TextMessage, line)
		case <-close:
			log.Info("Close the watch of log file")
			return nil
		}
	}
}
