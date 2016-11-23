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

package helper

import (
	"fmt"
	"os"
	"testing"

	"github.com/caicloud/cyclone/api"

	"gopkg.in/h2non/gock.v1"
)

// TestPushToCyclone tests pushing log to cyclone.
func TestPushToCyclone(t *testing.T) {
	mockServer := "http://mock.caicloud.com"
	userID := "unit-test-uid"
	versionID := "unit-test-vid"

	if err := os.Setenv(SERVER_HOST, mockServer); err != nil {
		t.Error("Expected error to be nil")
	}

	pushLogAPI = fmt.Sprintf("/api/%s/%s/versions/%s/logs", api.APIVersion, userID, versionID)

	gock.New(mockServer).
		Post(pushLogAPI).
		Reply(200)

	defer gock.Off()

	event := &api.Event{
		Service: api.Service{
			UserID: userID,
		},
		Version: api.Version{
			VersionID: versionID,
		},
	}

	if err := PushLogToCyclone(event); err != nil {
		t.Errorf("Expect error %v to be nil", err)
	}
}
