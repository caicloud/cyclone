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

package svn

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/nirvana/log"

	c_v1alpha1 "github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

func init() {
	if err := scm.RegisterProvider(v1alpha1.SVN, NewSVN); err != nil {
		log.Errorln(err)
	}
}

const (
	// EventTypeHeader represents the header key for event type of SVN, e.g X-Subversion-Event=Post-Commit.
	EventTypeHeader = "X-Subversion-Event"

	// PostCommitEvent represents post commit type event
	PostCommitEvent = "Post-Commit"
)

// SVN represents the SCM provider of SVN.
type SVN struct {
	scmCfg *v1alpha1.SCMSource
}

// NewSVN new SVN client
func NewSVN(scmCfg *v1alpha1.SCMSource) (scm.Provider, error) {
	return &SVN{scmCfg}, nil
}

// GetToken ...
func (s *SVN) GetToken() (string, error) {
	return s.scmCfg.Token, nil
}

// ListRepos ...
func (s *SVN) ListRepos() ([]scm.Repository, error) {
	return nil, cerr.ErrorNotImplemented.Error("list svn repos")
}

// ListBranches ...
func (s *SVN) ListBranches(repo string) ([]string, error) {
	return nil, cerr.ErrorNotImplemented.Error("list svn branches")
}

// ListTags ...
func (s *SVN) ListTags(repo string) ([]string, error) {
	return nil, cerr.ErrorNotImplemented.Error("list svn tags")
}

// ListDockerfiles ...
func (s *SVN) ListDockerfiles(repo string) ([]string, error) {
	return nil, cerr.ErrorNotImplemented.Error("list svn dockerfiles")
}

// CreateStatus ...
func (s *SVN) CreateStatus(status c_v1alpha1.StatusPhase, targetURL, repoURL, commitSha string) error {
	return cerr.ErrorNotImplemented.Error("create status")
}

// GetPullRequestSHA ...
func (s *SVN) GetPullRequestSHA(repoURL string, number int) (string, error) {
	return "", cerr.ErrorNotImplemented.Error("get pull request SHA")
}

// CheckToken ...
func (s *SVN) CheckToken() error {
	return nil
}

// CreateWebhook ...
func (s *SVN) CreateWebhook(repo string, webhook *scm.Webhook) error {
	return nil
}

// DeleteWebhook ...
func (s *SVN) DeleteWebhook(repo string, webhookURL string) error {
	return nil
}

// ParseEvent parses data from SVN events, only support Post-Commit.
func ParseEvent(request *http.Request) *scm.EventData {
	event, err := parseWebhook(request)
	if err != nil {
		log.Errorln(err)
		return nil
	}

	switch event := event.(type) {
	case *PostCommit:
		return &scm.EventData{
			Type: scm.PostCommitEventType,
			Repo: event.RepoUUID,
			Ref:  event.Revision,
		}
	default:
		log.Warningln("Skip unsupported Gitlab event")
		return nil
	}
}

// parseWebhook parses the body from webhook requeset.
func parseWebhook(r *http.Request) (payload interface{}, err error) {
	eventType := r.Header.Get(EventTypeHeader)
	switch eventType {
	case PostCommitEvent:
		payload = &PostCommit{}
	default:
		return nil, fmt.Errorf("event type %v not support", eventType)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read request body")
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

// PostCommit is PostCommit Event type body struct
type PostCommit struct {
	// RepoUUID represents svn repository uuid of this trigger commit
	// Exec command:
	// `svn info --show-item repos-uuid --username {user} --password {password} --non-interactive
	// --trust-server-cert-failures unknown-ca,cn-mismatch,expired,not-yet-valid,other
	// --no-auth-cache {url}`
	// or
	// `svnlook uuid {repoPath}`
	// can get the repo uuid
	RepoUUID string `json:"repoUUID"`
	// RepoName represents the revision of this trigger commit
	Revision string `json:"revision"`
}
