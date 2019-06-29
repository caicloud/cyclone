package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"golang.org/x/oauth2"

	api "github.com/caicloud/cyclone/pkg/server/apis/v1"
)

const (
	//Gitlab
	GitlabURL    = "GITLAB_URL"
	GitlabClient = "GITLAB_CLIENT"
	GitlabSecret = "GITLAB_SECRET"

	CycloneServer      = "CYCLONE_SERVER"
	ConsoleWebEndpoint = "CONSOLE_WEB_ENDPOINT"
)

var gitlabConf = &oauth2.Config{
	ClientID:     os.Getenv(GitlabClient),
	ClientSecret: os.Getenv(GitlabSecret),
	RedirectURL:  os.Getenv(CycloneServer) + "/apis/v1/scms/gitlab/callback",
	Scopes:       []string{},
	Endpoint: oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/oauth/authorize", os.Getenv(GitlabURL)),
		TokenURL: fmt.Sprintf("%s/oauth/token", os.Getenv(GitlabURL)),
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
		return "", fmt.Errorf("gitlab oauth hasn't been implemented")
	default:
		return "", fmt.Errorf("unknow scm type, please choose gitlab or github")
	}
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
		consoleWebServer := os.Getenv(ConsoleWebEndpoint)
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
		return fmt.Errorf("gitlab oauth hasn't been implemented")
	default:
		return fmt.Errorf("unknow scm type, please choose gitlab or github")
	}
}

//TODO: support gitlab API V4 version
func getUserInfo(token *oauth2.Token) (string, string, error) {
	accessToken := token.AccessToken
	gitlabServer := os.Getenv(GitlabURL)
	userV3API := fmt.Sprintf("%s/api/v3/user?access_token=%s", gitlabServer, accessToken)
	if req, err := http.NewRequest(http.MethodGet, userV3API, nil); err != nil {
		log.Error(err)
		return "", "", nil
	} else {
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Error(err)
			log.Error(err)
			return "", "", nil
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return "", "", err
		}
		var userInfo = &api.GitlabUserInfo{}
		err = json.Unmarshal(body, &userInfo)
		if err != nil {
			log.Error(err)
			return "", "", err
		}
		log.Infof("gitlab user %s info: %v\n", userInfo.Username, userInfo)
		userName := userInfo.Username

		return userName, gitlabServer, nil
	}
}
