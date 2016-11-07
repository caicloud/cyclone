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

package provider

import (
	"os"
	"testing"

	"github.com/caicloud/cyclone/api"
)

const (
	mockSMTPServer   = "localhost"
	mockSMTPPort     = 8504
	mockSMTPUsername = "caicloud"
	mockSMTPPassword = "password"
)

// TestReadConfig tests readContextFromConfigFile function.
func TestReadConfig(t *testing.T) {
	if err := os.Setenv(SUCCESSTEMPLATE, "./success.html"); err != nil {
		t.Error("Expected error to be nil")
	}
	if err := os.Setenv(ERRORTEMPLATE, "./error.html"); err != nil {
		t.Error("Expected error to be nil")
	}
	if err := readContextFromConfigFile(); err != nil {
		t.Error("Expected error to be nil")
	}
}

// TestSendEmailWithWrongConfig sends email with wrong config.
func TestSendEmailWithWrongConfig(t *testing.T) {
	smtpServer, err := NewEmailNotifier(mockSMTPServer, mockSMTPPort, mockSMTPUsername, mockSMTPPassword)
	if err != nil {
		t.Error("Expected error to be nil")
	}
	if err := smtpServer.Notify(&api.Service{}, &api.Version{}, ""); err == nil {
		t.Error("Expected error to occur but it was nil")
	}
}
