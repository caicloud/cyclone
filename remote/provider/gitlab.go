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
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/store"
	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

// GitLab if the type for GitLab remote provider.
type GitLab struct {
}

// NewGitLab returns a new GitLab remoter.
func NewGitLab() *GitLab {
	return &GitLab{}
}

// Pack the information into oauth.config that is used to get token
// ClientID、ClientSecret,these values use to assemble the token request url and
// there values come from gitlab or other by registering some information
func (g *GitLab) getConf() (*oauth2.Config, error) {
	//cyclonePath http request listen address
	cyclonePath := osutil.GetStringEnv(CYCLONE_SERVER_HOST, "http://127.0.0.1:7099")
	clientID := osutil.GetStringEnv("CLIENTID_GITLAB", "")
	clientSecret := osutil.GetStringEnv("CLIENTIDSECRET_GITLAB", "")
	gitlabServer := osutil.GetStringEnv("SERVER_GITLAB", "https://gitlab.com")
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("%s/api/%s/remotes/%s/authcallback", cyclonePath, api.APIVersion, "gitlab"),
		Scopes:       []string{"api"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", gitlabServer),
			TokenURL: fmt.Sprintf("%s/oauth/token", gitlabServer),
		},
	}, nil
}

// GetTokenQuestURL gets the URL for token request.
func (g *GitLab) GetTokenQuestURL(userID string) (string, error) {
	//get a object to request token
	conf, err := g.getConf()
	if err != nil {
		return "", err
	}
	//get the request url and send to the gitlab
	url := conf.AuthCodeURL(userID) //set state random  need add into db
	log.InfoWithFields("cyclone receives creating token request",
		log.Fields{"request url": url})

	if !strings.Contains(url, "gitlab") {
		log.ErrorWithFields("Unable to get the  url", log.Fields{"user_id": userID})
		return "", fmt.Errorf("Unable to get the  url")
	}
	return url, nil
}

// Authcallback is the callback handler.
func (g *GitLab) Authcallback(code, state string) (string, error) {
	if code == "" || state == "" {
		return "", fmt.Errorf("code or state is nil, in %s", "gitlab")
	}

	//caicloud web address,eg caicloud.io
	uiPath := osutil.GetStringEnv("CONSOLE_WEB_ENDPOINT", "http://localhost:8000")
	redirectURL := fmt.Sprintf("%s/cyclone/add?type=gitlab&code=%s&state=%s", uiPath, code, state)

	//sync to get token
	go g.getToken(code, state)
	return redirectURL, nil
}

// get token by using code from gitlab
func (g *GitLab) getToken(code, state string) error {
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

	// To communication with gitlab  to get token.
	var tok *oauth2.Token
	tok, err = conf.Exchange(oauth2.NoContext, code) //post a token request and receive token
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
		Vsc:      "gitlab",
		Vsctoken: *tok,
	}

	ds := store.NewStore()
	defer ds.Close()

	_, err = ds.FindtokenByUserID(state, "gitlab")
	if err != nil {
		err = ds.NewTokenDocument(&vcstoken)
		if err != nil {
			log.ErrorWithFields("NewTokenDocument", log.Fields{"user_id": state, "token": tok, "error": err})
			return err
		}
	} else {
		err = ds.UpdateToken(&vcstoken)
		if err != nil {
			log.ErrorWithFields("UpdateToken", log.Fields{"user_id": state, "token": tok, "error": err})
			return err
		}
	}

	return nil
}

// GetRepos gets token by using code from gitlab.
func (g *GitLab) GetRepos(userID string) (repos []api.Repo, username string, avatarURL string, err error) {
	ds := store.NewStore()
	defer ds.Close()

	tok, err := ds.FindtokenByUserID(userID, "gitlab")
	if err != nil {
		log.ErrorWithFields("find token failed", log.Fields{"user_id": userID, "error": err})
		return repos, username, avatarURL, err
	}

	gitlabServer := osutil.GetStringEnv("SERVER_GITLAB", "https://gitlab.com")
	client := gitlab.NewOAuthClient(nil, tok.Vsctoken.AccessToken)
	client.SetBaseURL(gitlabServer + "/api/v3/")

	opt := &gitlab.ListProjectsOptions{}
	projects, _, err := client.Projects.ListProjects(opt)
	if err != nil {
		log.Errorf("ListProjects returned error: %v", err)
		return repos, username, avatarURL, err
	}

	repos = make([]api.Repo, len(projects))
	for i, repo := range projects {
		repos[i].Name = repo.Name
		repos[i].URL = repo.HTTPURLToRepo
		repos[i].Owner = (*repo.Namespace).Name
		if repo.Owner != nil {
			username = repos[i].Owner
		}
	}

	optUsers := &gitlab.ListUsersOptions{Username: &username}
	users, _, err := client.Users.ListUsers(optUsers)
	if err != nil {
		log.Errorf("ListProjects returned error: %v", err)
		return repos, username, avatarURL, err
	}
	avatarURL = users[0].AvatarURL

	return repos, username, avatarURL, nil
}

// LogOut logs out.
func (g *GitLab) LogOut(userID string) error {
	// Find the token by userid and code repository.
	ds := store.NewStore()
	defer ds.Close()
	_, err := ds.FindtokenByUserID(userID, "gitlab")
	if err != nil {
		log.ErrorWithFields("find token failed", log.Fields{"user_id": userID, "error": err})
		return err
	}

	// Remove the token saved in DB.
	err = ds.RemoveTokeninDB(userID, "gitlab")
	if err != nil {
		log.ErrorWithFields("remove token failed", log.Fields{"user_id": userID, "error": err})
		return err
	}

	return nil
}

