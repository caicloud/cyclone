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
	"path"
	"strings"
	"time"

	log "github.com/zoumo/logdog"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/executil"
	"github.com/caicloud/cyclone/pkg/worker/scm"
)

// Git is the type for git provider.
type Git struct{}

func init() {
	if err := scm.RegisterProvider(api.GitHub, new(Git)); err != nil {
		log.Error(err)
	}

	if err := scm.RegisterProvider(api.GitLab, new(Git)); err != nil {
		log.Error(err)
	}
}

func (g *Git) Clone(url, ref, destPath string) (string, error) {
	log.Info("About to clone git repository.", log.Fields{"url": url, "destPath": destPath})

	base := path.Base(destPath)
	dir := path.Dir(destPath)
	args := []string{"clone", "-b", ref, url, base}

	output, err := executil.RunInDir(dir, "git", args...)

	if err != nil {
		log.Error("Error when clone", log.Fields{"error": err})
	} else {
		log.Info("Successfully cloned git repository.", log.Fields{"url": url, "destPath": destPath})
	}
	return string(output), err
}

// GetTagCommit implements VCS interface.
func (g *Git) GetTagCommit(repoPath string, tag string) (string, error) {
	args := []string{"rev-list", "-n", "1", tag}
	output, err := executil.RunInDir(repoPath, "git", args...)

	return strings.Trim(string(output), "\n"), err
}

func (g *Git) getTagAuthor(repoPath string, tag string) (string, error) {
	args := []string{"log", "-n", "1", tag, `--pretty=format:"%an"`}
	output, err := executil.RunInDir(repoPath, "git", args...)

	return strings.Trim(strings.Trim(string(output), "\n"), "\""), err
}

func (g *Git) getTagDate(repoPath string, tag string) (time.Time, error) {
	args := []string{"log", "-n", "1", tag, `--pretty=format:"%ad"`, `--date=rfc`}
	output, err := executil.RunInDir(repoPath, "git", args...)
	if err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse(time.RFC1123Z, strings.Trim(strings.Trim(string(output), "\n"), "\""))
	if err != nil {
		return time.Time{}, err
	}

	return t.Local(), err
}

func (g *Git) getTagMessage(repoPath string, tag string) (string, error) {
	args := []string{"log", "-n", "1", tag, `--pretty=format:"%s"`}
	output, err := executil.RunInDir(repoPath, "git", args...)

	return strings.Trim(strings.Trim(string(output), "\n"), "\""), err
}

// GetTagAuthor implements VCS interface.
func (g *Git) GetTagCommitLog(repoPath string, tag string) api.CommitLog {
	commitLog := api.CommitLog{}

	author, erra := g.getTagAuthor(repoPath, tag)
	if erra != nil {
		log.Warningf("get tag author fail %s", erra.Error())
	}

	commitLog.Author = author
	date, errd := g.getTagDate(repoPath, tag)
	if errd != nil {
		log.Warningf("get tag date fail %s", errd.Error())
	}

	commitLog.Date = date
	message, errm := g.getTagMessage(repoPath, tag)
	if errm != nil {
		log.Warningf("get tag message fail %s", errm.Error())
	}

	commitLog.Message = message
	return commitLog
}
