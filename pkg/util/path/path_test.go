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

package pathutil

import (
	"os"
	"testing"
)

const (
	legalPath = "/var"
)

// TestEnsureParentDir tests the EnsureParentDir func.
func TestEnsureParentDir(t *testing.T) {
	if err := EnsureParentDir(legalPath, os.ModePerm); err != nil {
		t.Error("Expected error to be nil")
	}
}
