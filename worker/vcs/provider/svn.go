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
	"encoding/base64"
	"fmt"
	"path"
	"strings"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/executil"
	"github.com/caicloud/cyclone/pkg/log"
	steplog "github.com/caicloud/cyclone/worker/log"
)

// Svn is the type for svn provider.
type Svn struct{}

// NewSvn returns a new Svn worker.
func NewSvn() *Svn {
	return &Svn{}
}

// Ping check whether svn repo is valid
func (s *Svn) Ping(url, destPath string, event *api.Event) error {

	dir := path.Dir(destPath)
	// git ls-remote url --heads HEAD
	args := []string{"ls", url, "--depth", "empty",
		"--username", event.Service.Repository.Username,
		"--password", getPwdFromBase64(event.Service.Repository.Password),
		"--non-interactive", "--trust-server-cert", "--no-auth-cache"}

	output, err := executil.RunInDir(dir, "svn", args...)
	if event.Version.VersionID != "" {
		fmt.Fprintf(steplog.Output, "%s", string(output))
	}
	if err != nil {
		log.ErrorWithFields("Error when check valid", log.Fields{"error": err})
		return err
	}

	log.InfoWithFields("valid svn repository.", log.Fields{"url": url})
	return nil
}

// Clone implements VCS interface.
func (s *Svn) Clone(url, destPath string, event *api.Event) error {
	log.InfoWithFields("About to svn checkout repository.", log.Fields{"url": url, "destPath": destPath})

	args := []string{"checkout",
		"--username", event.Service.Repository.Username,
		"--password", getPwdFromBase64(event.Service.Repository.Password),
		"--non-interactive", "--trust-server-cert", "--no-auth-cache",
		url, destPath}
	output, err := executil.RunInDir("./", "svn", args...)
	if event.Version.VersionID != "" {
		fmt.Fprintf(steplog.Output, "%q", string(output))
	}

	if err != nil {
		log.ErrorWithFields("Error when clone", log.Fields{"error": err})
	} else {
		log.InfoWithFields("Successfully svn checkout repository.", log.Fields{"url": url, "destPath": destPath})
	}
	return err
}

// NewTagFromLatest implements VCS interface.
func (s *Svn) NewTagFromLatest(repoPath string, event *api.Event) error {
	service := event.Service
	version := event.Version
	if !strings.Contains(version.URL, "/trunk") {
		return fmt.Errorf("not standard SVN dirs, cannot create tag")
	}

	tagURL := strings.Split(version.URL, "/trunk")[0] + "/tags/" + version.Name + "/"
	log.Infof("trunk[%s] tag[%s]", version.URL, tagURL)
	args := []string{"copy", version.URL, tagURL, "-m", "Cyclone auto tag " + version.Name,
		"--username", service.Repository.Username,
		"--password", getPwdFromBase64(event.Service.Repository.Password),
		"--non-interactive", "--trust-server-cert", "--no-auth-cache"}

	output, err := executil.RunInDir(repoPath, "svn", args...)
	log.Infof("Command output: %+v", string(output))
	if err == nil {
		log.InfoWithFields("Successfully svn create tag.", log.Fields{"repoPath": repoPath, "version": version})
	}

	return err
}

// CheckoutTag implements VCS interface.
func (s *Svn) CheckoutTag(repoPath string, tag string) error {
	args := []string{"tags/" + tag}
	output, err := executil.RunInDir(repoPath, "cd", args...)

	// TODO: Need a force checkout if tree is dirty (in what cases could
	// local tree be dirty?
	log.Debugf("Command output: %+v", string(output))
	if err == nil {
		log.InfoWithFields("Successfully checked out to svn tag.", log.Fields{"repoPath": repoPath, "tag": tag})
	}
	return err
}

// GetTagCommit implements VCS interface.
func (s *Svn) GetTagCommit(repoPath string, tag string) (string, error) {
	args := []string{tag}
	output, err := executil.RunInDir(repoPath+"/tags", "ls", args...)
	if err != nil {
		log.InfoWithFields("failed checked out to svn tag.", log.Fields{"repoPath": repoPath, "tag": tag})
	}
	return strings.Trim(string(output), "\n"), err
}

// CheckOutByCommitID check out code in repo by special commit id.
func (s *Svn) CheckOutByCommitID(commitID string, repoPath string, event *api.Event) error {
	args := []string{"update", "-r", commitID,
		"--username", event.Service.Repository.Username,
		"--password", getPwdFromBase64(event.Service.Repository.Password),
		"--non-interactive", "--trust-server-cert", "--no-auth-cache"}
	output, err := executil.RunInDir(repoPath, "svn", args...)
	fmt.Fprintf(steplog.Output, "%q\n", string(output))

	if err != nil {
		log.ErrorWithFields("Error when svn check out by commitID", log.Fields{"error": err})
		return err
	}
	return nil
}

// IsCommitToSpecialURL gets if the commit is to a specific url.
func (s *Svn) IsCommitToSpecialURL(commitID string, service *api.Service) (bool, string, error) {
	args := []string{"log", service.Repository.URL, "-r", commitID,
		"--username", service.Repository.Username,
		"--password", getPwdFromBase64(service.Repository.Password),
		"--non-interactive", "--trust-server-cert", "--no-auth-cache"}
	output, err := executil.RunInDir("./", "svn", args...)
	log.Info(string(output))

	if err != nil {
		log.ErrorWithFields("Error when call IsCommitToSpecialURL", log.Fields{"error": err})
		return false, string(output), err

	}

	return strings.Contains(string(output), "r"+commitID), string(output), nil
}

func getPwdFromBase64(pwdBase64 string) string {
	var pwd string
	pwdB, err := base64.StdEncoding.DecodeString(pwdBase64)
	if err != nil {
		pwd = pwdBase64
	} else {
		pwd = string(pwdB)
	}
	return pwd
}
