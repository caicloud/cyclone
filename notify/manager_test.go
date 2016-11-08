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

package notify

import (
	"testing"

	"github.com/caicloud/cyclone/api"
)

const (
	mockSMTPServer   = "localhost"
	mockSMTPPort     = 8504
	mockSMTPUsername = "caicloud"
	mockSMTPPassword = "password"
)

// TestEmailSendLogic tests the email send logic.
func TestEmailSendLogic(t *testing.T) {
	service := api.Service{
		Profile: api.NotifyProfile{
			Setting: api.SendWhenFailed,
		},
	}

	// the version is healthy, its YamlDeployStatus is health, but its DeployPlansStatuses
	// are bad.
	version := api.Version{
		Status:           api.VersionHealthy,
		YamlDeployStatus: api.DeploySuccess,
		DeployPlansStatuses: []api.DeployPlanStatus{
			{
				Status: api.DeployFailed,
			},
		},
	}

	if shouldSend := shouldSendNotifyEvent(&service, &version); shouldSend != true {
		t.Error("Expected send the mail but not.")
	}
}

// TestNewManagerWithoutTemplate tests the manager with wrong template path.
func TestNewManagerWithoutTemplate(t *testing.T) {
	_, err := newManager(mockSMTPServer, mockSMTPPort, mockSMTPUsername, mockSMTPPassword)
	if err == nil {
		t.Error("Expected error to occur but it was nil")
	}
}
