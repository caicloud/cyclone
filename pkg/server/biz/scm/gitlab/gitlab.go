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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	v4 "gopkg.in/xanzy/go-gitlab.v0"

	c_v1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

const (
	apiPathForGitlabVersion = "%s/api/v4/version"

	v3APIVersion = "v3"

	v4APIVersion = "v4"

	// mergeRefTemplate represents reference template for Gitlab merge request and merge target branch
	mergeRefTemplate = "refs/merge-requests/%d/head:%s"

	openedPullRequestState string = "opened"
)

var gitlabServerAPIVersions = make(map[string]string)

func init() {
	if err := scm.RegisterProvider(v1alpha1.GitLab, NewGitlab); err != nil {
		log.Errorln(err)
	}
}

// NewGitlab news Gitlab v3 or v4 client according to the API version detected from Gitlab server,
func NewGitlab(scmCfg *v1alpha1.SCMSource) (scm.Provider, error) {
	version, err := getAPIVersion(scmCfg)
	if err != nil {
		log.Errorf("Fail to get API version for server %s as %v", scmCfg.Server, err)
		return nil, err
	}
	log.Infof("New Gitlab %s client", version)

	switch version {
	case v3APIVersion:
		client, err := newGitlabV3Client(scmCfg.Server, scmCfg.User, scmCfg.Token)
		if err != nil {
			log.Errorf("fail to new Gitlab v3 client as %v", err)
			return nil, err
		}

		return &V3{scmCfg, client}, nil
	case v4APIVersion:
		v4Client, err := newGitlabV4Client(scmCfg.Server, scmCfg.User, scmCfg.Token)
		if err != nil {
			log.Errorf("fail to new Gitlab v4 client as %v", err)
			return nil, err
		}

		return &V4{scmCfg, v4Client}, nil
	default:
		err = fmt.Errorf("gitlab API version %s is not supported, only support %s and %s", version, v3APIVersion, v4APIVersion)
		log.Errorln(err)
		return nil, err
	}
}

// newGitlabV4Client news Gitlab v4 client by token. If username is empty, use private-token instead of oauth2.0 token.
func newGitlabV4Client(server, username, token string) (*v4.Client, error) {
	var client *v4.Client
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	if len(username) == 0 {
		client = v4.NewClient(httpClient, token)
	} else {
		client = v4.NewOAuthClient(httpClient, token)
	}

	if err := client.SetBaseURL(server + "/api/" + v4APIVersion); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return client, nil
}

// newGitlabV3Client news Gitlab v3 client by token. If username is empty, use private-token instead of oauth2.0 token.
func newGitlabV3Client(server, username, token string) (*gitlab.Client, error) {
	var client *gitlab.Client
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	if len(username) == 0 {
		client = gitlab.NewClient(httpClient, token)
	} else {
		client = gitlab.NewOAuthClient(httpClient, token)
	}

	if err := client.SetBaseURL(server + "/api/" + v3APIVersion); err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return client, nil
}

func getAPIVersion(scmCfg *v1alpha1.SCMSource) (string, error) {
	// Directly get API version if it has been recorded.
	server := scm.ParseServerURL(scmCfg.Server)
	if v, ok := gitlabServerAPIVersions[server]; ok {
		return v, nil
	}

	// Dynamically detect API version if it has not been recorded, and record it for later use.
	version, err := detectAPIVersion(scmCfg)
	if err != nil {
		return "", err
	}

	gitlabServerAPIVersions[server] = version

	return version, nil
}

// versionResponse represents the response of Gitlab version API.
type versionResponse struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
}

