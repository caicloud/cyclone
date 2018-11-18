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
			t.Errorf("failed as %v Expect error to be nil", err)
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
			t.Errorf("failed as %v Expect error to be nil", err)
		}
	}
}

func TestGetURL(t *testing.T) {
	testCases := map[string]struct {
		codeSource  *CodeSource
		expectedURL string
	}{
		"empty gitlab": {
			&CodeSource{
				Type:   Gitlab,
				Gitlab: &GitSource{},
			},
			"",
		},
		"gitlab": {
			&CodeSource{
				Type: Gitlab,
				Gitlab: &GitSource{
					Url: "https://gitlab.com",
				},
			},
			"https://gitlab.com",
		},
		"gitlab with github": {
			&CodeSource{
				Type: Gitlab,
				Gitlab: &GitSource{
					Url: "https://gitlab.com",
				},
				Github: &GitSource{
					Url: "https://github.com",
				},
			},
			"https://gitlab.com",
		},
		"github": {
			&CodeSource{
				Type: Github,
				Github: &GitSource{
					Url: "https://github.com",
				},
			},
			"https://github.com",
		},
		"svn": {
			&CodeSource{
				Type: SVN,
				SVN: &GitSource{
					Url: "https://svn.com",
				},
			},
			"https://svn.com",
		},
	}

	for d, tc := range testCases {
		result, err := GetURL(tc.codeSource)
		if err != nil {
			t.Errorf("Fail to get URL from %s as %v", d, err)
		}

		if result != tc.expectedURL {
			t.Errorf("Fail to get URL from %s: expect %s, but got %s", d, tc.expectedURL, result)
		}
	}
}
