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

package provider

import (
	"fmt"
	"strings"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/store"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const CYCLONE_SERVER_HOST = "CYCLONE_SERVER_HOST"

// GitHub is the type for Github remote provider.
type GitHub struct {
}

// NewGitHub returns a new GitHub remoter.
func NewGitHub() *GitHub {
	return &GitHub{}
}

// Pack the information into oauth.config that is used to get token
// ClientID、ClientSecret,these values use to assemble the token request url and
// there values come from github or other by registering some information.
func (g *GitHub) getConf() (*oauth2.Config, error) {
	//cyclonePath http request listen address
	cyclonePath := osutil.GetStringEnv(CYCLONE_SERVER_HOST, "http://localhost:7099")
	clientID := osutil.GetStringEnv("CLIENTID", "")
	clientSecret := osutil.GetStringEnv("CLIENTIDSECRET", "")
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("%s/api/%s/remotes/%s/authcallback", cyclonePath, api.APIVersion, "github"),
		Scopes:       []string{"repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}, nil
}

// GetTokenQuestURL gets the URL for token request.
func (g *GitHub) GetTokenQuestURL(userID string) (string, error) {
	// Get a object to request token.
	conf, err := g.getConf()
	if err != nil {
		return "", err
	}
	// Get the request url and send to the github or other.
	url := conf.AuthCodeURL(userID) // Use userid as state.
	log.InfoWithFields("cyclone receives creating token request",
		log.Fields{"request url": url})

	if !strings.Contains(url, "github") {
		log.ErrorWithFields("Unable to get the  url", log.Fields{"user_id": userID})
		return "", fmt.Errorf("Unable to get the  url")
	}
	return url, nil
}

// Authcallback is the callback handler.
func (g *GitHub) Authcallback(code, state string) (string, error) {
	if code == "" || state == "" {
		return "", fmt.Errorf("code or state is nil")
	}

	// Caicloud web address,eg caicloud.io
	uiPath := osutil.GetStringEnv("CONSOLE_WEB_ENDPOINT", "http://localhost:8000")
	redirectURL := fmt.Sprintf("%s/cyclone/add?type=github&code=%s&state=%s", uiPath, code, state)

	// Sync to get token.
	go g.getToken(code, state)
	return redirectURL, nil
}

// get token by using code from github.
func (g *GitHub) getToken(code, state string) error {
	if code == "" || state == "" {
		log.ErrorWithFields("code or state is nil", log.Fields{"code": code, "state": state})
		return fmt.Errorf("code or state is nil")
	}
	log.InfoWithFields("cyclone receives auth code", log.Fields{"request code": code})

	// Get a object to request token.
	conf, err := g.getConf()
	if err != nil {
		log.Warnf("Unable to get the conf according coderepository")
		return err
	}

	// To communication with githubo or other vcs to get token.
	var tok *oauth2.Token
	tok, err = conf.Exchange(oauth2.NoContext, code) // Post a token request and receive toeken.
	if err != nil {
		log.Error(err)
		return err
	}

	if !tok.Valid() {
		log.Fatalf("Token invalid. Got: %#v", tok)
		return err
	}
	log.Info("get the token successfully!")

	// Create service in database (but not ready to be used yet).
	vcstoken := api.VscToken{
		UserID:   state,
		Vsc:      "github",
		Vsctoken: *tok,
	}

	ds := store.NewStore()
	defer ds.Close()

	_, err = ds.FindtokenByUserID(state, "github")
	if err != nil {
		err = ds.NewTokenDocument(&vcstoken)
		if err != nil {
			log.ErrorWithFields("NewTokenDocument", log.Fields{"user_id": state,
				"token": tok, "error": err})
			return err
		}
	} else {
		err = ds.UpdateToken(&vcstoken)
		if err != nil {
			log.ErrorWithFields("UpdateToken", log.Fields{"user_id": state,
				"token": tok, "error": err})
			return err
		}
	}

	return nil
}

// GetRepos gets token by using code from github.
func (g *GitHub) GetRepos(userID string) (Repos []api.Repo, username, avatarURL string, err error) {
	ds := store.NewStore()
	defer ds.Close()

	tok, err := ds.FindtokenByUserID(userID, "github")
	if err != nil {
		log.ErrorWithFields("find token failed", log.Fields{"user_id": userID, "error": err})
		return Repos, username, avatarURL, err
	}

	// Use token to get repo list.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tok.Vsctoken.AccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)

	// List all repositories for the authenticated user.
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 30},
	}
	// Get all pages of results.
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.List("", opt)
		if err != nil {
			message := "Unable to list repo by token"
			log.ErrorWithFields(message, log.Fields{"user_id": userID, "token": tok, "error": err})
			return Repos, username, avatarURL, fmt.Errorf("Unable to list repo by token")
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	Repos = make([]api.Repo, len(allRepos))
	for i, repo := range allRepos {
		Repos[i].Name = *repo.Name
		Repos[i].URL = *repo.CloneURL
		Repos[i].Owner = *repo.Owner.Login
	}

	user, _, err := client.Users.Get("")
	if err != nil {
		log.ErrorWithFields("Users.Get returned error", log.Fields{"user_id": userID,
			"token": tok, "error": err})
		return Repos, username, avatarURL, err
	}
	username = *user.Login
	avatarURL = *user.AvatarURL

	return Repos, username, avatarURL, nil
}

// LogOut logs out.
func (g *GitHub) LogOut(userID string) error {
	// Fiind the token by userid and code repository.
	ds := store.NewStore()
	defer ds.Close()
	tok, err := ds.FindtokenByUserID(userID, "github")
	if err != nil {
		log.ErrorWithFields("find token failed", log.Fields{"user_id": userID, "error": err})
		return err
	}

	conf, err := g.getConf()
	if err != nil {
		log.ErrorWithFields("Unable to get the conf according coderepository",
			log.Fields{"user_id": userID, "error": err})
		return err
	}

	tp := github.BasicAuthTransport{
		Username: conf.ClientID,
		Password: conf.ClientSecret,
	}

	client := github.NewClient(tp.Client())
	_, err = client.Authorizations.Revoke(conf.ClientID, tok.Vsctoken.AccessToken)
	if err != nil {
		log.ErrorWithFields("revoke failed", log.Fields{"user_id": userID, "error": err})
		return err
	}

	// Remove the token saved in DB.
	err = ds.RemoveTokeninDB(userID, "github")
	if err != nil {
		log.ErrorWithFields("remove token failed", log.Fields{"user_id": userID, "error": err})
		return err
	}

	return nil
}

// CreateHook is a helper to register webhook.
func (g *GitHub) CreateHook(service *api.Service) error {
	webhooktype := service.Repository.Webhook
	if webhooktype == "" {
		return fmt.Errorf("no need webhook registry")
	}

	if webhooktype == api.GITHUB {
		url := getHookURL(webhooktype, service.ServiceID)
		if url == "" {
			log.Infof("url is empty", log.Fields{"user_id": service.UserID})
			return nil
		}

		ds := store.NewStore()
		defer ds.Close()

		tok, err := ds.FindtokenByUserID(service.UserID, api.GITHUB)
		if err != nil {
			log.ErrorWithFields("find token failed", log.Fields{"user_id": service.UserID, "error": err})
			return err
		}

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: tok.Vsctoken.AccessToken},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client := github.NewClient(tc)

		var hook github.Hook
		hook.Name = github.String("web")
		hook.Events = []string{"push", "pull_request"}
		hook.Config = map[string]interface{}{}
		hook.Config["url"] = url
		hook.Config["content_type"] = "json"
		onwer, name := parseURL(service.Repository.URL)
		_, _, err = client.Repositories.CreateHook(onwer, name, &hook)
		return err
	}
	log.WarnWithFields("not support vcs repository", log.Fields{"vcs repository": webhooktype})
	return fmt.Errorf("not support vcs repository in create webhook")
}

