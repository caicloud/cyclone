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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
)

// CreateVersion creates a version and streams response from server to stdout.
func CreateVersion(userID string, version *api.Version, response *api.VersionCreationResponse) error {
	buf, err := json.Marshal(version)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/versions", BaseURL, userID), bytes.NewBuffer(buf))
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

// GetVersion retrieves a version from user ID and service ID.
func GetVersion(userID, versionID string, response *api.VersionGetResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/versions/%s", BaseURL, userID, versionID), nil)
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

// GetVersionLogs gets a version's log.
func GetVersionLogs(userID, versionID string) (int, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/versions/%s/logs", BaseURL, userID, versionID), nil)
	if err != nil {
		return -1, err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	hasLine := false
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			return resp.StatusCode, nil
		} else if err != nil {
			return -1, err
		}
		hasLine = true
		log.Infof("Response frame %v", string(line))
	}

	if !hasLine {
		return -1, errors.New("unexpected empty logs")
	}
	return resp.StatusCode, nil
}

// ListVersions lists all versions of a user.
func ListVersions(userID, serviceID string, response *api.VersionListResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/services/%s/versions", BaseURL, userID, serviceID), nil)
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
