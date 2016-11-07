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
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
)

const (
	// Defines the usernames and pwds in docker registry server.
	// - AdminUser is used to run normal test cases, it has all access rights
	//   to docker repositories.
	// - ListUser is used solely for listing services.
	// - Others are normal users, we use them to test different cases, e.g. build
	//   from user1 should fail to push to user2.
	AdminUser     = "admin"
	AdminUID      = "adminUID"
	AdminPassword = "admin_password"

	ListUser     = "list"
	ListUID      = "listUID"
	ListPassword = "list_password"

	AliceUser     = "alice"
	AliceUID      = "aliceUID"
	AlicePassword = "alice_password"

	BobUser     = "bob"
	BobUID      = "bobUID"
	BobPassword = "bob_password"

	// Default registry address and docker host in e2e test.
	DefaultRegistryAddress = "localhost:5000"
	DefaultDockerHost      = "unix:///var/run/docker.sock"

	// Define the address to access caicloud auth. Right now, we've disabled
	// caicloud auth in e2e test.
	DefaultAuthAddress = "https://auth-canary.caicloud.io"

	// Local cyclone access information.
	cyclonePort     = "7099"
	apiVersion      = "v0.1"
	authAPITemplate = "%s/api/v0.1/users/%s/authenticate"
)

var (
	// BaseURL defines the url to cyclone.
	BaseURL = fmt.Sprintf("http://localhost:%s/api/%s", cyclonePort, apiVersion)
)

// TokenRequest is the request to get the token.
type TokenRequest struct {
	Password string `json:"password,omitempty"`
}

// TokenResponse is the response to get the token.
type TokenResponse struct {
	StatusMessage string  `json:"statusMessage,omitempty"`
	Token         string  `json:"token,omitempty"`
	UID           string  `json:"uid,omitempty"`
	Error         string  `json:"error,omitempty"`
	Profile       Profile `json:"profile,omitempty"`
}

// Profile is the profile object of a user.
type Profile struct {
	Email     string `json:"email,omitempty"`
	Cellphone int    `json:"cellphone,omitempty"`
}

// GetToken returns the token from ${DefaultAuthAddress}/api/v0.1/users/{username}/authenticate
func GetToken(username string, request TokenRequest, response *TokenResponse) error {
	buf, err := json.Marshal(request)
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf(authAPITemplate, DefaultAuthAddress, username), "application/json", bytes.NewBuffer(buf))
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
	if response.Error != "" {
		return errors.New(response.Error)
	}
	return nil
}

// SetupTokenUID will get the token and uid from auth server. This won't be called
// if authserver is disabled.
func SetupTokenUID(username, password string) (string, string) {
	// Get the token and UID from auth server.
	tokenRequest := TokenRequest{
		Password: password,
	}
	tokenResponse := &TokenResponse{}
	if err := GetToken(username, tokenRequest, tokenResponse); err != nil {
		log.Fatalf("Unable to get the token for %s: %s", username, err)
	}
	Token := tokenResponse.Token
	UID := tokenResponse.UID
	return Token, UID
}

// WaitComponents waits cyclone instance to be up.
func WaitComponents() {
	cycloneOK := false
	log.Info("Waiting for cyclone to start, please wait. If things went wrong, this may loop forever")
	for !cycloneOK {
		_, err := http.Get(fmt.Sprintf("http://localhost:%s", cyclonePort))
		if err == nil {
			cycloneOK = true
		}
		time.Sleep(time.Second)
	}
	log.Info("Cyclone started")
}

// RegisterResource register resources to mongo.
func RegisterResource() error {
	data := api.WorkerNode{
		Name:        "test",
		Description: "test",
		IP:          "127.0.0.1",
		DockerHost:  osutil.GetStringEnv("DOCKER_HOST", DefaultDockerHost),
		Type:        "system",
		TotalResource: api.NodeResource{
			Memory: 1024 * 1024 * 1024,
			CPU:    1024,
		},
	}
	buf, err := json.Marshal(&data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/system_worker_nodes", BaseURL), bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	client := &http.Client{}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("%v", resp))
	}

	log.Info("Register resource to mongo.")
	return nil
}

// IsAvailable returns whether the cyclone is running.
func IsAvailable() bool {
	_, err := http.Get(BaseURL)
	if err == nil {
		return true
	}
	return false
}
