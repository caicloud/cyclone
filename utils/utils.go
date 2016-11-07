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

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/caicloud/cyclone/pkg/log"
)

const (
	// DeployUser is the username for e2e-test.
	DeployUser = "deploy"
	// DeployUID is the UID for e2e-test.
	DeployUID = "deployUID"
)

var (
	codeDeployReady = 1
)

// InvokeUpdateImageAPI invokes update image API.
func InvokeUpdateImageAPI(userID, applicationName, clusterName, partitionName, containerName, imageName, endpoint string) error {
	// In e2e-test, we dont really send a http request. The call will be always successful.
	if userID == DeployUID {
		return nil
	}
	// Set up form data.
	values := make(url.Values)
	values.Set("uid", userID)
	values.Set("cid", clusterName)
	values.Set("partitionName", partitionName)
	values.Set("applicationName", applicationName)
	values.Set("containerName", containerName)
	values.Set("imageName", imageName)

	// Build a client.
	client := &http.Client{}
	// Submit form.
	resp, err := client.PostForm(endpoint, values)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.DebugWithFields("invokeUpdateImageAPI response",
		log.Fields{
			"status code": resp.StatusCode,
			"body":        string(body[:]),
		})
	// Check response.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %v, response body: %v", resp.StatusCode, string(body[:]))
	}

	return nil
}

// InvokeCheckDeployStateAPI invokes Check Deploy State API.
func InvokeCheckDeployStateAPI(jsonStr []byte, endpoint string) (bool, error) {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json;charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	log.InfoWithFields("InvokeCheckDeployStateAPI response",
		log.Fields{
			"status code": resp.StatusCode,
			"body":        string(body[:]),
		})

	// Check status code.
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("status code: %v, response body: %v", resp.StatusCode, string(body[:]))
	}

	result := make(map[string]int)
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	code, ok := result["code"]
	if !ok || code != codeDeployReady {
		return false, nil
	}

	return true, nil

}
