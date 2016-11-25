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
	"os"
	"strings"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/filebuffer"
	"github.com/caicloud/cyclone/pkg/log"
	steplog "github.com/caicloud/cyclone/worker/log"
	"github.com/docker/docker/builder/dockerfile/command"
	docker_parse "github.com/docker/docker/builder/dockerfile/parser"
	docker_client "github.com/fsouza/go-dockerclient"
)

// Manager manages all docker operations, like build, push, etc.
type Manager struct {
	Client     *docker_client.Client
	Registry   string
	AuthConfig *AuthConfig
	EndPoint   string
}

// NewManager creates a new docker manager.
func NewManager(endpoint string, certPath string, registry api.RegistryCompose) (*Manager, error) {
	// Get the AuthConfig from username and password in SYSTEM ENV.
	authConfig, err := NewAuthConfig(registry.RegistryUsername, registry.RegistryPassword)
	if err != nil {
		return nil, err
	}

	if certPath == "" {
		client, err := docker_client.NewClient(endpoint)
		if err != nil {
			return nil, err
		}

		return &Manager{
			Client:     client,
			Registry:   registry.RegistryLocation,
			AuthConfig: authConfig,
			EndPoint:   endpoint,
		}, nil
	}

	cert := fmt.Sprintf("%s/cert.pem", certPath)
	key := fmt.Sprintf("%s/key.pem", certPath)
	ca := fmt.Sprintf("%s/ca.pem", certPath)

	client, err := docker_client.NewTLSClient(endpoint, cert, key, ca)
	if err != nil {
		return nil, err
	}

	_, err = client.Version()
	if err != nil {
		log.Errorf("error connecting to docker daemon %s. %s.", endpoint, err)
		return nil, err
	}

	return &Manager{
		Client:     client,
		Registry:   registry.RegistryLocation,
		AuthConfig: authConfig,
		EndPoint:   endpoint,
	}, nil
}

// PullImage pulls an image by its name.
func (dm *Manager) PullImage(imageName string) error {
	opts := docker_client.PullImageOptions{
		Repository: imageName,
		Registry:   dm.Registry,
	}

	authOpt := docker_client.AuthConfiguration{
		Username: dm.AuthConfig.Username,
		Password: dm.AuthConfig.Password,
	}

	err := dm.Client.PullImage(opts, authOpt)
	if err == nil {
		log.InfoWithFields("Successfully pull docker image.", log.Fields{"image": imageName})
	}
	return err
}

// BuildImage builds image from event.
func (dm *Manager) BuildImage(event *api.Event, output filebuffer.FileBuffer) error {
	imagename, ok := event.Data["image-name"]
	tagname, ok2 := event.Data["tag-name"]
	contextdir, ok3 := event.Data["context-dir"]
	if !ok || !ok2 || !ok3 {
		return fmt.Errorf("Unable to retrieve image name")
	}

	imageName := imagename.(string) + ":" + tagname.(string)
	log.InfoWithFields("About to build docker image.", log.Fields{"image": imageName})

	contextDir := contextdir.(string)

	// Use to pull cargo.caicloud.io/:username/:imagename:tag.
	// TODO: we will consider more cases
	authOpt := docker_client.AuthConfiguration{
		Username: dm.AuthConfig.Username,
		Password: dm.AuthConfig.Password,
	}

	authOpts := docker_client.AuthConfigurations{
		Configs: make(map[string]docker_client.AuthConfiguration),
	}
	authOpts.Configs[dm.Registry] = authOpt

	opt := docker_client.BuildImageOptions{
		Name:           imageName,
		ContextDir:     contextDir,
		AuthConfigs:    authOpts,
		RmTmpContainer: true,
		Memswap:        -1,
		OutputStream:   output,
	}
	err := dm.Client.BuildImage(opt)
	if err == nil {
		log.InfoWithFields("Successfully built docker image.", log.Fields{"image": imageName})
	}
	return err
}

