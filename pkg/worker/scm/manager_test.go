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

	url := `https://gitlab.com/jmyue/kubernetes-plugin.git`
	repoName, err := getRepoNameByURL(url)
	if err != nil {
		t.Error("Expect error to be nil")
	}
	if repoName != "jmyue/kubernetes-plugin" {
		t.Errorf("Expect result %d equals to `jmyue/kubernetes-plugin`", repoName)
	}

	url = `https://github.com/caicloud/cyclone.git`
	repoName, err = getRepoNameByURL(url)
	if err != nil {
		t.Error("Expect error to be nil")
	}
	if repoName != "caicloud/cyclone" {
		t.Errorf("Expect result %d equals to `caicloud/cyclone`", repoName)
	}

	url = `http://192.168.21.100:10080/jmyue/kubernetes-plugin.git`
	repoName, err = getRepoNameByURL(url)
	if err != nil {
		t.Error("Expect error to be nil")
	}
	if repoName != "jmyue/kubernetes-plugin" {
		t.Errorf("Expect result %d equals to `jmyue/kubernetes-plugin`", repoName)
	}

	url = `http://localhost:10080/ci-test/ci-demo.git`
	repoName, err = getRepoNameByURL(url)
	if err != nil {
		t.Error("Expect error to be nil")
	}
	if repoName != "ci-test/ci-demo" {
		t.Errorf("Expect result %d equals to `ci-test/ci-demo`", repoName)
	}
}