// DeleteHook is a helper to unregister webhook.
func (g *GitHub) DeleteHook(service *api.Service) error {
	webhooktype := service.Repository.Webhook
	if webhooktype == "" {
		return fmt.Errorf("no need webhook registry")
	}

	if webhooktype == api.GITHUB {
		url := getHookURL(webhooktype, service.ServiceID)
		if url == "" {
			log.Infof("url is empty", log.Fields{"user_id": service.UserID})
			return nil
		}

		ds := store.NewStore()
		defer ds.Close()

		tok, err := ds.FindtokenByUserID(service.UserID, api.GITHUB)
		if err != nil {
			log.ErrorWithFields("find token failed", log.Fields{"user_id": service.UserID, "error": err})
			return err
		}

		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: tok.Vsctoken.AccessToken},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client := github.NewClient(tc)

		owner, name := parseURL(service.Repository.URL)
		hooks, _, err := client.Repositories.ListHooks(owner, name, nil)
		if err != nil {
			return err
		}

		var hook *github.Hook
		hasFoundHook := false
		for _, hook = range hooks {
			hookurl, ok := hook.Config["url"].(string)
			if !ok {
				continue
			}
			if strings.HasPrefix(hookurl, url) {
				hasFoundHook = true
				break
			}
		}
		if hasFoundHook {
			_, err = client.Repositories.DeleteHook(owner, name, *hook.ID)
			return err
		}
		return nil
	}
	log.WarnWithFields("not support vcs repository", log.Fields{"vcs repository": webhooktype})
	return fmt.Errorf("not support vcs repository in delete webhook")
}

// PostCommitStatus posts Commit Status To Github.
func (g *GitHub) PostCommitStatus(service *api.Service, version *api.Version) error {
	// Check if github webhook has set.
	if service.Repository.Webhook != api.GITHUB {
		return fmt.Errorf("vcs github webhook hasn't set")
	}

	// Check if has set commitID.
	if version.Commit == "" {
		return fmt.Errorf("commit hasn't set")
	}

	// Get token.
	ds := store.NewStore()
	defer ds.Close()
	tok, err := ds.FindtokenByUserID(service.UserID, api.GITHUB)
	if err != nil {
		log.ErrorWithFields("find token failed", log.Fields{"user_id": service.UserID, "error": err})
		return err
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tok.Vsctoken.AccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	// Post commit status.
	owner, repo := parseURL(service.Repository.URL)
	urlHost := osutil.GetStringEnv(CYCLONE_SERVER_HOST, "https://fornax-canary.caicloud.io")

	var state string
	if version.Status == api.VersionHealthy {
		state = api.CISuccess
	} else if version.Status == api.VersionFailed || version.Status == api.VersionCancel {
		state = api.CIFailure
	} else {
		state = api.CIPending
	}

	log.Infof("Now, version status is %s, post %s to github", version.Status, state)
	urlLog := fmt.Sprintf("%s/log?user=%s&service=%s&version=%s", urlHost, service.UserID,
		service.ServiceID, version.VersionID)
	log.Infof("Log getting url: %s", urlLog)
	status := &github.RepoStatus{
		State:       github.String(state),
		TargetURL:   github.String(urlLog),
		Description: github.String(service.Name + " " + version.ErrorMessage),
		Context:     github.String("Cyclone"),
	}

	_, _, err = client.Repositories.CreateStatus(owner, repo, version.Commit, status)
	return err
}
