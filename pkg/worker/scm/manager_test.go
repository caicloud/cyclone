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
	"testing"
)

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
