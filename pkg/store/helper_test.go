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

package store

import (
	"testing"

	"github.com/caicloud/cyclone/pkg/api"
)

var (
	project        = &api.Project{}
	defaultSaltKey = "1234567812345678"
)

func TestEncryptPasswordsForProjects(t *testing.T) {
	testCases := map[string]struct {
		scm      *api.SCMConfig
		registry *api.Registry
		saltKey  string
	}{
		"Github token": {
			scm: &api.SCMConfig{
				Type:  api.Github,
				Token: "dakjfajefpoaiesjfoajefasjfasfeoija",
			},
		},
		"Gitlab token": {
			scm: &api.SCMConfig{
				Type:  api.Gitlab,
				Token: "fsdrgraeofasfjaefoawefasfjasfaweojf",
			},
		},
		"empty token": {
			scm: &api.SCMConfig{
				Type:  api.Github,
				Token: "",
			},
		},
		"registry": {
			registry: &api.Registry{
				Password: "password",
			},
		},
		"SCM and registry": {
			scm: &api.SCMConfig{
				Type:  api.Github,
				Token: "dakjfajefpoaiesjfoajefasjfasfeoija",
			},
			registry: &api.Registry{
				Password: "password",
			},
		},
		"salt key length 24": {
			registry: &api.Registry{
				Password: "password",
			},
			saltKey: "123456781234567812345678",
		},
	}

	for d, tc := range testCases {
		project.SCM = tc.scm
		if tc.saltKey == "" {
			tc.saltKey = defaultSaltKey
		}
		if err := encryptPasswordsForProjects(project, tc.saltKey); err != nil {
			t.Fatalf("%s: fail to encrypt project: %v", d, err)
		}
	}

	errTestCases := map[string]struct {
		scm      *api.SCMConfig
		registry *api.Registry
		saltKey  string
	}{
		"too short salt key": {
			scm: &api.SCMConfig{
				Type:  api.Github,
				Token: "dakjfajefpoaiesjfoajefasjfasfeoija",
			},
			saltKey: "1234567",
		},
	}

	for d, tc := range errTestCases {
		project.SCM = tc.scm
		project.Registry = tc.registry
		if err := encryptPasswordsForProjects(project, tc.saltKey); err == nil {
			t.Fatalf("%s: expect error but got nil", d)
		}
	}
}

func TestDecryptPasswordsForProjects(t *testing.T) {
	type input struct {
		scm      *api.SCMConfig
		registry *api.Registry
	}
	type expected struct {
		scmToken         string
		registryPassword string
	}
	testCases := map[string]struct {
		input    input
		expected expected
	}{
		"scm token": {
			input: input{
				scm: &api.SCMConfig{
					Type:  api.Github,
					Token: "quJH1mp13LNGD1JS/ABmka1A62F4hjHq",
				},
			},
			expected: expected{
				scmToken: "password",
			},
		},
		"registry password": {
			input: input{
				registry: &api.Registry{
					Password: "8EcVzXDz0LPvOKWNgfi2Sf33qbG2rQ==",
				},
			},
			expected: expected{
				registryPassword: "123456",
			},
		},
		"scm token and registry password": {
			input: input{
				scm: &api.SCMConfig{
					Type:  api.Github,
					Token: "quJH1mp13LNGD1JS/ABmka1A62F4hjHq",
				},
				registry: &api.Registry{
					Password: "8EcVzXDz0LPvOKWNgfi2Sf33qbG2rQ==",
				},
			},
			expected: expected{
				scmToken:         "password",
				registryPassword: "123456",
			},
		},
	}

	for d, tc := range testCases {
		project.SCM = tc.input.scm
		project.Registry = tc.input.registry
		if err := decryptPasswordsForProjects(project, defaultSaltKey); err != nil {
			t.Fatalf("%s: fail to encrypt project: %v", d, err)
		}

		if project.SCM != nil && tc.expected.scmToken != "" && project.SCM.Token != tc.expected.scmToken {
			t.Fatalf("%s: decrypted token expeced %s, but got %s", d, tc.expected.scmToken, project.SCM.Token)
		}

		if project.Registry != nil && tc.expected.registryPassword != "" && project.Registry.Password != tc.expected.registryPassword {
			t.Fatalf("%s: decrypted token expeced %s, but got %s", d, tc.expected.registryPassword, project.Registry.Password)
		}
	}
}
