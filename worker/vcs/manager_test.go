/*
Copyright 2016 caicloud authors. All rights reserved.

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

package vcs

import (
	"fmt"
	"testing"

	"github.com/caicloud/cyclone/api"
)

const (
	subvcs = api.GITHUB
	URL    = "github.com"
	token  = "TestToken"
)

// TestManager tests getUrlwithToken function.
func TestManager(t *testing.T) {
	expectedResult := fmt.Sprintf("%s@%s", token, URL)
	if result := getUrlwithToken(URL, subvcs, token); result != expectedResult {
		t.Errorf("Expect result %s equals to %s", result, expectedResult)
	}
}
