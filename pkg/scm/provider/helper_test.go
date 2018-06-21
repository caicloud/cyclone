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

import (
	"testing"
)

func TestParseServerURL(t *testing.T) {
	testCases := map[string]struct {
		url    string
		server string
	}{
		"public gitlab": {
			"https://gitlab.com",
			"gitlab.com",
		},
		"public github": {
			"https://github.com",
			"github.com",
		},
		"private gitlab": {
			"http://gitlab.caicloud.com",
			"gitlab.caicloud.com",
		},
		"server with port": {
			"https://gitlab.caicloud.com:8080",
			"gitlab.caicloud.com:8080",
		},
		"server with ip": {
			"https://192.168.21.110:8080",
			"192.168.21.110:8080",
		},
	}

	for d, tc := range testCases {
		result := ParseServerURL(tc.url)
		if result != tc.server {
			t.Errorf("fail to parse %s: expect %s, but got %s", d, tc.server, result)
		}
	}
}

func TestParseRepoURL(t *testing.T) {
	type expect struct {
		owner string
		name  string
	}

	testCases := map[string]struct {
		url    string
		expect expect
	}{
		"public gitlab repo": {
			"https://gitlab.com/123/test",
			expect{
				"123",
				"test",
			},
		},
		"public github repo": {
			"https://github.com/caicloud/cyclone.git",
			expect{
				"caicloud",
				"cyclone",
			},
		},
		"server with port": {
			"http://gitlab.caicloud.com:8080/123abc/456.git",
			expect{
				"123abc",
				"456",
			},
		},
		"server with ip": {
			"https://192.168.21.110:8080/hello-world/test",
			expect{
				"hello-world",
				"test",
			},
		},
	}

	for d, tc := range testCases {
		owner, name := ParseRepoURL(tc.url)
		if owner != tc.expect.owner || name != tc.expect.name {
			t.Errorf("fail to parse %s: expect {%s %s}, but got %v", d, owner, name, tc.expect)
		}
	}
}
