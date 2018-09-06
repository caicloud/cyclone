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

package handler

import (
	"testing"
)

func TestCheckAndTransTimes(t *testing.T) {
	d := map[string]struct {
		start string
		end   string
	}{
		"end = start + n*86400": {
			"1521345720",
			"1522468920",
		},
		"end-1 = start + n*86400": {
			"1521345720",
			"1522468919",
		},
		"end+1 = start + n*86400": {
			"1521345720",
			"1522468921",
		},
		"end = start": {
			"1521345720",
			"1521345720",
		},
	}

	for k, v := range d {
		s, e, err := checkAndTransTimes(v.start, v.end)
		if err != nil {
			t.Errorf("%s failed: expected nil, but got %v, startTime:%v , endTime:%v", k, err, s, e)
		}
	}
}

func TestCheckAndTransTimesErr(t *testing.T) {
	d := map[string]struct {
		start string
		end   string
	}{
		"empty ": {
			"",
			"",
		},
		"end < start": {
			"1521345720",
			"1521345719",
		},
		"start error": {
			"abc",
			"1521345719",
		},
		"end error": {
			"1521345720",
			"abc",
		},
	}

	for k, v := range d {
		_, _, err := checkAndTransTimes(v.start, v.end)
		if err == nil {
			t.Errorf("%s failed: expected err, but got nil", k)
		}
	}
}
