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

package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// TestLog tests the log service.
func TestLog(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(logHandler))
	defer s.Close()

	if err := os.Setenv(LOG_HTML_TEMPLATE, "../http/web/log.html"); err != nil {
		t.Error("Expected error to be nil")
	}

	res, err := http.Get(s.URL)
	if err != nil {
		t.Error("Expected get status 200")
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Error("Expected get body")
	}
	t.Log(body)
}
