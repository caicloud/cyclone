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

package scm

import (
	"fmt"
	"testing"

	"github.com/caicloud/cyclone/pkg/api"
)

// testProvider represents SCM provider just for test.
type testProvider struct{}

func (t *testProvider) Clone(token, url, ref, destPath string) (string, error) {
	return "", nil
}

func (t *testProvider) GetCommitID(repoPath string) (string, error) {
	return "", nil
}

func (t *testProvider) GetCommitLog(repoPath string) api.CommitLog {
	return api.CommitLog{}
}

func TestRegisterProvider(t *testing.T) {
	testCases := map[string]struct {
		scmType api.SCMType
		err     error
	}{
		"first register gitlab": {
			api.Gitlab,
			nil,
		},
		"second register gitlab": {
			api.Gitlab,
			fmt.Errorf("SCM provider %s already exists.", api.Gitlab),
		},
		"register github": {
			api.Github,
			fmt.Errorf("SCM provider %s already exists.", api.Github),
		},
		"register svn": {
			api.SVN,
			fmt.Errorf("SCM provider %s already exists.", api.SVN),
		},
	}

	for d, tc := range testCases {
		err := RegisterProvider(tc.scmType, new(testProvider))
		if err != tc.err {
			if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
				t.Errorf("%s failed: expect %v, but got %v", d, tc.err, err)
			}
		}
	}
}

func TestGetSCMProvider(t *testing.T) {
	// Init SCM providers.
	scmProviders = make(map[api.SCMType]SCMProvider)
	if err := RegisterProvider(api.Gitlab, new(testProvider)); err != nil {
		t.Errorf("fail to init SCM providers")
	}

	testCases := map[string]struct {
		scmType api.SCMType
		err     error
	}{
		"get gitlab provider": {
			api.Gitlab,
			nil,
		},
		"get github provider": {
			api.Github,
			fmt.Errorf("unsupported SCM type %s", api.Github),
		},
	}

	for d, tc := range testCases {
		_, err := GetSCMProvider(tc.scmType)
		if err != tc.err {
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("%s failed: expect %v, but got %v", d, err, tc.err)
			}
		}
	}
}

func TestGetRepoNameByURL(t *testing.T) {
	testCases := map[string]struct {
		url  string
		pass string
	}{
		"general gitlab": {
			"https://gitlab.com/jmyue/kubernetes-plugin.git",
			"jmyue/kubernetes-plugin",
		},
		"general github": {
			"https://github.com/caicloud/cyclone.git",
			"caicloud/cyclone",
		},
		"url with ip": {
			"http://192.168.21.100:10080/jmyue/kubernetes-plugin.git",
			"jmyue/kubernetes-plugin",
		},
		"url with localhost": {
			"http://localhost:10080/ci-test/ci-demo.git",
			"ci-test/ci-demo",
		},
	}

	for d, tc := range testCases {
		repoName, err := getRepoNameByURL(tc.url)
		if err != nil {
			t.Error("%s failed as error Expect error to be nil")
		}
		if repoName != tc.pass {
			t.Errorf("%s failed as error : Expect result %s equals to %s", d, repoName, tc.pass)
		}
	}
}

