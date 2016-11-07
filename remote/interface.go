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
	"github.com/caicloud/cyclone/api"
)

// Remote is the interface of all operations needed for remote repository.
type Remote interface {
	GetTokenQuestURL(string) (string, error)
	Authcallback(code, state string) (string, error)
	GetRepos(string) ([]api.Repo, string, string, error)
	LogOut(userID string) error
	CreateHook(service *api.Service) error
	DeleteHook(service *api.Service) error
	PostCommitStatus(service *api.Service, version *api.Version) error
}
