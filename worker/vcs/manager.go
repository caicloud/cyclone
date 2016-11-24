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
	"fmt"
	"strings"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/pathutil"
	steplog "github.com/caicloud/cyclone/worker/log"
	"github.com/caicloud/cyclone/worker/vcs/provider"
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

func getUrlwithToken(url, subvcs, token string) string {
	switch subvcs {
	case api.GITHUB:
		if token == "" {
			log.Warn("Sub VCS is GitHub, but the token is empty, maybe there is an error")
			break
		}
		position := strings.Index(url, "github.com")
		if position == -1 {
			log.Warn("subvcs is github and the token is not empty, but the url dosen't contain 'github.com'")
			break
		}
		url = insert(url, token+"@", position)
	case api.GITLAB:
		if token == "" {
			log.Warn("Sub VCS is GitLab, but the token is empty, maybe there is an error")
			break
		}
		if strings.HasPrefix(url, "http://") {
			position := len("http://")
			url = insert(url, "oauth2:"+token+"@", position)
		} else if strings.HasPrefix(url, "https://") {
			position := len("https://")
			url = insert(url, "oauth2:"+token+"@", position)
		}
	default:
	}
	return url
}

// CloneServiceRepository clones a service's repo and records service status accordingly.
func (vm *Manager) CloneServiceRepository(event *api.Event) error {
	// Get the path to store cloned repository.
	destPath := vm.GetCloneDir(&event.Service, &event.Version)
	if err := pathutil.EnsureParentDir(destPath, 0750); err != nil {
		event.Service.Repository.Status = api.RepositoryInternalError
		return fmt.Errorf("Unable to create parent directory for %s: %v\n", destPath, err)
	}

	// Find version control system worker and return if error occurs.
	worker, err := vm.findVcsForService(&event.Service)
	if err != nil {
		event.Service.Repository.Status = api.RepositoryUnknownVcs
		return fmt.Errorf("Unable to write to event output for event: %v\n", err)
	}

	url := getUrlwithToken(event.Service.Repository.URL, event.Service.Repository.SubVcs, event.Data["Token"].(string))
	if err := worker.CloneRepo(url, destPath, event); err != nil {
		event.Service.Repository.Status = api.RepositoryMissing
		return fmt.Errorf("Unable to clone repository for service: %v\n", err)
	}

	// Happy path - update status to healthy and return nil error. Database status
	// will be updated via defer function. If we encounter error during database
	// update, repository status will be set to internal error.
	event.Service.Repository.Status = api.RepositoryHealthy
	return nil
}

// CloneVersionRepository clones a version's repo
func (vm *Manager) CloneVersionRepository(event *api.Event) error {
	// Get the path to store cloned repository.
	destPath := vm.GetCloneDir(&event.Service, &event.Version)
	if err := pathutil.EnsureParentDir(destPath, 0750); err != nil {
		return fmt.Errorf("Unable to create parent directory for %s: %v\n", destPath, err)
	}

	// Find version control system worker and return if error occurs.
	worker, err := vm.findVcsForService(&event.Service)
	if err != nil {
		return fmt.Errorf("Unable to write to event output for event: %v\n", err)
	}

	url := getUrlwithToken(event.Version.URL, event.Service.Repository.SubVcs, event.Data["Token"].(string))
	steplog.InsertStepLog(event, steplog.CloneRepository, steplog.Start, nil)
	if err := worker.CloneRepo(url, destPath, event); err != nil {
		steplog.InsertStepLog(event, steplog.CloneRepository, steplog.Stop, err)
		return fmt.Errorf("Unable to clone repository for version: %v\n", err)
	}
	// create version call by UI API, the commit is empty
	// create version call by webhook, the commit is not empty
	if "" == event.Version.Commit {
		// set version commit
		if commit, err := worker.GetTagCommit(destPath, "master"); err != nil {
			log.Error("cannot get tag commit")
		} else {
			// write to DB in posthook
			event.Version.Commit = commit
		}
	} else {
		// checkout special commit
		if err = worker.CheckOutByCommitID(event.Version.Commit, destPath, event); err != nil {
			event.Service.Repository.Status = api.RepositoryMissing
			steplog.InsertStepLog(event, steplog.CloneRepository, steplog.Stop, err)
			return fmt.Errorf("Unable to check out commit %s :%v\n", event.Version.Commit, err)
		}
	}

	if api.APIOperator == event.Version.Operator {
		// create tag
		if err := worker.NewTagFromLatest(destPath, event); err != nil {
			log.Errorf("Unable to push new commit %s :%v\n", event.Version.Commit, err)
		}
	}
	steplog.InsertStepLog(event, steplog.CloneRepository, steplog.Finish, nil)
	return nil
}

// NewTagFromLatest creates a new tag from latest source for a service.
func (vm *Manager) NewTagFromLatest(event *api.Event) error {
	// Find version control system worker and return if error occurs.
	worker, err := vm.findVcsForService(&event.Service)
	if err != nil {
		return fmt.Errorf("Unable to checkout latest source %#+v: %v", event.Service, err)
	}

	// Do the actual work.
	repositoryPath := vm.GetCloneDir(&event.Service, &event.Version)
	err = worker.NewTagFromLatest(repositoryPath, event)
	if err != nil {
		return fmt.Errorf("Unable to create tag for service %#+v: %v\n", event.Service, err)
	}
	return nil
}

// CheckoutTag checkout to given tag in version.
func (vm *Manager) CheckoutTag(service *api.Service, version *api.Version) error {
	// Find version control system worker and return if error occurs.
	worker, err := vm.findVcsForService(service)
	if err != nil {
		return err
	}

	err = worker.CheckoutTag(vm.GetCloneDir(service, version), version.Name)
	if err != nil {
		return fmt.Errorf("Unable to checkout tag for service %#+v: %v\n", service, err)
	}
	return nil
}

// GetTagCommit finds commit/revision hash of a given tag.
func (vm *Manager) GetTagCommit(service *api.Service, version *api.Version) (string, error) {
	// Find version control system worker and return if error occurs.
	worker, err := vm.findVcsForService(service)
	if err != nil {
		return "", err
	}

	commit, err := worker.GetTagCommit(vm.GetCloneDir(service, version), version.Name)
	if err != nil {
		return "", err
	}
	return commit, nil
}

// GetCloneDir returns the directory where a repository should be cloned to. It honors
// flag 'cloneDir' which is primarily used for testing/debugging. We use the combination
// of cloneDir, userID, serviceID, and versionID as repository path. As repositories come
// and go, we don't have to have a persistent directory structure; therefore, if version
// is not nil, use cloneDir/userID/versionID; otherwise, use cloneDir/userID/serviceID.
func (vm *Manager) GetCloneDir(service *api.Service, version *api.Version) string {
	if version.VersionID == "" {
		return fmt.Sprintf("/%s/%s", service.UserID, service.ServiceID)
	}
	return fmt.Sprintf("/%s/%s", service.UserID, version.VersionID)
}

// findVcsForService is a helper method which finds the VCS worker based on service spec.
func (vm *Manager) findVcsForService(service *api.Service) (VCS, error) {
	switch service.Repository.Vcs {
	case api.Git:
		return &provider.Git{}, nil
	case api.Svn:
		return &provider.Svn{}, nil
	case api.Fake:
		return provider.NewFake(service.Repository.URL)
	default:
		return nil, fmt.Errorf("Unknown version control system %s", service.Repository.Vcs)
	}
}
