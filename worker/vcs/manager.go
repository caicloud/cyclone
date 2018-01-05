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

package vcs

import (
	"encoding/base64"
	"fmt"
	neturl "net/url"
	"strings"

	"github.com/caicloud/cyclone/api"
	newAPI "github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/pathutil"
	"github.com/caicloud/cyclone/store"
	steplog "github.com/caicloud/cyclone/worker/log"
	"github.com/caicloud/cyclone/worker/vcs/provider"
)

const (
	// The dir which the repo clone to.
	CLONE_DIR = "/root/code"
)

// Manager manages all version control operations, like clone, cherry-pick.
// Based on the operations, some are handled asychronously and some are not.
// Asynchronous operations are time consuming and usually involve stream output
// to clients, like clone, fetch, etc; synchronous operations are not time
// consuming and usually don't have to send output, like checkout a tag, etc.
// The above constants define all async operations; all other operations are
// synchronous. Manager is also responsible for managing repository status;
// it knows whether a repository is healthy or not, and set repository status
// accordingly.
// synchronous.
type Manager struct {
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{}
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

// getAuthURL rebuilds url with auth token or username and password
// for private git repository
func getAuthURL(job *store.Job) string {

	url := job.Pipeline.Repository.URL

	var token string
	if t, ok := job.PipelineRecord.Data["Token"]; ok {
		token = t.(string)
	}
	username := job.Pipeline.Repository.Username
	pwd := job.Pipeline.Repository.Password

	// rebuild token
	if token != "" {
		if job.Pipeline.Repository.SubVcs == api.GITLAB {
			token = "oauth2:" + token
		}
	} else if username != "" && pwd != "" {
		token = queryEscape(username, pwd)
	}

	// insert token
	if token != "" && job.Pipeline.Repository.Vcs == api.Git {
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

	return url
}

// CloneVersionRepository clones a pipelineRecord's repo.
func (vm *Manager) CloneVersionRepository(job *store.Job) error {
	// Get the path to store cloned repository.
	destPath := vm.GetCloneDir()
	if err := pathutil.EnsureParentDir(destPath, 0750); err != nil {
		return fmt.Errorf("Unable to create parent directory for %s: %v\n", destPath, err)
	}

	// Find version control system worker and return if error occurs.
	worker, err := vm.findVcsForPipeline(job.Pipeline)
	if err != nil {
		return fmt.Errorf("Unable to write to job output for job: %v\n", err)
	}

	url := getAuthURL(job)

	if err := worker.Clone(url, destPath, job); err != nil {
		return fmt.Errorf("Unable to clone repository for version: %v\n", err)
	}

	// create version call by UI API, the commit is empty
	// create version call by webhook, the commit is not empty
	if "" == job.PipelineRecord.Commit {
		//checkout branch/tag
		if err := worker.CheckoutTag(destPath, job.PipelineRecord.Ref); err != nil {
			return fmt.Errorf("Unable to checkout branch/tag %s : %v\n", event.Version.Ref, err)
		}

		// set version commit
		if commit, err := worker.GetTagCommit(destPath, job.PipelineRecord.Ref); err != nil {
			log.Error("cannot get tag commit")
		} else {
			// write to DB in posthook
			job.PipelineRecord.Commit = commit
		}
	} else {
		// checkout special commit
		if err = worker.CheckOutByCommitID(job.PipelineRecord.Commit, destPath, job); err != nil {
			job.Pipeline.Repository.Status = api.RepositoryMissing
			return fmt.Errorf("Unable to check out commit %s :%v\n", job.PipelineRecord.Commit, err)
		}
	}

	if api.APIOperator == job.PipelineRecord.Operator {
		// create tag
		if err := worker.NewTagFromLatest(destPath, job); err != nil {
			log.Errorf("Unable to push new commit %s :%v\n", job.PipelineRecord.Commit, err)
		}
	}
	return nil
}

// GetCloneDir returns the directory where a repository should be cloned to.
func (vm *Manager) GetCloneDir() string {
	return CLONE_DIR
}

// findVcsForPipeline is a helper method which finds the VCS worker based on pipeline spec.
func (vm *Manager) findVcsForPipeline(pipeline *newAPI.Pipeline) (VCS, error) {
	switch pipeline.Repository.Vcs {
	case api.Git:
		return &provider.Git{}, nil
	case api.Svn:
		return &provider.Svn{}, nil
	case api.Fake:
		return provider.NewFake(pipeline.Repository.URL)
	default:
		return nil, fmt.Errorf("Unknown version control system %s", pipeline.Repository.Vcs)
	}
}