func detectAPIVersion(scmCfg *v1alpha1.SCMSource) (string, error) {
	if scmCfg.Token == "" {
		token, err := getOauthToken(scmCfg)
		if err != nil {
			log.Error(err)
			return "", err
		}
		scmCfg.Token = token
	}

	url := fmt.Sprintf(apiPathForGitlabVersion, scmCfg.Server)
	log.Infof("URL: %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error(err)
		return "", err
	}

	// Set headers.
	req.Header.Set("Content-Type", "application/json")
	if scmCfg.User == "" {
		// Use private token when username is empty.
		req.Header.Set("PRIVATE-TOKEN", scmCfg.Token)
	} else {
		// Use Oauth token when username is not empty.
		req.Header.Set("Authorization", "Bearer "+scmCfg.Token)
	}

	// Use client with redirect disabled, then status code will be 302
	// if Gitlab server does not support /api/v4/version request.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return "", convertGitlabError(err, resp)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Fail to close response body as: %v", err)
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Read body error: %v", err)
			return "", err
		}

		gv := &versionResponse{}
		err = json.Unmarshal(body, gv)
		if err != nil {
			log.Error(err)
			return "", err
		}

		log.Infof("Gitlab version is %s, will use %s API", gv.Version, v4APIVersion)
		return v4APIVersion, nil
	case http.StatusNotFound, http.StatusFound:
		log.Infof("Check v4 api version with status code, %d, will use v3", resp.StatusCode)
		return v3APIVersion, nil
	default:
		log.Errorf("Status code of Gitlab API version request is %d", resp.StatusCode)
		return "", convertGitlabError(fmt.Errorf("Gitlab version detection error"), resp)
	}
}

func getOauthToken(scm *v1alpha1.SCMSource) (string, error) {
	if len(scm.User) == 0 || len(scm.Password) == 0 {
		return "", fmt.Errorf("GitLab username or password is missing")
	}

	bodyData := struct {
		GrantType string `json:"grant_type"`
		Username  string `json:"username"`
		Password  string `json:"password"`
	}{
		GrantType: "password",
		Username:  scm.User,
		Password:  scm.Password,
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return "", fmt.Errorf("fail to new request body for token as %s", err.Error())
	}

	// If use the public Gitlab with HTTP protocol, record it.
	if strings.Trim(scm.Server, "/") == "http://gitlab.com" {
		log.Infof("SCM server %s uses HTTP protocol for public Gitlab", scm.Server)
	}

	tokenURL := fmt.Sprintf("%s%s", scm.Server, "/oauth/token")
	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Errorf("Fail to new the request for token as %s", err.Error())
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("Fail to request for token as %s", err.Error())
		return "", convertGitlabError(err, resp)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Errorf("Fail to close response body as: %v", err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Fail to request for token as %s", err.Error())
		return "", err
	}

	if resp.StatusCode/100 == 2 {
		var token oauth2.Token
		err := json.Unmarshal(body, &token)
		if err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}

	err = fmt.Errorf("fail to request for token as %s", body)
	return "", convertGitlabError(err, resp)
}

// MergeCommentEvent ...
type MergeCommentEvent struct {
	ObjectKind string `json:"object_kind"`
	//User       *User  `json:"user"`
	ProjectID int `json:"project_id"`
	Project   struct {
		Name              string `json:"name"`
		Description       string `json:"description"`
		AvatarURL         string `json:"avatar_url"`
		GitSSHURL         string `json:"git_ssh_url"`
		GitHTTPURL        string `json:"git_http_url"`
		Namespace         string `json:"namespace"`
		PathWithNamespace string `json:"path_with_namespace"`
		DefaultBranch     string `json:"default_branch"`
		Homepage          string `json:"homepage"`
		URL               string `json:"url"`
		SSHURL            string `json:"ssh_url"`
		HTTPURL           string `json:"http_url"`
		WebURL            string `json:"web_url"`
		//VisibilityLevel   VisibilityLevelValue `json:"visibility_level"`
	} `json:"project"`
	//Repository       *Repository `json:"repository"`
	ObjectAttributes struct {
		ID         int    `json:"id"`
		Note       string `json:"note"`
		AuthorID   int    `json:"author_id"`
		CreatedAt  string `json:"created_at"`
		UpdatedAt  string `json:"updated_at"`
		ProjectID  int    `json:"project_id"`
		Attachment string `json:"attachment"`
		LineCode   string `json:"line_code"`
		CommitID   string `json:"commit_id"`
		System     bool   `json:"system"`
		//StDiff       *Diff  `json:"st_diff"`
		URL string `json:"url"`
	} `json:"object_attributes"`
	MergeRequest *MergeRequest `json:"merge_request"`
}