func TestRebuildToken(t *testing.T) {
	type input struct {
		token      string
		codeSource *api.CodeSource
	}

	testCases := map[string]struct {
		input         input
		expectedToken string
	}{
		"gitlab token": {
			input{
				token: "123456",
				codeSource: &api.CodeSource{
					Type: api.Gitlab,
				},
			},
			"oauth2:123456",
		},
		"github token without username and password": {
			input{
				token: "abcdef",
				codeSource: &api.CodeSource{
					Type: api.Github,
				},
			},
			"abcdef",
		},
		"github token with only username": {
			input{
				token: "aaaaa",
				codeSource: &api.CodeSource{
					Type: api.Github,
					Github: &api.GitSource{
						Username: "robin",
					},
				},
			},
			"aaaaa",
		},
		"without github token but with username and password": {
			input{
				token: "",
				codeSource: &api.CodeSource{
					Type: api.Github,
					Github: &api.GitSource{
						Username: "robin",
						Password: "123456",
					},
				},
			},
			"robin:123456",
		},
		"svn token": {
			input{
				token: "$^#$&%*#",
				codeSource: &api.CodeSource{
					Type: api.SVN,
				},
			},
			"$^#$&%*#",
		},
	}

	for d, tc := range testCases {
		result, err := rebuildToken(tc.input.token, tc.input.codeSource)
		if err != nil {
			t.Errorf("Fail to rebuild %s as %v", d, err)
		}

		if result != tc.expectedToken {
			t.Errorf("Fail to rebuild %s: expect %s, but got %s", d, tc.expectedToken, result)
		}
	}

	errorTestCases := map[string]struct {
		input       input
		expectToken string
	}{
		"wrong SCM type": {
			input{
				token: "123456",
				codeSource: &api.CodeSource{
					Type: api.SCMType("Gitkub"),
				},
			},
			"oauth2:123456",
		},
	}

	for d, tc := range errorTestCases {
		_, err := rebuildToken(tc.input.token, tc.input.codeSource)
		if err == nil {
			t.Errorf("Fail to rebuild %s as err %v", d, err)
		}
	}
}

func TestGetURL(t *testing.T) {
	testCases := map[string]struct {
		codeSource  *api.CodeSource
		expectedURL string
	}{
		"empty gitlab": {
			&api.CodeSource{
				Type:   api.Gitlab,
				Gitlab: &api.GitSource{},
			},
			"",
		},
		"gitlab": {
			&api.CodeSource{
				Type: api.Gitlab,
				Gitlab: &api.GitSource{
					Url: "https://gitlab.com",
				},
			},
			"https://gitlab.com",
		},
		"gitlab with github": {
			&api.CodeSource{
				Type: api.Gitlab,
				Gitlab: &api.GitSource{
					Url: "https://gitlab.com",
				},
				Github: &api.GitSource{
					Url: "https://github.com",
				},
			},
			"https://gitlab.com",
		},
		"github": {
			&api.CodeSource{
				Type: api.Github,
				Github: &api.GitSource{
					Url: "https://github.com",
				},
			},
			"https://github.com",
		},
		"svn": {
			&api.CodeSource{
				Type: api.SVN,
				SVN: &api.GitSource{
					Url: "https://svn.com",
				},
			},
			"https://svn.com",
		},
	}

	for d, tc := range testCases {
		result, err := getURL(tc.codeSource)
		if err != nil {
			t.Errorf("Fail to get URL from %s as %v", d, err)
		}

		if result != tc.expectedURL {
			t.Errorf("Fail to get URL from %s: expect %s, but got %s", d, tc.expectedURL, result)
		}
	}
}

func TestGetRef(t *testing.T) {
	testCases := map[string]struct {
		codeSource  *api.CodeSource
		expectedURL string
	}{
		"empty gitlab": {
			&api.CodeSource{
				Type:   api.Gitlab,
				Gitlab: &api.GitSource{},
			},
			"master",
		},
		"gitlab": {
			&api.CodeSource{
				Type: api.Gitlab,
				Gitlab: &api.GitSource{
					Ref: "dev",
					Url: "https://gitlab.com",
				},
			},
			"dev",
		},
		"gitlab with github": {
			&api.CodeSource{
				Type: api.Gitlab,
				Gitlab: &api.GitSource{
					Ref: "dev",
					Url: "https://gitlab.com",
				},
				Github: &api.GitSource{
					Ref: "master",
					Url: "https://github.com",
				},
			},
			"dev",
		},
		"github": {
			&api.CodeSource{
				Type: api.Github,
				Github: &api.GitSource{
					Ref: "pre-release",
					Url: "https://github.com",
				},
			},
			"pre-release",
		},
		"svn": {
			&api.CodeSource{
				Type: api.SVN,
				SVN: &api.GitSource{
					Ref: "master",
					Url: "https://svn.com",
				},
			},
			"",
		},
	}

	for d, tc := range testCases {
		result, err := getRef(tc.codeSource)
		if err != nil {
			t.Errorf("Fail to get URL from %s as %v", d, err)
		}

		if result != tc.expectedURL {
			t.Errorf("Fail to get URL from %s: expect %s, but got %s", d, tc.expectedURL, result)
		}
	}
}
