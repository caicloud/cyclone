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

package remote

import (
	"fmt"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/remote/provider"
)

// Manager is the type for remote manager.
type Manager struct {
}

// NewManager creates a new Manager.
func NewManager() *Manager {
	return &Manager{}
}

// FindRemote returns the remote by codereposity.
func (remote *Manager) FindRemote(codereposity string) (Remote, error) {
	switch codereposity {
	case api.GITHUB:
		return &provider.GitHub{}, nil
	case api.GITLAB:
		return &provider.GitLab{}, nil
	default:
		return nil, fmt.Errorf("Unknown remote version control system %s", codereposity)
	}
}
