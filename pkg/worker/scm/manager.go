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

package scm

import (
	"encoding/base64"
	"fmt"
	neturl "net/url"
	"path/filepath"
	"regexp"
	"time"

	"github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/pkg/api"
	pathutil "github.com/caicloud/cyclone/pkg/util/path"
)

const (
	// cloneDir represents the dir which the repo clone to.
	cloneDir = "/tmp/code"

	repoNameRegexp = `^http[s]?://(?:git[\w]+\.com|[\d]+\.[\d]+\.[\d]+\.[\d]+:[\d]+|localhost:[\d]+)/([\S]*)\.git$`
)

var (
	maxRetries   = 8
	initialDelay = 2 * time.Second
)

// scmProviders represents the set of SCM providers.
var scmProviders map[api.SCMType]SCMProvider

type SCMProvider interface {
	Clone(token, url, ref, destPath string) (string, error)
	GetCommitID(repoPath string) (string, error)
	GetCommitLog(repoPath string) api.CommitLog
}

func init() {
	scmProviders = make(map[api.SCMType]SCMProvider)
}

// RegisterProvider registers SCM providers.
func RegisterProvider(scmType api.SCMType, provider SCMProvider) error {
	if _, ok := scmProviders[scmType]; ok {
		return fmt.Errorf("SCM provider %s already exists.", scmType)
	}

	scmProviders[scmType] = provider
	return nil
}

// GetSCMProvider gets the SCM provider by the type.
func GetSCMProvider(scmType api.SCMType) (SCMProvider, error) {
	provider, ok := scmProviders[scmType]
	if !ok {
		return nil, fmt.Errorf("unsupported SCM type %s", scmType)
	}

	return provider, nil
}

// GetCloneDir returns the clone dir.
func GetCloneDir() string {
	return cloneDir
}

func GetRepoName(codeSource *api.CodeSource) (string, error) {
	if codeSource.Type == api.SVN {
		return codeSource.SVN.Url, nil
	}

	gitSource, err := api.GetGitSource(codeSource)
	if err != nil {
		logdog.Errorf(err.Error())
		return "", err
	}

	return getRepoNameByURL(gitSource.Url)
}

// input   url      =  `https://github.com/caicloud/cyclone.git`
// output  reponame =  `caicloud/cyclone`
func getRepoNameByURL(url string) (string, error) {
	r := regexp.MustCompile(repoNameRegexp)
	results := r.FindStringSubmatch(url)
	if len(results) < 2 {
		return url, nil
	}
	return results[1], nil
}

func GetCommitID(codeSource *api.CodeSource, folder string) (string, error) {
	cloneDir := filepath.Join(GetCloneDir(), folder)

	scmType := codeSource.Type
	p, err := GetSCMProvider(scmType)
	if err != nil {
		logdog.Error(err.Error())
		return "", err
	}

	id, err := p.GetCommitID(cloneDir)
	if err != nil {
		return "", err
	}

	return id, nil
}

func GetCommitLog(codeSource *api.CodeSource, folder string) (api.CommitLog, error) {
	cloneDir := filepath.Join(GetCloneDir(), folder)

	scmType := codeSource.Type
	p, err := GetSCMProvider(scmType)
	if err != nil {
		logdog.Error(err.Error())
		return api.CommitLog{}, err
	}

	commitLog := p.GetCommitLog(cloneDir)

	// append repo name
	repoName, err := GetRepoName(codeSource)
	if err != nil {
		return commitLog, err
	}
	commitLog.RepoName = repoName

	// Get commit ID
	commitID, err := GetCommitID(codeSource, folder)
	if err != nil {
		return commitLog, err
	}
	commitLog.ID = commitID

	return commitLog, nil
}

// CloneRepo represents clone main repo and dep repos.
func CloneRepos(token string, codeSources *api.CodeCheckoutStage, ref string) (string, error) {
	var logs string
	// clone main repo
	logs, err := CloneRepo(token, codeSources.MainRepo, ref, "")
	if err != nil {
		return logs, err
	}

	// clone dep repos
	for _, repo := range codeSources.DepRepos {
		log, err := CloneRepo(token, &repo.CodeSource, "", repo.Folder)
		if err != nil {
			return log, err
		}

		// append log
		logs = fmt.Sprintf("%s\n%s", logs, log)
	}

	return logs, nil
}

// CloneRepo represents clone repo to `destPath`, `ref` is for main repo.
func CloneRepo(token string, codeSource *api.CodeSource, ref string, folder string) (string, error) {
	destPath := filepath.Join(GetCloneDir(), folder)

	if err := pathutil.EnsureParentDir(destPath, 0750); err != nil {
		return "", fmt.Errorf("Unable to create parent directory for %s: %v\n", destPath, err)
	}

	scmType := codeSource.Type
	p, err := GetSCMProvider(scmType)
	if err != nil {
		logdog.Error(err.Error())
		return "", err
	}

	token, err = rebuildToken(token, codeSource)
	if err != nil {
		return "", err
	}

	url, err := getURL(codeSource)
	if err != nil {
		return "", err
	}

	var reference string
	if ref != "" {
		// main repo
		reference = ref
	} else {
		// dependent repo
		reference, err = getRef(codeSource)
		if err != nil {
			return "", err
		}
	}

	var logs string
	backoff := initialDelay
	for retries := 0; retries < maxRetries; retries++ {
		logs, err = p.Clone(token, url, reference, destPath)
		if err == nil {
			return logs, nil
		}

		logdog.Infof("failed to checkout, will retry after %d seconds, due to error: %v", backoff, err)
		time.Sleep(backoff)
		backoff *= 2
	}

	return logs, err
}

// rebuildToken rebuilds token username and password if token is empty.
// for private git repository
func rebuildToken(token string, codeSource *api.CodeSource) (string, error) {
	scmType := codeSource.Type

	if scmType == api.Gitlab && token != "" {
		token = "oauth2:" + token
	}

	gitSource, err := api.GetGitSource(codeSource)
	if err != nil {
		logdog.Errorf(err.Error())
		return "", err
	}

	// rebuild token
	if token == "" && gitSource.Username != "" && gitSource.Password != "" {
		token = queryEscape(gitSource.Username, gitSource.Password)
	}

	return token, nil
}

// getURL provide the URL of the code.
func getURL(codeSource *api.CodeSource) (string, error) {
	gitSource, err := api.GetGitSource(codeSource)
	if err != nil {
		logdog.Errorf(err.Error())
		return "", err
	}

	return gitSource.Url, nil
}

// getRef provide the ref(branch or tag) of the code.
func getRef(codeSource *api.CodeSource) (string, error) {
	if codeSource.Type == api.SVN {
		return "", nil
	}

	gitSource, err := api.GetGitSource(codeSource)
	if err != nil {
		logdog.Errorf(err.Error())
		return "", err
	}

	if gitSource.Ref == "" {
		logdog.Warnf("the ref of %s is empty", gitSource.Url)
		return "master", nil
	}
	return gitSource.Ref, nil
}

// queryEscape escapes the string so it can be safely placed
// inside a URL query.
func queryEscape(username, pwdBase64 string) string {
	var pwd string
	pwdB, err := base64.StdEncoding.DecodeString(pwdBase64)
	if err != nil {
		pwd = pwdBase64
	} else {
		pwd = string(pwdB)
	}
	return neturl.QueryEscape(username) + ":" + neturl.QueryEscape(pwd)
}