// PushImage pushes docker image to registry. output will be sent to event status output.
func (dm *Manager) PushImage(event *api.Event, output filebuffer.FileBuffer) error {
	imageName, ok := event.Data["image-name"]
	tagName, ok2 := event.Data["tag-name"]

	if !ok || !ok2 {
		return fmt.Errorf("Unable to retrieve image name")
	}

	log.InfoWithFields("About to push docker image.", log.Fields{"image": imageName, "tag": tagName})

	opt := docker_client.PushImageOptions{
		Name:         imageName.(string),
		Tag:          tagName.(string),
		OutputStream: output,
	}

	authOpt := docker_client.AuthConfiguration{
		Username: dm.AuthConfig.Username,
		Password: dm.AuthConfig.Password,
	}

	err := dm.Client.PushImage(opt, authOpt)
	if err == nil {
		log.InfoWithFields("Successfully pushed docker image.", log.Fields{"image": imageName})
	}

	return err
}

// RunContainer runs a container according to special config
func (dm *Manager) RunContainer(cco *docker_client.CreateContainerOptions) (string, error) {
	isImageExisted, err := dm.IsImagePresent(cco.Config.Image)
	if err != nil {
		return "", err
	}

	if isImageExisted == false {
		log.InfoWithFields("About to pull the image.", log.Fields{"image": cco.Config.Image})
		err := dm.PullImage(cco.Config.Image)
		if err != nil {
			return "", err
		}
		log.InfoWithFields("Successfully pull the image.", log.Fields{"image": cco.Config.Image})
	}

	log.InfoWithFields("About to create the container.", log.Fields{"config": *cco})
	client := dm.Client
	container, err := client.CreateContainer(*cco)
	if err != nil {
		return "", err
	}

	err = client.StartContainer(container.ID, cco.HostConfig)
	if err != nil {
		client.RemoveContainer(docker_client.RemoveContainerOptions{
			ID: container.ID,
		})
		return "", err
	}

	log.InfoWithFields("Successfully create the container.", log.Fields{"config": *cco})
	return container.ID, nil
}

// StopContainer stops a container by given ID.
func (dm *Manager) StopContainer(ID string) error {
	return dm.Client.StopContainer(ID, 0)
}

// RemoveContainer removes a container by given ID.
func (dm *Manager) RemoveContainer(ID string) error {
	return dm.Client.RemoveContainer(docker_client.RemoveContainerOptions{
		ID:            ID,
		RemoveVolumes: true,
		Force:         true,
	})
}

// StopAndRemoveContainer stops and removes a container by given ID.
func (dm *Manager) StopAndRemoveContainer(ID string) error {
	if err := dm.StopContainer(ID); err != nil {
		return err
	}
	return dm.RemoveContainer(ID)
}

// GetAuthOpts gets Auth options.
func (dm *Manager) GetAuthOpts() (authOpts docker_client.AuthConfigurations) {
	authOpt := docker_client.AuthConfiguration{
		Username: dm.AuthConfig.Username,
		Password: dm.AuthConfig.Password,
	}

	authOpts = docker_client.AuthConfigurations{
		Configs: make(map[string]docker_client.AuthConfiguration),
	}
	authOpts.Configs[dm.Registry] = authOpt

	return authOpts
}

