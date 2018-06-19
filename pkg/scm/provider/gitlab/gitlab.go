/*
Copyright 2017 caicloud authors. All rights reserved.

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

package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/golang/glog"
	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	gitlabv4 "gopkg.in/xanzy/go-gitlab.v0"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
	"github.com/caicloud/cyclone/pkg/scm/provider"
)

const (
	apiPathForGitlabVersion = "%s/api/v4/version"

	// gitLabServer represents the server address for public Gitlab.
	gitLabServer = "https://gitlab.com"

	v3APIVersion = "v3"

	v4APIVersion = "v4"
)

var gitlabServerAPIVersions = make(map[string]string)

func init() {
	if err := scm.RegisterProvider(api.Gitlab, NewGitlab); err != nil {
		log.Errorln(err)
	}
}

func NewGitlab(scmCfg *api.SCMConfig) (scm.SCMProvider, error) {
	version, err := getGitlabAPIVersion(scmCfg)
	if err != nil {
		log.Errorf("Fail to get API version for server %s as %v", scmCfg.Server, err)
		return nil, err
	}
	log.Infof("Gitlab version is %s", version)

	switch version {
	case v3APIVersion:
		client, err := newGitlabClient(scmCfg.Server, scmCfg.Username, scmCfg.Token)
		if err != nil {
			log.Error("fail to new Gitlab v3 client as %v", err)
			return nil, err
		}

		return &GitlabV3{scmCfg, client}, nil
	case v4APIVersion:
		v4Client, err := newGitlabV4Client(scmCfg.Server, scmCfg.Username, scmCfg.Token)
		if err != nil {
			log.Error("fail to new Gitlab v4 client as %v", err)
			return nil, err
		}

		return &GitlabV4{scmCfg, v4Client}, nil
	default:
		err = fmt.Errorf("Gitlab API version %s is not supported, only support %s and %s", version, v3APIVersion, v4APIVersion)
		log.Errorln(err)
		return nil, err
	}
}

// newGitlabV4Client news Gitlab v4 client by token. If username is empty, use private-token instead of oauth2.0 token.
func newGitlabV4Client(server, username, token string) (*gitlabv4.Client, error) {
	var client *gitlabv4.Client

	if len(username) == 0 {
		client = gitlabv4.NewClient(nil, token)
	} else {
		client = gitlabv4.NewOAuthClient(nil, token)
	}

	if err := client.SetBaseURL(server + "/api/" + v4APIVersion); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return client, nil
}

// newGitlabClient news Gitlab v3 client by token. If username is empty, use private-token instead of oauth2.0 token.
func newGitlabClient(server, username, token string) (*gitlab.Client, error) {
	var client *gitlab.Client

	if len(username) == 0 {
		client = gitlab.NewClient(nil, token)
	} else {
		client = gitlab.NewOAuthClient(nil, token)
	}

	if err := client.SetBaseURL(server + "/api/" + v3APIVersion); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return client, nil
}

func getGitlabAPIVersion(scmCfg *api.SCMConfig) (string, error) {
	// Directly get API version if it has been recorded.
	server := provider.ParseServerURL(scmCfg.Server)
	if v, ok := gitlabServerAPIVersions[server]; ok {
		return v, nil
	}

	// Dynamically detect API version if it has not been recorded, and record it for later use.
	version, err := detectGitlabAPIVersion(scmCfg)
	if err != nil {
		return "", err
	}

	gitlabServerAPIVersions[server] = version

	return version, nil
}

// gitlabVersionResponse represents the response of Gitlab version API.
type gitlabVersionResponse struct {
	Version   string `json:"version"`
	Reversion string `json:"reversion"`
}

func detectGitlabAPIVersion(scmCfg *api.SCMConfig) (string, error) {
	if scmCfg.Token == "" {
		token, err := getGitlabOauthToken(scmCfg)
		if err != nil {
			log.Error(err)
			return "", err
		}
		scmCfg.Token = token
	}

	url := fmt.Sprintf(apiPathForGitlabVersion, scmCfg.Server)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error(err)
		return "", err
	}

	// Set headers.
	req.Header.Set("Content-Type", "application/json")
	if scmCfg.Username == "" {
		// Use private token when username is empty.
		req.Header.Set("PRIVATE-TOKEN", scmCfg.Token)
	} else {
		// Use Oauth token when username is not empty.
		req.Header.Set("Authorization", "Bearer "+scmCfg.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(err)
		return "", err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// body, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	return "", err
		// }
		// defer resp.Body.Close()

		// gv := &GitlabVersionResponse{}
		// err = json.Unmarshal(body, gv)
		// if err != nil {
		// 	log.Error(err)
		// 	return "", err
		// }

		// TODO (robin) Remove this workround, and judge version by status code.
		log.Infof("Header of response: %v", resp.Header)
		if resp.Header.Get("Content-Type") != "application/json" {
			return v3APIVersion, nil
		}

		return v4APIVersion, nil
	case http.StatusNotFound:
		return v3APIVersion, nil
	default:
		log.Warningf("Status code of Gitlab API version request is %d, use v3 in default", resp.StatusCode)
		return v3APIVersion, nil
	}
}

func getGitlabOauthToken(scm *api.SCMConfig) (string, error) {
	if len(scm.Username) == 0 || len(scm.Password) == 0 {
		return "", fmt.Errorf("GitHub username or password is missing")
	}

	bodyData := struct {
		GrantType string `json:"grant_type"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}{
		GrantType: "password",
		Username:  scm.Username,
		Password:  scm.Password,
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return "", fmt.Errorf("fail to new request body for token as %s", err.Error())
	}

	// If use the public Gitlab, must use the HTTPS protocol.
	if strings.Contains(scm.Server, "gitlab.com") && strings.HasPrefix(scm.Server, "http://") {
		log.Infof("Convert SCM server from %s to %s to use HTTPS protocol for public Gitlab", scm.Server, gitLabServer)
		scm.Server = gitLabServer
	}

	tokenURL := fmt.Sprintf("%s%s", scm.Server, "/oauth/token")
	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Errorf("Fail to new the request for token as %s", err.Error())
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Fail to request for token as %s", err.Error())
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to request for token as %s", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		var token oauth2.Token
		err := json.Unmarshal(body, &token)
		if err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}

	err = fmt.Errorf("Fail to request for token as %s", body)
	return "", err
}
