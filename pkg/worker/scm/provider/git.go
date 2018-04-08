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
	"path"
	"strconv"
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
	if err := scm.RegisterProvider(api.Github, new(Git)); err != nil {
		log.Error(err)
	}

	if err := scm.RegisterProvider(api.Gitlab, new(Git)); err != nil {
		log.Error(err)
	}
}

func (g *Git) rebuildURL(token, url string) (string, error) {
	// insert token
	if token != "" {
		position := -1
		if strings.HasPrefix(url, "http://") {
			position = len("http://")
		} else if strings.HasPrefix(url, "https://") {
			position = len("https://")
		}
		if position > 0 {
			url = insert(url, token+"@", position)
		}
	}

	return url, nil
}

// This function is used to insert the string "insertion" into the "url"
// at the "index" postiion
func insert(url, insertion string, index int) string {
	result := make([]byte, len(url)+len(insertion))
	slice := []byte(url)
	at := copy(result, slice[:index])
	at += copy(result[at:], insertion)
	copy(result[at:], slice[index:])
	return string(result)
}

func (g *Git) Clone(token, url, ref, destPath string) (string, error) {
	log.Info("About to clone git repository.", log.Fields{"url": url, "destPath": destPath})
	url, err := g.rebuildURL(token, url)
	if err != nil {
		return "", err
	}

	base := path.Base(destPath)
	dir := path.Dir(destPath)
	type cmd struct {
		dir  string
		args []string
	}
	cmds := []cmd{
		cmd{dir, []string{"clone", url, base}},
		cmd{destPath, []string{"fetch", "origin", ref}},
		cmd{destPath, []string{"checkout", "-qf", "FETCH_HEAD"}},
	}

	var outputs string
	for _, cmd := range cmds {
		output, err := executil.RunInDir(cmd.dir, "git", cmd.args...)
		if err != nil {
			log.Error("Error when clone", log.Fields{"command": cmd, "error": err})
			return "", err
		}

		outputs = outputs + string(output)
	}

	log.Info("Successfully cloned git repository.", log.Fields{"url": url, "destPath": destPath})

	return outputs, nil
}

// GetCommitID implements VCS interface.
func (g *Git) GetCommitID(repoPath string) (string, error) {
	args := []string{"log", "-n", "1", `--pretty=format:"%H"`}
	output, err := executil.RunInDir(repoPath, "git", args...)

	return strings.Trim(strings.Trim(string(output), "\n"), "\""), err
}

func (g *Git) getAuthor(repoPath string) (string, error) {
	args := []string{"log", "-n", "1", `--pretty=format:"%an"`}
	output, err := executil.RunInDir(repoPath, "git", args...)

	return strings.Trim(strings.Trim(string(output), "\n"), "\""), err
}

func (g *Git) getDate(repoPath string) (time.Time, error) {
	args := []string{"log", "-n", "1", `--pretty=format:"%ad"`, `--date=raw`}
	output, err := executil.RunInDir(repoPath, "git", args...)
	if err != nil {
		return time.Time{}, err
	}

	// timeRaw  --> "timestamp timezone"  eg:"1520239308 +0800"
	timeRaw := strings.TrimSpace(strings.Trim(strings.Trim(string(output), "\n"), "\""))

	ts := strings.Split(timeRaw, " ")
	if len(ts) < 1 {
		return time.Time{}, fmt.Errorf("split time raw  %s fail", timeRaw)
	}

	timestamp, err := strconv.ParseInt(ts[0], 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
}

func (g *Git) getMessage(repoPath string) (string, error) {
	args := []string{"log", "-n", "1", `--pretty=format:"%s"`}
	output, err := executil.RunInDir(repoPath, "git", args...)

	return strings.Trim(strings.Trim(string(output), "\n"), "\""), err
}

// GetTagAuthor implements VCS interface.
func (g *Git) GetCommitLog(repoPath string) api.CommitLog {
	commitLog := api.CommitLog{}

	author, erra := g.getAuthor(repoPath)
	if erra != nil {
		log.Warningf("get tag author fail %s", erra.Error())
	}

	commitLog.Author = author
	date, errd := g.getDate(repoPath)
	if errd != nil {
		log.Warningf("get tag date fail %s", errd.Error())
	}

	commitLog.Date = date
	message, errm := g.getMessage(repoPath)
	if errm != nil {
		log.Warningf("get tag message fail %s", errm.Error())
	}

	commitLog.Message = message
	return commitLog
}
