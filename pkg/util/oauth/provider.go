package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/nirvana/log"
	"github.com/google/go-github/github"
	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	gitlabv4 "gopkg.in/xanzy/go-gitlab.v0"

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/pkg/util/os"
)

const (
	// GITLAB represents scm type as api path param
	GITLAB = "gitlab"
	// GITHUB represents scm type as api path param
	GITHUB = "github"

	// FakeState represents state of oauth params
	FakeState = "fakestate"

	// GithubURL github url
	GithubURL = "https://github.com"

	apiPathForGitlabVersion = "%s/api/v4/version"

	v4APIVersion = "v4"

	v3APIVersion = "v3"
)

// CallbackPath oauth server callback api path
var CallbackPath = os.GetStringEnv(options.CycloneServer, "") + "/api/v1/scms/%s/callback"

// GitlabURL gitlab url
var GitlabURL = os.GetStringEnv(options.GitlabURL, "")

// GitlabConfig gitlab oauth2 config
var GitlabConfig = oauth2.Config{
	ClientID:     os.GetStringEnv(options.GitlabClient, ""),
	ClientSecret: os.GetStringEnv(options.GitlabSecret, ""),
	RedirectURL:  fmt.Sprintf(CallbackPath, "gitlab"),
	Scopes:       []string{"api"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  fmt.Sprintf("%s/oauth/authorize", os.GetStringEnv(options.GitlabURL, "")),
		TokenURL: fmt.Sprintf("%s/oauth/token", os.GetStringEnv(options.GitlabURL, "")),
	},
}

// GithubConfig github oauth2 config
var GithubConfig = oauth2.Config{
	ClientID:     os.GetStringEnv(options.GithubClient, ""),
	ClientSecret: os.GetStringEnv(options.GithubSecret, ""),
	RedirectURL:  fmt.Sprintf(CallbackPath, "github"),
	Scopes:       []string{"repo"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
}

// versionResponse represents the response of Gitlab version API.
type versionResponse struct {
	Version   string `json:"version"`
	Reversion string `json:"reversion"`
}

// Client use for gitlab and github
type Client struct {
	config oauth2.Config
	scm    string
	server string
}

// NewOauthClient new oauth client
func NewOauthClient(scm string) (OauthProvider, error) {
	switch scm {
	case GITLAB:
		return &Client{GitlabConfig, GITLAB, GitlabURL}, nil
	case GITHUB:
		return &Client{GithubConfig, GITHUB, GithubURL}, nil
	default:
		return nil, errors.ErrorUnsupported.Error("scm type", scm)
	}

}

// OauthProvider provider apis for oauth
type OauthProvider interface {
	GetAuthCodeURL(state string) (url string)
	GetToken(state string, code string) (token string, err error)
	DetectAPIVersion(token string) (v string, err error)
	GetUserInfo(token string) (username string, server string, err error)
}

// GetAuthCodeURL accept scm type and return authcodeurl to frontend
func (oc *Client) GetAuthCodeURL(state string) string {
	return oc.config.AuthCodeURL(state)
}

// GetToken handle oauth server callback and return oauth token
func (oc *Client) GetToken(state string, code string) (string, error) {
	if code == "" || state != FakeState {
		return "", errors.ErrorValidationFailed.Error("state or code", "invalid value")
	}
	token, err := oc.config.Exchange(context.TODO(), code)
	if err != nil {
		return "", errors.ErrorUnknownInternal.Error(err)
	}
	if !token.Valid() {
		return "", errors.ErrorUnknownInternal.Error(err)
	}
	return token.AccessToken, nil
}

// GetUserInfo get the userinfo by github or gitlab client
func (oc *Client) GetUserInfo(token string) (string, string, error) {
	switch oc.scm {
	case GITLAB:
		version, err := oc.DetectAPIVersion(token)
		if err != nil {
			return "", "", err
		}

		if version == v4APIVersion {
			client := gitlabv4.NewOAuthClient(nil, token)
			client.SetBaseURL(GitlabURL + "/api/v4")
			user, _, err := client.Users.CurrentUser()
			if err != nil {
				return "", "", errors.ErrorUnknownInternal.Error(err)
			}
			return user.Username, GitlabURL, nil

		} else {
			// version == v3APIVersion
			client := gitlab.NewOAuthClient(nil, token)
			client.SetBaseURL(GitlabURL + "/api/v3")
			user, _, err := client.Users.CurrentUser()
			if err != nil {
				return "", "", errors.ErrorUnknownInternal.Error(err)
			}
			return user.Username, GitlabURL, nil
		}

	case GITHUB:
		client := github.NewClient(
			oauth2.NewClient(
				context.TODO(),
				oauth2.StaticTokenSource(
					&oauth2.Token{AccessToken: token},
				),
			),
		)
		user, _, err := client.Users.Get("")
		if err != nil {
			return "", "", errors.ErrorUnknownInternal.Error(err)
		}
		return *user.Login, GithubURL, nil
	default:
		return "", "", errors.ErrorUnsupported.Error("scm type", oc.scm)
	}
}

// DetectAPIVersion to deect api version of gitlab v3 or v4
func (oc *Client) DetectAPIVersion(token string) (string, error) {
	url := fmt.Sprintf(apiPathForGitlabVersion, oc.server)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error(err)
		return "", errors.ErrorUnknownInternal.Error(err)
	}

	// Set headers.
	req.Header.Set("Content-Type", "application/json")

	// Use Oauth token
	req.Header.Set("Authorization", "Bearer "+token)

	// Use client with redirect disabled, then status code will be 302
	// if Gitlab server does not support /api/v4/version request.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return "", errors.ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", errors.ErrorUnknownInternal.Error(err)
		}

		gv := &versionResponse{}
		err = json.Unmarshal(body, gv)
		if err != nil {
			log.Error(err)
			return "", errors.ErrorUnknownInternal.Error(err)
		}

		log.Infof("Gitlab version is %s, will use %s API", gv.Version, v4APIVersion)
		return v4APIVersion, nil
	case http.StatusNotFound, http.StatusFound:
		return v3APIVersion, nil
	default:
		log.Warningf("Status code of Gitlab API version request is %d, use v3 in default", resp.StatusCode)
		return v3APIVersion, nil
	}
}
