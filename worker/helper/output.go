// Copyright 2015 caicloud authors. All rights reserved.

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

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/osutil"
)

const (
	// SERVER_HOST is a Env variable name
	SERVER_HOST = "SERVER_HOST"
)

var (
	pushLogAPI string
	// ErrNoOutput is the error for no output file.
	ErrNoOutput = errors.New("event has no output file")
)

// PushLogToCyclone would push the log to cyclone.
func PushLogToCyclone(event *api.Event) error {
	versionLog, err := getLogFromOutputFile(event)
	if err != nil {
		// No output file, just return directly.
		if err == ErrNoOutput {
			return nil
		}
		return err
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
func getLogFromOutputFile(event *api.Event) (string, error) {
	if event.Output == nil {
		return "", ErrNoOutput
	}
	logFile, err := os.Open(event.Output.Name())
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