// CreateHook is a helper to register webhook.
func (g *GitLab) CreateHook(service *api.Service) error {
	webhooktype := service.Repository.Webhook
	if webhooktype == "" {
		return fmt.Errorf("no need webhook registry")
	}

	if webhooktype == api.GITLAB {
		url := getHookURL(webhooktype, service.ServiceID)
		if url == "" {
			log.Infof("url is empty", log.Fields{"user_id": service.UserID})
			return nil
		}

		ds := store.NewStore()
		defer ds.Close()

		tok, err := ds.FindtokenByUserID(service.UserID, api.GITLAB)
		if err != nil {
			log.ErrorWithFields("find token failed", log.Fields{"user_id": service.UserID, "error": err})
			return err
		}

		gitlabServer := osutil.GetStringEnv("SERVER_GITLAB", "https://gitlab.com")
		client := gitlab.NewOAuthClient(nil, tok.Vsctoken.AccessToken)
		client.SetBaseURL(gitlabServer + "/api/v3/")

		var hook gitlab.AddProjectHookOptions
		state := true
		hook.URL = &url
		hook.PushEvents = &state
		hook.MergeRequestsEvents = &state
		hook.TagPushEvents = &state

		onwer, name := parseURL(service.Repository.URL)
		_, _, err = client.Projects.AddProjectHook(onwer+"/"+name, &hook)
		return err
	}
	log.WarnWithFields("not support vcs repository", log.Fields{"vcs repository": webhooktype})
	return fmt.Errorf("not support vcs repository in create webhook")
}

// DeleteHook is a helper to unregister webhook.
func (g *GitLab) DeleteHook(service *api.Service) error {
	webhooktype := service.Repository.Webhook
	if webhooktype == "" {
		return fmt.Errorf("no need webhook registry")
	}

	if webhooktype == api.GITLAB {
		url := getHookURL(webhooktype, service.ServiceID)
		if url == "" {
			log.Infof("url is empty", log.Fields{"user_id": service.UserID})
			return nil
		}

		ds := store.NewStore()
		defer ds.Close()

		tok, err := ds.FindtokenByUserID(service.UserID, webhooktype)
		if err != nil {
			log.ErrorWithFields("find token failed", log.Fields{"user_id": service.UserID, "error": err})
			return err
		}

		gitlabServer := osutil.GetStringEnv("SERVER_GITLAB", "https://gitlab.com")
		client := gitlab.NewOAuthClient(nil, tok.Vsctoken.AccessToken)
		client.SetBaseURL(gitlabServer + "/api/v3/")

		owner, name := parseURL(service.Repository.URL)
		hooks, _, err := client.Projects.ListProjectHooks(owner+"/"+name, nil)
		if err != nil {
			return err
		}

		var hook *gitlab.ProjectHook
		hasFoundHook := false
		for _, hook = range hooks {
			if strings.HasPrefix(hook.URL, url) {
				hasFoundHook = true
				break
			}
		}
		if hasFoundHook {
			_, err = client.Projects.DeleteProjectHook(owner+"/"+name, hook.ID)
			return err
		}
		return nil
	}
	log.WarnWithFields("not support vcs repository", log.Fields{"vcs repository": webhooktype})
	return fmt.Errorf("not support vcs repository in delete webhook")
}

// PostCommitStatus posts Commit Status To gitlab.
func (g *GitLab) PostCommitStatus(service *api.Service, version *api.Version) error {
	// Check if gitlab webhook has set.
	if service.Repository.Webhook != api.GITLAB {
		return fmt.Errorf("vcs gitlab webhook hasn't set")
	}

	// Check if has set commitID.
	if version.Commit == "" {
		return fmt.Errorf("commit hasn't set")
	}

	// Get token.
	ds := store.NewStore()
	defer ds.Close()
	tok, err := ds.FindtokenByUserID(service.UserID, api.GITLAB)
	if err != nil {
		log.ErrorWithFields("find token failed", log.Fields{"user_id": service.UserID, "error": err})
		return err
	}

	// Post commit status.
	owner, name := parseURL(service.Repository.URL)
	urlHost := osutil.GetStringEnv(CYCLONE_SERVER_HOST, "https://fornax-canary.caicloud.io")

	var state string
	if version.Status == api.VersionHealthy {
		state = "success"
	} else if version.Status == api.VersionFailed || version.Status == api.VersionCancel {
		state = "failed"
	} else {
		state = "pending"
	}

	log.Infof("Now, version status is %s, post %s to gitlab", version.Status, state)
	urlLog := fmt.Sprintf("%s/log?user=%s&service=%s&version=%s", urlHost, service.UserID,
		service.ServiceID, version.VersionID)
	log.Infof("Log getting url: %s", urlLog)

	gitlabServer := osutil.GetStringEnv("SERVER_GITLAB", "https://gitlab.com")

	cmd := fmt.Sprintf(`curl -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer %s' 
						-d '{"state":"%s", "name":"%s","target_url":"%s","description":"%s"}'  
						%s/api/v3/projects/%s%%2F%s/statuses/%s`, tok.Vsctoken.AccessToken, state,
		"Cyclone", urlLog, version.ErrorMessage, gitlabServer, owner, name, version.Commit)
	return system(cmd)
}

// system func excute the shell scripts.
func system(s string) error {
	cmd := exec.Command("/bin/sh", "-c", s)
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}