// MergeRequest ...
type MergeRequest struct {
	ID             int    `json:"id"`
	IID            int    `json:"iid"`
	ProjectID      int    `json:"project_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	WorkInProgress bool   `json:"work_in_progress"`
	State          string `json:"state"`
	//CreatedAt      *time.Time `json:"created_at"`
	//UpdatedAt      *time.Time `json:"updated_at"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	TargetBranch string `json:"target_branch"`
	SourceBranch string `json:"source_branch"`
	Upvotes      int    `json:"upvotes"`
	Downvotes    int    `json:"downvotes"`
	Author       struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
	} `json:"author"`
	Assignee struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
	} `json:"assignee"`
	SourceProjectID int      `json:"source_project_id"`
	TargetProjectID int      `json:"target_project_id"`
	Labels          []string `json:"labels"`
	Milestone       struct {
		ID          int    `json:"id"`
		Iid         int    `json:"iid"`
		ProjectID   int    `json:"project_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		//CreatedAt   *time.Time `json:"created_at"`
		//UpdatedAt   *time.Time `json:"updated_at"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		DueDate   string `json:"due_date"`
	} `json:"milestone"`
	MergeWhenBuildSucceeds   bool   `json:"merge_when_build_succeeds"`
	MergeStatus              string `json:"merge_status"`
	Subscribed               bool   `json:"subscribed"`
	UserNotesCount           int    `json:"user_notes_count"`
	ShouldRemoveSourceBranch bool   `json:"should_remove_source_branch"`
	ForceRemoveSourceBranch  bool   `json:"force_remove_source_branch"`
	Changes                  []struct {
		OldPath     string `json:"old_path"`
		NewPath     string `json:"new_path"`
		AMode       string `json:"a_mode"`
		BMode       string `json:"b_mode"`
		Diff        string `json:"diff"`
		NewFile     bool   `json:"new_file"`
		RenamedFile bool   `json:"renamed_file"`
		DeletedFile bool   `json:"deleted_file"`
	} `json:"changes"`
	WebURL     string `json:"web_url"`
	LastCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"last_commit"`
}

const (
	// EventTypeHeader represents the header key for event type of Gitlab.
	EventTypeHeader = "X-Gitlab-Event"

	// NoteHookEvent represents comments event.
	NoteHookEvent = "Note Hook"
	// MergeRequestHookEvent represents merge request event.
	MergeRequestHookEvent = "Merge Request Hook"
	// TagPushHookEvent represents tag push event.
	TagPushHookEvent = "Tag Push Hook"
	// PushHookEvent represents commit push event.
	PushHookEvent = "Push Hook"
)

// parseWebhook parses the body from webhook request.
func parseWebhook(r *http.Request) (payload interface{}, err error) {
	eventType := r.Header.Get(EventTypeHeader)
	switch eventType {
	case NoteHookEvent:
		//payload = &gitlab.MergeCommentEvent{}
		// can not unmarshal request body to gitlab.MergeCommentEvent{}
		// due to gitlab.MergeCommentEvent.MergeRequest.CreatedAt's type(*time.Time),
		// parsing time "2018-05-31 02:19:38 UTC" as "2006-01-02T15:04:05Z07:00" will fail.
		payload = &MergeCommentEvent{}
	case MergeRequestHookEvent:
		payload = &gitlab.MergeEvent{}
	case TagPushHookEvent:
		payload = &gitlab.TagEvent{}
	case PushHookEvent:
		payload = &gitlab.PushEvent{}
	default:
		return nil, fmt.Errorf("event type %v not support", eventType)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Fail to read request body")
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

const timeLayout = "2006-01-02 15:04:05 MST"

func parseTime(s string) time.Time {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		log.Warningf("Failed to parse time(%s): %v", s, err)
	}
	return t
}

// ParseEvent parses data from Gitlab events.
func ParseEvent(request *http.Request) *scm.EventData {
	event, err := parseWebhook(request)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	switch event := event.(type) {
	case *gitlab.TagEvent:
		if event.Before != "0000000000000000000000000000000000000000" {
			log.Warning("Skip unsupported action 'Tag updated or deleted' of Gitlab.")
			return nil
		}
		return &scm.EventData{
			Type: scm.TagReleaseEventType,
			Repo: event.Project.PathWithNamespace,
			Ref:  event.Ref,
		}
	case *gitlab.MergeEvent:
		objectAttributes := event.ObjectAttributes
		if objectAttributes.Action != "open" && objectAttributes.Action != "update" {
			log.Warningf("Skip unsupported action %s of Gitlab merge event, only support open and update action.", objectAttributes.Action)
			return nil
		}
		return &scm.EventData{
			Type: scm.PullRequestEventType,
			Repo: event.Project.PathWithNamespace,
			// NOTE: v3 ObjectAttributes has `Iid`, but v4 replaces it with `IID`. This has no effect as both of their json field are `iid`.
			Ref:       fmt.Sprintf(mergeRefTemplate, objectAttributes.Iid, objectAttributes.TargetBranch),
			CommitSHA: objectAttributes.LastCommit.ID,
			Branch:    objectAttributes.TargetBranch,
			CreatedAt: parseTime(objectAttributes.UpdatedAt),
		}
	case *MergeCommentEvent:
		if event.MergeRequest == nil {
			log.Warningln("Only handle comments on merge requests.")
			return nil
		}
		return &scm.EventData{
			Type:      scm.PullRequestCommentEventType,
			Repo:      event.Project.PathWithNamespace,
			Ref:       fmt.Sprintf(mergeRefTemplate, event.MergeRequest.IID, event.MergeRequest.TargetBranch),
			Comment:   event.ObjectAttributes.Note,
			CommitSHA: event.MergeRequest.LastCommit.ID,
			CreatedAt: parseTime(event.ObjectAttributes.CreatedAt),
		}
	case *gitlab.PushEvent:
		if event.After == "0000000000000000000000000000000000000000" {
			log.Warning("Skip unsupported action 'Branch deleted' of Gitlab.")
			return nil
		}
		return &scm.EventData{
			Type:   scm.PushEventType,
			Repo:   event.Project.PathWithNamespace,
			Ref:    event.Ref,
			Branch: event.Ref,
		}
	default:
		log.Warningln("Skip unsupported Gitlab event")
		return nil
	}
}

// transStatus trans api.Status to state and description of gitlab statuses.
func transStatus(status c_v1alpha1.StatusPhase) (string, string) {
	// GitLab : pending, running, success, failed, canceled.
	state := "pending"
	description := ""

	switch status {
	case c_v1alpha1.StatusRunning:
		state = "running"
		description = "The Cyclone CI build is in progress."
	case c_v1alpha1.StatusSucceeded:
		state = "success"
		description = "The Cyclone CI build passed."
	case c_v1alpha1.StatusFailed:
		state = "failed"
		description = "The Cyclone CI build failed."
	case c_v1alpha1.StatusCancelled:
		state = "canceled"
		description = "The Cyclone CI build failed."
	default:
		log.Errorf("not supported state:%s", status)
	}

	return state, description
}

func convertGitlabError(err error, resp interface{}) error {
	if err == nil {
		return nil
	}

	if resp == nil || reflect.ValueOf(resp).IsNil() {
		return cerr.AutoAnalyse(err)
	}

	code := 0
	server := "GitLab"
	switch v := resp.(type) {
	case *gitlab.Response:
		code = v.StatusCode
		server = "GitLab(v3)"
	case *v4.Response:
		code = v.StatusCode
		server = "GitLab(v4)"
	case *http.Response:
		code = v.StatusCode
	}

	if code == http.StatusInternalServerError {
		return cerr.ErrorExternalSystemError.Error(server, err)
	}

	if code == http.StatusUnauthorized {
		return cerr.ErrorExternalAuthorizationFailed.Error(err)
	}

	if code == http.StatusForbidden {
		return cerr.ErrorExternalAuthenticationFailed.Error(err)
	}
	return cerr.AutoAnalyse(err)
}
