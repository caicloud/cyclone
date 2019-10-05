/*
Copyright 2017 caicloud authors. All rights reserved.
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

import "strings"

const (
	// ListOptPerPage represents the value for PerPage in list options.
	// Max is 100 for both GitHub and Gitlab, refer to https://developer.github.com/v3/guides/traversing-with-pagination/#basics-of-pagination
	ListOptPerPage = 100
)

// ParseServerURL is a helper func to parse the server url, such as https://github.com/ to return server(github.com).
func ParseServerURL(url string) string {
	strs := strings.SplitN(strings.TrimSuffix(url, "/"), "://", -1)
	return strs[1]
}

// ParseRepo parses owner and name from full repo name.
// For example, parse caicloud/cyclone will return owner(caicloud) and name(cyclone).
func ParseRepo(url string) (string, string) {
	strs := strings.SplitN(url, "/", -1)
	if len(strs) < 2 {
		return strs[0], ""
	}

	return strs[0], strs[1]
}

// Repository represents the information of a repository.
type Repository struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// IsDockerfile judges whether the file is Dockerfile. Dockerfile should meet requirements:
// * File should not be in dep folders.
// * File name should has Dockerfile prefix.
func IsDockerfile(path string) bool {
	if IsInDep(path) {
		return false
	}

	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	return strings.HasPrefix(name, "Dockerfile")
}

// IsInDep judges whether the path is in dep folders.
func IsInDep(path string) bool {
	// Will exclude more dep folders for different languages if necessary.
	// * Golang: vendor
	depFolders := []string{"vendor/"}

	for _, d := range depFolders {
		if strings.HasPrefix(path, d) {
			return true
		}
	}

	return false
}
