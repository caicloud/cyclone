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

package provider

import "strings"

const (
	// ListPerPageOpt represents the value for PerPage in list options.
	ListPerPageOpt = 30
)

// ParseServerURL is a helper func to parse the server url, such as https://github.com/ to return server(github.com).
func ParseServerURL(url string) string {
	strs := strings.SplitN(strings.TrimSuffix(url, "/"), "://", -1)
	return strs[1]
}

// ParseRepoURL is a helper func to parse the repo url, such as https://github.com/caicloud/test.git
// to return owner(caicloud) and name(test).
func ParseRepoURL(url string) (string, string) {
	strs := strings.SplitN(url, "/", -1)
	name := strings.SplitN(strs[4], ".", -1)
	return strs[3], name[0]
}
