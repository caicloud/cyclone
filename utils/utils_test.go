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

package utils

import (
	"testing"

	"gopkg.in/h2non/gock.v1"
)

// TestInvokeUpdateImageAPI tests InvokeUpdateImageAPI function.
func TestInvokeUpdateImageAPI(t *testing.T) {
	// Mock the endpoint.
	gock.New("http://mock.caicloud.com").
		Post("/InvokeUpdateImageAPI").
		Reply(200)

	defer gock.Off()

	applicationName := "applicationName"
	partitionName := "partitionName"
	clusterName := "clusterName"
	containerName := "containerName"
	imageName := "imageName"
	userID := "caicloud"
	endpoint := "http://mock.caicloud.com/InvokeUpdateImageAPI"

	if err := InvokeUpdateImageAPI(userID, applicationName, clusterName, partitionName, containerName, imageName, endpoint); err != nil {
		t.Error("Expect error to be nil")
	}
}

// TestInvokeCheckDeployStateAPI tests InvokeCheckDeployStateAPI function.
func TestInvokeCheckDeployStateAPI(t *testing.T) {
	// Mock the endpoint.
	gock.New("http://mock.caicloud.com").
		Post("/InvokeCheckDeployStateAPI").
		Reply(200).
		JSON(map[string]int{
		"code": codeDeployReady,
	})

	defer gock.Off()

	jsonStr := []byte{}
	endpoint := "http://mock.caicloud.com/InvokeCheckDeployStateAPI"

	if result, err := InvokeCheckDeployStateAPI(jsonStr, endpoint); result != true || err != nil {
		t.Error("Expect error to be nil")
	}
}
