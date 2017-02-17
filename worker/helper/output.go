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

package helper

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/osutil"
	step_log "github.com/caicloud/cyclone/worker/log"
)

const (
	// SERVER_HOST is a Env variable name
	SERVER_HOST           = "SERVER_HOST"
	DOCKER_IMAGE_LOG_FLAG = "layer"
)

var (
	pushLogAPI string
	// ErrNoOutput is the error for no output file.
	ErrNoOutput = errors.New("event has no output file")
)

// PushLogToCyclone would push the log to cyclone.
func PushLogToCyclone(event *api.Event) error {
	versionLogRaw, err := getLogFromOutputFile()
	if err != nil {
		// No output file, just return directly.
		if err == ErrNoOutput {
			return nil
		}
		return err
	}
	versionLog := ""
	array := strings.Split(versionLogRaw, "\n")
	for _, arr := range array {
		if arr != "\r" && arr != "" {
			if isDockerImageOperationLog(arr) != true {
				versionLog += arr
				versionLog += "\n"
			}
		}
	}
	response := &api.VersionLogCreateResponse{}
	logCreateRequest := api.VersionLog{
		Logs:      versionLog,
		VerisonID: event.Version.VersionID,
	}
	buf, err := json.Marshal(logCreateRequest)
	if err != nil {
		return err
	}
	cycloneServer := osutil.GetStringEnv(SERVER_HOST, "")
	if cycloneServer == "" {
		return errors.New("No cyclone spicified.")
	}
	pushLogAPI = fmt.Sprintf("%s/api/%s/%s/versions/%s/logs", cycloneServer, api.APIVersion,
		event.Service.UserID, event.Version.VersionID)
	req, err := http.NewRequest("POST", pushLogAPI, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")
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

// getLogFromOutputFile returns log in string format.
func getLogFromOutputFile() (string, error) {
	if step_log.Output == nil {
		return "", ErrNoOutput
	}
	logFile, err := os.Open(step_log.Output.Name())
	if err != nil {
		return "", err
	}
	buf := bufio.NewReader(logFile)
	content, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// isDockerImageOperationLog check the log whether is the log of pulling or pushing docker image.
func isDockerImageOperationLog(log string) bool {
	return strings.HasPrefix(log, DOCKER_IMAGE_LOG_FLAG)
}
