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
	"fmt"
	"strings"

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
	executil "github.com/caicloud/cyclone/pkg/util/exec"
	"github.com/caicloud/cyclone/pkg/util/http/errors"
)

func init() {
	if err := scm.RegisterProvider(api.SVN, NewSVN); err != nil {
		log.Errorln(err)
	}
}

// SVN represents the SCM provider of SVN.
type SVN struct {
	scmCfg *api.SCMConfig
}

func NewSVN(scmCfg *api.SCMConfig) (scm.SCMProvider, error) {
	return &SVN{scmCfg}, nil
}

func (s *SVN) spilitToken(token string) (string, string, error) {
	userPwd := strings.Split(token, api.SVNUsernPwdSep)
	if len(userPwd) != 2 {
		err := fmt.Errorf("split token fail as the length of userPwd equals %v", len(userPwd))
		return "", "", err
	}

	return userPwd[0], userPwd[1], nil
}

func (s *SVN) GetToken() (string, error) {
	return s.scmCfg.Username + api.SVNUsernPwdSep + s.scmCfg.Password, nil
}

func (s *SVN) ListRepos() ([]api.Repository, error) {
	return nil, errors.ErrorNotImplemented.Error("list svn repos")
}

func (s *SVN) ListBranches(repo string) ([]string, error) {
	return nil, errors.ErrorNotImplemented.Error("list svn branches")
}

func (s *SVN) ListTags(repo string) ([]string, error) {
	return nil, errors.ErrorNotImplemented.Error("list svn tags")
}

func (s *SVN) CheckToken() bool {
	//username, password, err := s.spilitToken(scm.Token)
	//if err != nil {
	//	return false
	//}
	//fmt.Println(username)
	//fmt.Println(password)
	//
	//url := scm.Server
	//args := []string{"list", "--username", username, "--password", password,
	//	"--non-interactive", "--trust-server-cert-failures", "unknown-ca", "--no-auth-cache", url}
	//_, err = executil.RunInDir("./", "svn", args...)
	//if err != nil {
	//	log.Errorf("Error when list repos as : %v", err)
	//	return false
	//}
	return true
}

func (s *SVN) NewTagFromLatest(tagName, description, commitID, url string) error {
	username, password, err := s.spilitToken(s.scmCfg.Token)
	if err != nil {
		return err
	}

	if !strings.Contains(url, "/trunk") {
		return fmt.Errorf("not standard SVN dirs, cannot create tag")
	}

	tagURL := strings.Split(url, "/trunk")[0] + "/tags/" + tagName + "/"
	log.Infof("trunk[%s] tag[%s]", url, tagURL)
	args := []string{"copy", url, tagURL, "-m", "Cyclone auto tag " + tagName,
		"--username", username, "--password", password,
		"--non-interactive", "--trust-server-cert-failures", "unknown-ca", "--no-auth-cache"}

	output, err := executil.RunInDir("./", "svn", args...)
	log.Infof("Command output: %+v", string(output))
	if err == nil {
		log.Infof("Successfully svn create tag.")
	}
	return err
}

func (s *SVN) CreateWebHook(repoURL string, webHook *scm.WebHook) error {
	return errors.ErrorNotImplemented.Error("create svn webhook")
}
func (s *SVN) DeleteWebHook(repoURL string, webHookUrl string) error {
	return errors.ErrorNotImplemented.Error("delete svn webhook")
}

func (s *SVN) GetTemplateType(repo string) (string, error) {
	return "", errors.ErrorNotImplemented.Error("get svn template type")
}

// CreateStatus generate a new status for repository.
func (g *SVN) CreateStatus(recordStatus api.Status, targetURL, repoURL, statusesURL string) error {
	return errors.ErrorNotImplemented.Error("create svn status")
}

func (s *SVN) GetPullRequestSHA(repoURL string, number int) (string, error) {
	return "", errors.ErrorNotImplemented.Error("get pull request sha")
}
