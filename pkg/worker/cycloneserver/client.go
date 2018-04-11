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

package cycloneserver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/gorilla/websocket"

	"github.com/caicloud/cyclone/pkg/api"
	. "github.com/caicloud/cyclone/pkg/util/http/errors"
	websocketutil "github.com/caicloud/cyclone/pkg/util/websocket"
)

const (
	cycloneAPIVersion = "/api/v1"

	apiPathForEvent     = "/events/%s"
	apiPathForLogStream = "/projects/%s/pipelines/%s/records/%s/stagelogstream"
)

type CycloneServerClient interface {
	GetEvent(id string) (*api.Event, error)
	SendEvent(event *api.Event) error
	PushLogStream(project, pipeline, recordID string, stage api.PipelineStageName, task string, filePath string, close chan struct{}) error
}

type client struct {
	baseURL string
	client  *http.Client
}

func NewClient(cycloneServer string) CycloneServerClient {
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

func (c *client) SendEvent(event *api.Event) error {
	id := event.ID
	path := fmt.Sprintf(apiPathForEvent, id)
	resp, err := c.do(http.MethodPut, path, event)
	if err != nil {
		return ErrorUnknownInternal.Format(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorUnknownInternal.Format(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrorContentNotFound.Format(fmt.Sprintf("event %s", id))
	}

	log.Errorf("Set event %s from Cyclone server with error %s", id, string(body))
	return ErrorUnknownInternal.Format(body)
}

func (c *client) GetEvent(id string) (*api.Event, error) {
	path := fmt.Sprintf(apiPathForEvent, id)
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, ErrorUnknownInternal.Format(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Format(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		log.Infof("Event %s is got from Cyclone server", id)
		event := &api.Event{}
		if err := json.Unmarshal(body, event); err != nil {
			log.Errorf("Fail to unmarshal event %s as %s", id, err.Error())
			return nil, err
		}

		return event, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrorContentNotFound.Format(fmt.Sprintf("event %s", id))
	}

	log.Errorf("Get event %s from Cyclone server with error %s", id, body)
	return nil, ErrorUnknownInternal.Format(body)
}

func (c *client) PushLogStream(project, pipeline, recordID string, stage api.PipelineStageName, task string, filePath string, close chan struct{}) error {
	path := fmt.Sprintf(apiPathForLogStream, project, pipeline, recordID)
	log.Infof("Path: %s", path)

	host := strings.TrimPrefix(c.baseURL, "http://")
	host = strings.TrimPrefix(host, "https://")
	requestUrl := url.URL{
		Host:     host,
		Path:     cycloneAPIVersion + path,
		RawQuery: fmt.Sprintf("stage=%s", stage),
		Scheme:   "ws",
	}

	if task != "" {
		requestUrl.RawQuery = requestUrl.RawQuery + "&task=" + task
	}

	header := http.Header{
		"Connection":            []string{"Upgrade"},
		"Upgrade":               []string{"websocket"},
		"Host":                  []string{host},
		"Sec-Websocket-Key":     []string{"SGVsbG8sIHdvcmxkIQ=="},
		"Sec-Websocket-Version": []string{"13"},
	}
	filteredHeader := websocketutil.FilterHeader(header)

	ws, _, err := websocket.DefaultDialer.Dial(requestUrl.String(), filteredHeader)
	if err != nil {
		log.Errorf("Fail to new the WebSocket connection as %s", err.Error())
		return err
	}
	defer ws.Close()

	return watchLogs(ws, filePath, close)
}

func watchLogs(ws *websocket.Conn, filePath string, close chan struct{}) error {
	logFile, err := os.Open(filePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer logFile.Close()

	buf := bufio.NewReader(logFile)
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
			fmt.Println("Close the watch of log file")
			return nil
		}
	}
}
