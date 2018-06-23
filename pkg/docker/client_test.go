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

package docker

import (
	"fmt"
	"testing"

	docker_client "github.com/fsouza/go-dockerclient"
	"github.com/golang/mock/gomock"

	mock_docker "github.com/caicloud/cyclone/pkg/docker/mock"
)

func TestIsImagePresent(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockClient := mock_docker.NewMockClientInterface(ctl)
	mockClient.EXPECT().InspectImage("cyclone").Return(nil, nil)
	mockClient.EXPECT().InspectImage("abc").Return(nil, docker_client.ErrNoSuchImage)
	mockClient.EXPECT().InspectImage("123").Return(nil, fmt.Errorf("test error"))

	dm := &DockerManager{
		Client: mockClient,
	}

	testCases := map[string]struct {
		image   string
		present bool
	}{
		"existent image": {
			"cyclone",
			true,
		},
		"nonexistent image": {
			"abc",
			false,
		},
	}

	for d, tc := range testCases {
		result, _ := dm.IsImagePresent(tc.image)
		if result != tc.present {
			t.Errorf("Fail to judge %s: expect %t, but got %t", d, tc.present, result)
		}
	}

	errorTestCases := map[string]struct {
		image   string
		present bool
	}{
		"error image": {
			"123",
			false,
		},
	}

	for d, tc := range errorTestCases {
		result, err := dm.IsImagePresent(tc.image)
		if err == nil {
			t.Errorf("Fail to judge %s: expect err not nil", d)
		}

		if result != tc.present {
			t.Errorf("Fail to judge %s: expect %t, but got %t", d, tc.present, result)
		}
	}
}

func TestAppendLatestTagIfNecessary(t *testing.T) {
	testCases := map[string]struct {
		image    string
		expected string
	}{
		"image with tag": {
			"cyclone-server:v1.0.0",
			"cyclone-server:v1.0.0",
		},
		"image without tag": {
			"cyclone-server",
			"cyclone-server:latest",
		},
	}

	for d, tc := range testCases {
		result := AppendLatestTagIfNecessary(tc.image)
		if result != tc.expected {
			t.Errorf("Fail to handle %s: expect %s, but got %s", d, tc.expected, result)
		}
	}
}
