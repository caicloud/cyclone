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

package common

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/worker/helper"
	docker_client "github.com/fsouza/go-dockerclient"
)

const (
	// Fake token for request.
	fakeToken = "fake-token"
	// Console-web endpoint for canary.
	defaultConsoleWebEndpoint = "https://console-canary.caicloud.io"
)

// Generate the request with the global token and the given contentType.
func generateRequestWithToken(request *http.Request, contentType string) error {
	request.Header.Add("content-type", contentType)
	request.Header.Add("token", fakeToken)
	return nil
}

// PushImageToLocalRegistry pushes the image to local registry.
func PushImageToLocalRegistry(dm *docker.Manager, image string) error {
	if strings.Index(image, ":") == -1 {
		return errors.New("Invalid image name.")
	}
	parts := strings.Split(image, ":")
	return dm.Client.PushImage(docker_client.PushImageOptions{
		Name:     fmt.Sprintf("%s/%s", dm.Registry, parts[0]),
		Tag:      parts[1],
		Registry: dm.Registry,
	}, docker_client.AuthConfiguration{
		Username: dm.AuthConfig.Username,
		Password: dm.AuthConfig.Password,
	})
}

// CallUpdateImageAPI invokes update image API.
func CallUpdateImageAPI(userID, applicationName, clusterName, partitionName, containerName, imageName string) error {
	endpoint := defaultConsoleWebEndpoint + "/api/application/updateImage"
	if err := helper.InvokeUpdateImageAPI(userID, applicationName, clusterName, partitionName, containerName,
		imageName, endpoint); err != nil {
		return err
	}
	return nil
}
