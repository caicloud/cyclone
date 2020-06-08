/*
Copyright 2020 caicloud authors. All rights reserved.

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

package slice

import (
	"reflect"
	"testing"
)

func TestContainsString(t *testing.T) {
	d := map[string]struct {
		slice  []string
		s      string
		expect bool
	}{
		"slice nil": {
			nil,
			"a",
			false,
		},
		"contain": {
			[]string{"a", "b"},
			"a",
			true,
		},
		"not contain": {
			[]string{"a", "b"},
			"c",
			false,
		},
	}

	for k, v := range d {
		r := ContainsString(v.slice, v.s)
		if r != v.expect {
			t.Errorf("%s failed: expected %v, but got %v", k, r, v.expect)
		}
	}
}

func TestRemoveString(t *testing.T) {
	d := map[string]struct {
		slice  []string
		s      string
		expect []string
	}{
		"slice nil": {
			nil,
			"a",
			nil,
		},
		"contain": {
			[]string{"a", "b"},
			"a",
			[]string{"b"},
		},
		"not contain": {
			[]string{"a", "b"},
			"c",
			[]string{"a", "b"},
		},
	}

	for k, v := range d {
		r := RemoveString(v.slice, v.s)
		if !reflect.DeepEqual(r, v.expect) {
			t.Errorf("%s failed: expected %v, but got %v", k, r, v.expect)
		}
	}
}
