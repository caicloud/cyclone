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
	"time"

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	executil "github.com/caicloud/cyclone/pkg/util/exec"
	"github.com/caicloud/cyclone/pkg/worker/scm"
)

// Svn is the type for svn provider.
type Svn struct{}

func init() {
	if err := scm.RegisterProvider(api.SVN, new(Svn)); err != nil {
		log.Error(err)
	}
}

// spilitToken split svn token to username and password.
func (s *Svn) spilitToken(token string) (string, string, error) {
	userPwd := strings.Split(token, api.SVNUsernPwdSep)
	if len(userPwd) != 2 {
		err := fmt.Errorf("split token fail as the length of userPwd equals %v", len(userPwd))
		return "", "", err
	}

	return userPwd[0], userPwd[1], nil
}

// Clone implements SCMProvider interface.
func (s *Svn) Clone(token, url, ref, destPath string) (string, error) {
	log.Infof("About to svn checkout repository, url: %s, dest path: %s, ref: %s", url, destPath, ref)

	var revision string
	if strings.HasPrefix(ref, api.SVNPostCommitRefPrefix) {
		revision = strings.TrimPrefix(ref, api.SVNPostCommitRefPrefix)
	} else {
		url = strings.TrimSuffix(url, "/") + "/" + ref
	}

	username, password, err := s.spilitToken(token)
	if err != nil {
		return "", err
	}

	var args []string
	if revision == "" {
		args = []string{"checkout", "--username", username, "--password", password,
			"--non-interactive", "--trust-server-cert-failures", "unknown-ca,cn-mismatch,expired,not-yet-valid,other", "--no-auth-cache", url, destPath}
	} else {
		args = []string{"checkout", "--username", username, "--password", password, "--revision", revision,
			"--non-interactive", "--trust-server-cert-failures", "unknown-ca,cn-mismatch,expired,not-yet-valid,other", "--no-auth-cache", url, destPath}
	}

	output, err := executil.RunInDir("./", "svn", args...)
	if err != nil {
		log.Errorf("fail to clone as %v", err)
		return "", err
	}
	log.Infof("Successfully svn checkout repository, url: %s, dest path: %s", url, destPath)

	return string(output), err
}

// GetCommitID implements SCMProvider interface. returns latest commit id.
func (s *Svn) GetCommitID(repoPath string) (string, error) {
	log.Infof("About to get commit info in repo path", repoPath)
	args := []string{"info", "--non-interactive", "--trust-server-cert-failures", "unknown-ca", "--no-auth-cache"}
	output, err := executil.RunInDir(repoPath, "svn", args...)
	if err != nil {
		log.Infof("failed get commit reversion in repo path", repoPath)
	}

	var id string
	lineinfos := strings.Split(string(output), "\n")
	for _, lineinfo := range lineinfos {
		if strings.HasPrefix(lineinfo, "Last Changed Rev:") {
			id = strings.TrimSpace(strings.TrimLeft(lineinfo, "Last Changed Rev:"))
		}
	}

	return id, nil
}

// GetCommitLog implements SCMProvider interface.
func (s *Svn) GetCommitLog(repoPath string) api.CommitLog {
	commitLog := api.CommitLog{}
	log.Infof("About to get commit log in repo path %s", repoPath)
	args := []string{"info", "--non-interactive", "--trust-server-cert-failures", "unknown-ca", "--no-auth-cache"}
	output, err := executil.RunInDir(repoPath, "svn", args...)
	if err != nil {
		log.Errorf("get commit log error as %v", err)
	}

	lineinfos := strings.Split(string(output), "\n")
	for _, lineinfo := range lineinfos {
		if strings.HasPrefix(lineinfo, "Last Changed Rev:") {
			commitLog.ID = strings.TrimSpace(strings.TrimPrefix(lineinfo, "Last Changed Rev:"))
		}
		if strings.HasPrefix(lineinfo, "Last Changed Author:") {
			commitLog.Author = strings.TrimSpace(strings.TrimPrefix(lineinfo, "Last Changed Author:"))
		}
		if strings.HasPrefix(lineinfo, "Last Changed Date:") {
			date := strings.TrimSpace(strings.TrimPrefix(lineinfo, "Last Changed Date:"))
			t, err := time.Parse("2006-01-02 15:04:05 -0700 (Mon, 02 Jan 2006)", date)
			if err != nil {
				log.Errorf("parse last changed date error as %v", err)
			}
			commitLog.Date = t
		}
	}

	message, err := s.getCommitMessage(repoPath)
	if err != nil {
		log.Errorf("get commit message error as %v", err)
	}
	commitLog.Message = message
	return commitLog
}

// getCommitMessage get the latest commit message.
func (s *Svn) getCommitMessage(repoPath string) (string, error) {
	var message string
	log.Infof("About to get commit message in repo path", repoPath)
	args := []string{"log", "-r", "COMMITTED", "--xml", "--non-interactive", "--trust-server-cert-failures", "unknown-ca", "--no-auth-cache"}
	output, err := executil.RunInDir(repoPath, "svn", args...)
	if err != nil {
		log.Errorf("failed get commit message, %v", err)
	}

	lineinfos := strings.Split(string(output), "\n")
	for _, lineinfo := range lineinfos {
		if strings.HasPrefix(lineinfo, "<msg>") && strings.HasSuffix(lineinfo, "</msg>") {
			message = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(lineinfo, "<msg>"), "</msg>"))
		}
	}

	return message, nil
}
