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

package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/cyclone/api"
)

// CreateService creates a service with given configurations.
func CreateService(userID string, service *api.Service, response *api.ServiceCreationResponse) error {
	buf, err := json.Marshal(service)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/services", BaseURL, userID), bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
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

// GetService retrieves a service from user ID and service ID.
func GetService(userID, serviceID string, response *api.ServiceGetResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/services/%s", BaseURL, userID, serviceID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
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

// ListServices lists all services of a user.
func ListServices(userID string, response *api.ServiceListResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/services", BaseURL, userID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
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

// DeleteService deletes a service from service ID and user ID.
func DeleteService(userID, serviceID string, response *api.ServiceDelResponse) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s/services/%s", BaseURL, userID, serviceID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
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