// BuildImageSpecifyDockerfile builds docker image with params from event with
// specify Dockerfile. Build output will be sent to event status output.
func (dm *Manager) BuildImageSpecifyDockerfile(event *api.Event,
	dockerfilePath string, dockerfileName string, output filebuffer.FileBuffer) error {
	imagename, ok := event.Data["image-name"]
	tagname, ok2 := event.Data["tag-name"]
	contextdir, ok3 := event.Data["context-dir"]
	if !ok || !ok2 || !ok3 {
		return fmt.Errorf("Unable to retrieve image name")
	}
	imageName := imagename.(string) + ":" + tagname.(string)
	log.InfoWithFields("About to build docker image.", log.Fields{"image": imageName})

	contextDir := contextdir.(string)

	// Use to pull cargo.caicloud.io/:username/:imagename:tag.
	// TODO: we will consider more cases
	authOpt := docker_client.AuthConfiguration{
		Username: dm.AuthConfig.Username,
		Password: dm.AuthConfig.Password,
	}

	authOpts := docker_client.AuthConfigurations{
		Configs: make(map[string]docker_client.AuthConfiguration),
	}
	authOpts.Configs[dm.Registry] = authOpt

	if "" != dockerfilePath {
		contextDir = contextDir + "/" + dockerfilePath
	}

	if "" == dockerfileName {
		dockerfileName = "Dockerfile"
	}
	opt := docker_client.BuildImageOptions{
		Name:           imageName,
		Dockerfile:     dockerfileName,
		ContextDir:     contextDir,
		OutputStream:   output,
		AuthConfigs:    authOpts,
		RmTmpContainer: true,
		Memswap:        -1,
	}
	steplog.InsertStepLog(event, steplog.BuildImage, steplog.Start, nil)
	err := dm.Client.BuildImage(opt)
	if err == nil {
		steplog.InsertStepLog(event, steplog.BuildImage, steplog.Finish, nil)
		log.InfoWithFields("Successfully built docker image.", log.Fields{"image": imageName})
	} else {
		steplog.InsertStepLog(event, steplog.BuildImage, steplog.Stop, err)
	}

	return err
}

// CleanUp removes images generated during building image
func (dm *Manager) CleanUp(event *api.Event) error {
	imagename, ok := event.Data["image-name"]
	tagname, ok2 := event.Data["tag-name"]
	if !ok || !ok2 {
		return fmt.Errorf("Unable to retrieve image name")
	}
	imageName := imagename.(string) + ":" + tagname.(string)
	log.InfoWithFields("About to clean up docker image.", log.Fields{"image": imageName})

	err := dm.RemoveImage(imageName)
	if err == nil {
		log.InfoWithFields("Successfully remove docker image.", log.Fields{"image": imageName})
	} else {
		log.InfoWithFields("remove docker image failed.", log.Fields{"image": imageName})
	}

	return err
}

// RemoveImage removes an image by its name or ID.
func (dm *Manager) RemoveImage(name string) error {
	return dm.Client.RemoveImage(name)
}

// parse parses the "FROM" in the repo's Dockerfile to check the images which the build images base on
// It returns two parameters, the first one is used for recording image name, the second is used
// For storage the error inforamtion.
func parse(despath string) ([]string, error) {
	var str []string
	f, err := os.Open(despath + "/Dockerfile")
	if err != nil {
		log.ErrorWithFields("open dockerfile fail", log.Fields{"error": err})
		return str, err
	}

	defer f.Close()

	nodes, _ := docker_parse.Parse(f)
	for _, node := range nodes.Children {
		if node.Value == command.From {
			if node.Next != nil {
				for n := node.Next; n != nil; n = n.Next {
					str = append(str, n.Value)
				}
			}
			break
		}
	}
	if len(str) <= 0 {
		return str, fmt.Errorf("there is no FROM")
	}
	return str, nil
}

// IsImagePresent checks if given image exists.
func (dm *Manager) IsImagePresent(image string) (bool, error) {
	_, err := dm.Client.InspectImage(image)
	if err == nil {
		return true, nil
	}
	if err == docker_client.ErrNoSuchImage {
		return false, nil
	}
	return false, err
}

// GetImageNameWithTag gets the image name with tag from registry, username, service name and version name.
func (dm *Manager) GetImageNameWithTag(username, serviceName, versionName string) string {
	return fmt.Sprintf("%s/%s/%s:%s", dm.Registry, strings.ToLower(username), strings.ToLower(serviceName), versionName)
}

// GetImageNameNoTag gets the image name without tag from registry, username, service name.
func (dm *Manager) GetImageNameNoTag(username, serviceName string) string {
	return fmt.Sprintf("%s/%s/%s", dm.Registry, strings.ToLower(username), strings.ToLower(serviceName))
}
