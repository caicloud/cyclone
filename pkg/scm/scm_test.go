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

import (
	"fmt"
	"testing"

	"github.com/caicloud/cyclone/pkg/api"
)

func testProviderFunc(*api.SCMConfig) (SCMProvider, error) {
	return nil, nil
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
		err := RegisterProvider(tc.scmType, testProviderFunc)
		if err != tc.err {
			if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
				t.Errorf("%s failed: expect %v, but got %v", d, tc.err, err)
			}
		}
	}
}

func TestGetSCMProvider(t *testing.T) {
	// Init SCM providers.
	scmProviders = make(map[api.SCMType]newSCMProviderFunc)
	if err := RegisterProvider(api.Gitlab, testProviderFunc); err != nil {
		t.Errorf("fail to init SCM providers")
	}

	testCases := map[string]struct {
		scmCfg *api.SCMConfig
		err    error
	}{
		"scm should not be nil": {
			nil,
			fmt.Errorf("SCM config is nil"),
		},
		"get gitlab provider": {
			&api.SCMConfig{
				Type: api.Gitlab,
			},
			nil,
		},
		"get github provider": {
			&api.SCMConfig{
				Type: api.Github,
			},
			fmt.Errorf("unsupported SCM type %s", api.Github),
		},
	}

	for d, tc := range testCases {
		_, err := GetSCMProvider(tc.scmCfg)
		if err != tc.err {
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("%s failed: expect %v, but got %v", d, err, tc.err)
			}
		}
	}
}
