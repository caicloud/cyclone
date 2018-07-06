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

package provider

import (
	"strings"
	"golang.org/x/oauth2"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/osutil"
	"fmt"
	"github.com/caicloud/cyclone/cmd/worker/options"
)

// parseURL is a helper func to parse the url,such as https://github.com/caicloud/test.git
// to return owner(caicloud) and name(test).
func parseURL(url string) (string, string) {
	strs := strings.SplitN(url, "/", -1)
	name := strings.SplitN(strs[4], ".", -1)
	return strs[3], name[0]
}

// getAuthCodeURL gets the URL for token request.
func getAuthCodeURL(state string, scmType string) (string, error) {
	conf, err := getConfig(scmType)
	if err != nil {
		return "", err
	}

	return conf.AuthCodeURL(state), nil
}

// getConfig pack the information into oauth.config that is used to get token
// ClientID、ClientSecret, these values are used to assemble the token request url
// and they come from github or gitlab or others by registering application information.
func getConfig(scmType string) (*oauth2.Config, error) {
	var clientID, clientSecret, redirectURL, authURL, tokenURL string
	var scopes []string

	switch scmType {

		case api.GITLAB:
			cyclonePath := osutil.GetStringEnv(options.CycloneServer, "")
			gitlabServer := osutil.GetStringEnv(options.GitlabURL, "")
			clientID = osutil.GetStringEnv(options.GitlabClient, "")
			clientSecret = osutil.GetStringEnv(options.GitlabSecret, "")
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