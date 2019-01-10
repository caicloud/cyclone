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

package file

import (
	"testing"
)

func TestDirExists(t *testing.T) {
	d := map[string]struct {
		path   string
		expect bool
	}{
		"check parent dir": {
			"../file",
			true,
		},
		"check wrong dir": {
			"abc",
			false,
		},
	}

	for k, v := range d {
		r := DirExists(v.path)
		if r != v.expect {
			t.Errorf("%s failed: expected %v, but got %v", k, r, v.expect)
		}
	}
}

func TestFileExists(t *testing.T) {
	d := map[string]struct {
		path   string
		expect bool
	}{
		"check correct file": {
			"file.go",
			true,
		},
		"check wrong file": {
			"abc",
			false,
		},
	}

	for k, v := range d {
		r := Exists(v.path)
		if r != v.expect {
			t.Errorf("%s failed: expected %v, but got %v", k, v.expect, r)
		}
	}
}
