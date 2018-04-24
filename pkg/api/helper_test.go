/*
Copyright 2018 caicloud authors. All rights reserved.

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

package api

import (
	"testing"
)

func TestGetGitSource(t *testing.T) {
	source := &GitSource{}
	testCases := map[string]struct {
		codesource *CodeSource
		source     *GitSource
	}{
		"gitlab": {
			&CodeSource{
				Type:   Gitlab,
				Gitlab: source,
			},
			source,
		},
		"github": {
			&CodeSource{
				Type:   Github,
				Gitlab: source,
			},
			source,
		},
		"svn": {
			&CodeSource{
				Type:   SVN,
				Gitlab: source,
			},
			source,
		},
	}

	for _, tc := range testCases {
		_, err := GetGitSource(tc.codesource)
		if err != nil {
			t.Error("%s failed as error Expect error to be nil")
		}
	}
}

func TestGetGitSourceError(t *testing.T) {
	source := &GitSource{}
	testCases := map[string]struct {
		codesource *CodeSource
		error      string
	}{
		"gitlabb": {
			&CodeSource{
				Type:   "gitlabb",
				Gitlab: source,
			},
			"SCM type gitlabb is not supported",
		},
	}

	for _, tc := range testCases {
		_, err := GetGitSource(tc.codesource)
		if err.Error() != tc.error {
			t.Error("%s failed as error Expect error to be nil")
		}
	}
}
