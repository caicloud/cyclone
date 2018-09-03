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
	"reflect"
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

func TestPullImage(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockClient := mock_docker.NewMockClientInterface(ctl)
	defaultAuth := docker_client.AuthConfiguration{
		ServerAddress: "registry.caicloud.io",
		Username:      "caicloud",
		Password:      "123456",
	}
	opts1 := docker_client.PullImageOptions{
		Repository: "image1",
	}
	mockClient.EXPECT().PullImage(opts1, defaultAuth).Return(nil)

	opts2 := docker_client.PullImageOptions{
		Repository: "image2",
	}
	auth2 := docker_client.AuthConfiguration{
		ServerAddress: "registry2.caicloud.io",
		Username:      "caicloud",
		Password:      "111111",
	}
	mockClient.EXPECT().PullImage(opts2, docker_client.AuthConfiguration{}).Return(nil)

	opts3 := docker_client.PullImageOptions{
		Repository: "registry3.caicloud.io/caicloud/image3",
	}
	auth3 := docker_client.AuthConfiguration{
		ServerAddress: "registry3.caicloud.io",
		Username:      "caicloud",
		Password:      "111111",
	}
	mockClient.EXPECT().PullImage(opts3, auth3).Return(nil)

	dm := &DockerManager{
		Client:     mockClient,
		AuthConfig: &defaultAuth,
	}

	testCases := map[string]struct {
		image string
		auth  docker_client.AuthConfiguration
	}{
		"image1 using manager auth": {
			"image1",
			docker_client.AuthConfiguration{},
		},
		"image2 without auth": {
			"image2",
			auth2,
		},
		"image3 using provided auth": {
			"registry3.caicloud.io/caicloud/image3",
			auth3,
		},
	}

	for d, tc := range testCases {
		err := dm.PullImage(tc.image, tc.auth)
		if err != nil {
			t.Errorf("Fail to pull image %s as %v", d, err)
		}
	}
}

func TestPushImage(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockClient := mock_docker.NewMockClientInterface(ctl)
	defaultAuth := docker_client.AuthConfiguration{
		ServerAddress: "registry.caicloud.io",
		Username:      "caicloud",
		Password:      "123456",
	}
	opts1 := docker_client.PushImageOptions{
		Name: "image1",
	}
	mockClient.EXPECT().PushImage(opts1, defaultAuth).Return(nil)

	opts2 := docker_client.PushImageOptions{
		Name: "image2",
		Tag:  "v3",
	}
	auth2 := docker_client.AuthConfiguration{
		ServerAddress: "registry2.caicloud.io",
		Username:      "caicloud",
		Password:      "111111",
	}
	mockClient.EXPECT().PushImage(opts2, auth2).Return(nil)

	opts3 := docker_client.PushImageOptions{
		Registry: "registry3.caicloud.io",
		Name:     "image3",
		Tag:      "v3",
	}
	auth3 := docker_client.AuthConfiguration{
		ServerAddress: "registry3.caicloud.io",
		Username:      "caicloud",
		Password:      "111111",
	}
	mockClient.EXPECT().PushImage(opts3, auth3).Return(nil)

	dm := &DockerManager{
		Client:     mockClient,
		AuthConfig: &defaultAuth,
	}

	testCases := map[string]struct {
		image docker_client.PushImageOptions
		auth  docker_client.AuthConfiguration
	}{
		"image1 using manager auth": {
			opts1,
			docker_client.AuthConfiguration{},
		},
		"image2 using provided auth": {
			opts2,
			auth2,
		},
		"image3 using provided auth": {
			opts3,
			auth3,
		},
	}

	for d, tc := range testCases {
		err := dm.PushImage(tc.image, tc.auth)
		if err != nil {
			t.Errorf("Fail to push image %s as %v", d, err)
		}
	}
}

func TestRemoveContainer(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	mockClient := mock_docker.NewMockClientInterface(ctl)
	mockClient.EXPECT().RemoveContainer(docker_client.RemoveContainerOptions{
		ID:    "c1",
		Force: true,
	}).Return(nil)
	mockClient.EXPECT().RemoveContainer(docker_client.RemoveContainerOptions{
		ID:    "c2",
		Force: true,
	}).Return(&docker_client.NoSuchContainer{"c2", nil})

	dm := &DockerManager{
		Client: mockClient,
	}

	testCases := map[string]struct {
		cid string
		err error
	}{
		"c1": {
			"c1",
			nil,
		},
		"c2": {
			"c2",
			&docker_client.NoSuchContainer{"c2", nil},
		},
	}

	for d, tc := range testCases {
		err := dm.RemoveContainer(tc.cid)
		if !reflect.DeepEqual(err, tc.err) {
			t.Errorf("Fail to remove container %s: expect err as %v, but got %v", d, tc.err, err)
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
