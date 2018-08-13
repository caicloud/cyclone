/*
Copyright 2018 caicloud authors. All rights reserved.

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
	"fmt"

	log "github.com/golang/glog"

	"encoding/json"
	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/util/os"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
)

// getAuthCodeURL gets the URL for token request.
func getAuthCodeURL(state string, scmType api.SCMType) (string, error) {
	conf, err := getConfig(scmType)
	if err != nil {
		return "", err
	}
	return conf.AuthCodeURL(state), nil
}

// getConfig pack the information into oauth.config that is used to get token
// ClientID、ClientSecret, these values are used to assemble the token request url
// and they come from github or gitlab or others by registering application information.
func getConfig(scmType api.SCMType) (*oauth2.Config, error) {
	var clientID, clientSecret, redirectURL, authURL, tokenURL string
	var scopes []string
	switch scmType {
	case api.Gitlab:
		cyclonePath := os.GetStringEnv(options.CycloneServer, "")
		gitlabServer := os.GetStringEnv(options.GitlabURL, "")
		clientID = os.GetStringEnv(options.GitlabClient, "")
		clientSecret = os.GetStringEnv(options.GitlabSecret, "")
		redirectURL = fmt.Sprintf("%s/%s/scm/%s/callback", cyclonePath, "api/v1", api.GITLAB)
		scopes = []string{"api"}
		authURL = fmt.Sprintf("%s/oauth/authorize", gitlabServer)
		tokenURL = fmt.Sprintf("%s/oauth/token", gitlabServer)
	default:
		return nil, fmt.Errorf("unknown scm type %s", scmType)
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}, nil
}

//getToken get token with gitlab code  by the way gitlab oauth
func getToken(code, state string) (*oauth2.Token, error) {
	config, err := getConfig(api.Gitlab)
	if err != nil {
		return nil, err
	}
	var token *oauth2.Token
	token, err = config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}
	if !token.Valid() {
		return nil, fmt.Errorf("token invalid. Got: %v", token)
	}
	return token, nil
}

func getUserInfo(token *oauth2.Token) (string, string, error) {
	accessToken := token.AccessToken
	gitlabServer := os.GetStringEnv(options.GitlabURL, "http://192.168.21.100:10080")
	userApi := fmt.Sprintf("%s/api/v3/user?access_token=%s", gitlabServer, accessToken)
	if req, err := http.NewRequest(http.MethodGet, userApi, nil); err != nil {
		log.Error(err.Error())
		return "", "", nil
	} else {
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Error(err.Error())
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
		}
		var userInfo = &api.GitlabUserInfo{}
		json.Unmarshal(body, &userInfo)
		userName := userInfo.Username
		fmt.Println(userInfo)
		server := gitlabServer
		return userName, server, nil
	}
}
