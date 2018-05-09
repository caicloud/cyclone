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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/golang/glog"
	gitlab "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	mgo "gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/pkg/scm"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/store"
)

// gitLabServer represents the server address for public GitLab.
const gitLabServer = "https://gitlab.com"

// GitLab represents the SCM provider of GitLab.
type GitLab struct{}

func init() {
	if err := scm.RegisterProvider(api.GitLab, new(GitLab)); err != nil {
		log.Errorln(err)
	}

	if err := scm.RegisterManager(api.GITLAB, new(GitLabManager)); err != nil {
		log.Errorln(err)
	}
}

// GetToken gets the token by the username and password of SCM config.
func (g *GitLab) GetToken(scm *api.SCMConfig) (string, error) {
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

// CheckToken checks whether the token has the authority of repo by trying ListRepos with the token.
func (g *GitLab) CheckToken(scm *api.SCMConfig) bool {
	if _, err := g.ListRepos(scm); err != nil {
		return false
	}
	return true
}

// ListRepos lists the repos by the SCM config.
func (g *GitLab) ListRepos(scm *api.SCMConfig) ([]api.Repository, error) {
	client, err := newGitLabClient(scm.Server, scm.Username, scm.Token)
	if err != nil {
		return nil, err
	}

	opt := &gitlab.ListProjectsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: listPerPageOpt,
		},
	}

	// Get all pages of results.
	var allProjects []*gitlab.Project
	for {
		projects, resp, err := client.Projects.ListProjects(opt)
		if err != nil {
			return nil, err
		}

		allProjects = append(allProjects, projects...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	repos := make([]api.Repository, len(allProjects))
	for i, repo := range allProjects {
		repos[i].Name = repo.PathWithNamespace
		repos[i].URL = repo.HTTPURLToRepo
	}

	return repos, nil
}

// ListBranches lists the branches for specified repo.
func (g *GitLab) ListBranches(scm *api.SCMConfig, repo string) ([]string, error) {
	client, err := newGitLabClient(scm.Server, scm.Username, scm.Token)
	if err != nil {
		return nil, err
	}

	branches, _, err := client.Branches.ListBranches(repo)
	if err != nil {
		log.Errorf("Fail to list branches for %s", repo)
		return nil, err
	}

	branchNames := make([]string, len(branches))
	for i, branch := range branches {
		branchNames[i] = branch.Name
	}

	return branchNames, nil
}

// newGitLabClient news GitLab client by token.If username is empty, use private-token instead of oauth2.0 token.
func newGitLabClient(server, username, token string) (*gitlab.Client, error) {
	var client *gitlab.Client

	if len(username) == 0 {
		client = gitlab.NewClient(nil, token)
	} else {
		client = gitlab.NewOAuthClient(nil, token)
	}

	if err := client.SetBaseURL(server + "/api/v3/"); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return client, nil
}

// GitLabManager represents the manager for gitlab.
type GitLabManager struct {
	DataStore *store.DataStore
}

// Authcallback is the callback handler.
func (g *GitLabManager) Authcallback(code, state string) (string, error) {
	if code == "" || state == "" {
		return "", fmt.Errorf("code: %s or state: %s is nil", code, state)
	}

	uiPath := osutil.GetStringEnv(cloud.ConsoleWebEndpoint, "")
	redirectURL := fmt.Sprintf("%s/devops/workspace/add?type=gitlab&code=%s&state=%s", uiPath, code, state)
	token, err := g.getToken(code, state)
	if err != nil {
		return "", err
	}
	// add token, username, server to redirectURL
	userName, server, err := g.getUserInfo(token)
	redirectURL = redirectURL + fmt.Sprintf("&token=%s&username=%s&server=%s",
		token.AccessToken, userName, server)
	return redirectURL, nil
}

func (g *GitLabManager) getToken(code, state string) (*oauth2.Token, error) {
	config, err := getConfig(api.GITLAB)
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

// setToken sets the token.
func (g *GitLabManager) setToken(code, state string) error {
	// Get the oauth config to request token.
	config, err := getConfig(api.GITLAB)
	if err != nil {
		return err
	}

	// To communicate with gitlab or other scm to get token.
	var token *oauth2.Token
	token, err = config.Exchange(oauth2.NoContext, code) // Post a token request and receive token.
	if err != nil {
		return err
	}

	if !token.Valid() {
		return fmt.Errorf("token invalid. Got: %#v", token)
	}

	return nil
}

func (g *GitLabManager) getUserInfo(token *oauth2.Token) (string, string, error) {
	accessToken := token.AccessToken
	gitlabServer := osutil.GetStringEnv(cloud.GitlabURL, "http://192.168.21.100:10080")
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

// GetRepos gets the list of repositories with token from gitlab.
func (g *GitLabManager) GetRepos(projectID string) (repos []api.Repository, username string, avatarURL string, err error) {
	token, err := g.DataStore.Findtoken(projectID, api.GITLAB)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, username, avatarURL, httperror.ErrorContentNotFound.Format("token", projectID, api.GitHub)
		}
		return nil, username, avatarURL, err
	}

	gitlabServer := osutil.GetStringEnv(cloud.GitlabURL, "https://gitlab.com")
	client := gitlab.NewOAuthClient(nil, token.Token.AccessToken)
	client.SetBaseURL(gitlabServer + "/api/v3/")

	opt := &gitlab.ListProjectsOptions{}
	projects, _, err := client.Projects.ListProjects(opt)
	if err != nil {
		return nil, username, avatarURL, err
	}

	repos = make([]api.Repository, len(projects))
	for i, repo := range projects {
		repos[i].Name = repo.Name
		repos[i].URL = repo.HTTPURLToRepo
	}

	optUsers := &gitlab.ListUsersOptions{Username: &username}
	users, _, err := client.Users.ListUsers(optUsers)
	if err != nil {
		return repos, username, avatarURL, err
	}
	avatarURL = users[0].AvatarURL

	return repos, username, avatarURL, nil
}

// LogOut logs out and deletes the token.
func (g *GitLabManager) LogOut(projectID string) error {
	return g.DataStore.DeleteToken(projectID, api.GITLAB)
}

// GetAuthCodeURL gets the URL for token request.
func (g *GitLabManager) GetAuthCodeURL(state string, scmType string) (string, error) {
	return getAuthCodeURL(state, scmType)
}
