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

package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/cyclone/api"
)

// HTTPHandler identifies the type of a http handler
type HTTPHandler struct {
	BaseURL string
}

// NewHTTPHandler returns a new HTTP request handler.
func NewHTTPHandler(baseURL string) *HTTPHandler {
	return &HTTPHandler{
		BaseURL: baseURL,
	}
}

// GetEvent retrieves a event from event ID.
func (ap *HTTPHandler) GetEvent(eventID string, response *api.GetEventResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/events/%s", ap.BaseURL, eventID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json", eventID)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// SetEvent set a event.
func (ap *HTTPHandler) SetEvent(eventID string, event *api.SetEvent, response *api.SetEventResponse) error {
	buf, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/events/%s", ap.BaseURL, eventID), bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json", eventID)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// Generate the request with the global token and the given contentType.
func generateRequestWithToken(request *http.Request, contentType, token string) error {
	request.Header.Add("content-type", contentType)
	request.Header.Add("token", token)
	return nil
}
