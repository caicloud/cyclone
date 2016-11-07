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
	"errors"
	"os"
	"os/exec"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
)

// Fake is the type for fake(local file system) vcs provider.
type Fake struct {
	fakeRepoPath string
}

// NewFake returns a fake implementation of vcs.
func NewFake(fakeRepoPath string) (*Fake, error) {
	dir, err := os.Stat(fakeRepoPath)
	if err != nil {
		return nil, err
	}
	if !dir.IsDir() {
		return nil, errors.New("fakeRepoPath must be a directory")
	}
	return &Fake{
		fakeRepoPath: fakeRepoPath,
	}, nil
}

// CloneRepo clones the repo to local file system.
func (f *Fake) CloneRepo(url, destPath string, event *api.Event) error {
	log.InfoWithFields("About to clone fake repository.", log.Fields{"url": url, "destPath": destPath})

	cmd := exec.Command("cp", "-r", f.fakeRepoPath, destPath)
	err := cmd.Run()
	if err != nil {
		log.Errorf("Error when clone: %v", err)
	} else {
		log.InfoWithFields("Successfully cloned fake repository.", log.Fields{"url": url, "destPath": destPath})
	}
	return err
}

// NewTagFromLatest creates a new tag from latest source for a service.
func (f *Fake) NewTagFromLatest(repoPath string, event *api.Event) error {
	log.InfoWithFields("Successfully created a new fake tag.", log.Fields{"repoPath": repoPath})
	return nil
}

// CheckoutTag checkout to given tag.
func (f *Fake) CheckoutTag(repoPath string, tag string) error {
	log.InfoWithFields("Successfully checked out to fake tag.", log.Fields{"repoPath": repoPath, "tag": tag})
	return nil
}

// GetTagCommit finds commit/revision hash of a given tag.
func (f *Fake) GetTagCommit(repoPath string, tag string) (string, error) {
	return "hash" + tag + "hash", nil
}

// CheckOutByCommitID check out code in repo by special commit id.
func (f *Fake) CheckOutByCommitID(commitID string, repoPath string, event *api.Event) error {
	return nil
}
