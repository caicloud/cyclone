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
	"github.com/caicloud/cyclone/api"
)

// VCS is the interface of all operations needed for managing repository.
type VCS interface {
	// CloneRepo pulls repository from url into destination path. Code base lives
	// directly under 'destPath', do not nest the code base. E.g.
	//   CloneRepo("https://github.com/caicloud/cyclone", /tmp/xxx/yyy")
	// should clone cyclone directly under /tmp/xxx/yyy, not /tmp/xxx/yyy/cyclone/
	// You can always assume 'destPath' doesn't exist, but its parent direcotry
	// exists.
	CloneRepo(url, destPath string, event *api.Event) error

	// TODO: Those methods are not used, clean them up when we have better design.

	// NewTagFromLatest creates a new tag from latest source for a service.
	NewTagFromLatest(repoPath string, event *api.Event) error

	// CheckoutTag checkout to given tag.
	CheckoutTag(repoPath string, tag string) error

	// GetTagCommit finds commit/revision hash of a given tag.
	GetTagCommit(repoPath string, tag string) (string, error)

	// CheckOutByCommitID checks out code in repo by special commit id.
	CheckOutByCommitID(commitID string, repoPath string, event *api.Event) error
}
