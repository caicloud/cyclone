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

package websocket

import (
	"net/http"
	"reflect"
	"testing"
)

func TestFilterHeader(t *testing.T) {
	testCases := map[string]struct {
		input  http.Header
		output http.Header
	}{
		"header without additional keys": {
			input: http.Header{
				"Upgrade":                  []string{},
				"Connection":               []string{},
				"Sec-Websocket-Key":        []string{},
				"Sec-Websocket-Version":    []string{},
				"Sec-Websocket-Extensions": []string{},
			},
			output: http.Header{},
		},
		"header with additional keys": {
			input: http.Header{
				"Upgrade":                  []string{},
				"Connection":               []string{},
				"Sec-Websocket-Key":        []string{},
				"Sec-Websocket-Version":    []string{},
				"Sec-Websocket-Extensions": []string{},
				"X-User":                   []string{"robin"},
				"Authorization":            []string{"aasdfljaoefalsdfja[osej"},
			},
			output: http.Header{
				"X-User":        []string{"robin"},
				"Authorization": []string{"aasdfljaoefalsdfja[osej"},
			},
		},
	}

	for d, tc := range testCases {
		result := FilterHeader(tc.input)
		if !reflect.DeepEqual(result, tc.output) {
			t.Errorf("Fail to filter %s: expect %v, but got %v", d, tc.output, result)
		}
	}
}
