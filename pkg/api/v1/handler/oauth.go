package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/util/os"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"golang.org/x/oauth2"
)

const ()

var gitlabConf = &oauth2.Config{
	ClientID:     os.GetStringEnv(options.GitlabClient, ""),
	ClientSecret: os.GetStringEnv(options.GitlabSecret, ""),
	RedirectURL:  os.GetStringEnv(options.CycloneServer, "") + "/api/v1/scms/gitlab/callback",
	Scopes:       []string{},
	Endpoint: oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/oauth/authorize", os.GetStringEnv(options.GitlabURL, "")),
		TokenURL: fmt.Sprintf("%s/oauth/token", os.GetStringEnv(options.GitlabURL, "")),
	},
}

var fakeState = "fakestate"

// GetAuthCodeURL get the redirect url from config
// then return url to frontend
func GetAuthCodeURL(ctx context.Context, scm string) (string, error) {
	switch scm {
	case "gitlab":
		url := gitlabConf.AuthCodeURL(fakeState)
		log.Infof("new redirect url for gitlab oauth2: %s", url)
		return url, nil
	case "github":
		//TODO: support github oauth2
		return "", nil
	}
	return "", nil
}

// GetToken use code to change token and redirect to frontend
func GetToken(ctx context.Context, scm string, code string, state string) error {
	if code == "" || state != fakeState {
		return fmt.Errorf("code is nil or state is not right: code %s, state %s", code, state)
	}
	switch scm {
	case "gitlab":
		oauthToken, err := gitlabConf.Exchange(context.TODO(), code)
		if err != nil {
			return err
		}
		if !oauthToken.Valid() {
			return fmt.Errorf("token invalid. token: %v", oauthToken)
		}
		userName, server, err := getUserInfo(oauthToken)
		if err != nil {
			return err
		}
		consoleWebServer := os.GetStringEnv(options.ConsoleWebEndpoint, "")
		redirectURL := fmt.Sprintf(
			"%s/devops/workspace/add?type=gitlab&code=%s&state=%s&token=%s&username=%s&server=%s",
			consoleWebServer, code, state, oauthToken.AccessToken, userName, server,
		)

		httpCtx := service.HTTPContextFrom(ctx)
		w := httpCtx.ResponseWriter()
		r := httpCtx.Request()
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)

		return nil
	case "github":
		//TODO: support github oauth2
		return nil
	}
	return nil
}

//TODO: support gitlab API V4 version
func getUserInfo(token *oauth2.Token) (string, string, error) {
	accessToken := token.AccessToken
	gitlabServer := os.GetStringEnv(options.GitlabURL, "")
	userV3API := fmt.Sprintf("%s/api/v3/user?access_token=%s", gitlabServer, accessToken)
	if req, err := http.NewRequest(http.MethodGet, userV3API, nil); err != nil {
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
		log.Infof("gitlab user %s info: %v\n", userInfo.Username, userInfo)
		userName := userInfo.Username

		return userName, gitlabServer, nil
	}
}
