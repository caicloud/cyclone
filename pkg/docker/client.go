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
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	docker_client "github.com/fsouza/go-dockerclient"
	log "github.com/golang/glog"
)

const (
	defaultEndpoint = "unix:///var/run/docker.sock"
)

// DockerManager represents the manager of Docker, it packages the Docker client to easily use it.
// The Docker client can be direclty used for some functions not provided by this manager.
type DockerManager struct {
	// Client represets the Docker client.
	Client *docker_client.Client
}

func NewDockerManager(endpoint string) (*DockerManager, error) {
	if len(strings.TrimSpace(endpoint)) == 0 {
		endpoint = defaultEndpoint
	}

	client, err := docker_client.NewClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("new Docker client with error %s", err.Error())
	}

	if _, err := client.Version(); err != nil {
		return nil, fmt.Errorf("connect Docker server with error %s", err.Error())
	}

	return &DockerManager{
		Client: client,
	}, nil
}

func (dm *DockerManager) StartContainer(options docker_client.CreateContainerOptions) (string, error) {
	// Create the container
	container, err := dm.Client.CreateContainer(options)
	if err != nil {
		return "", fmt.Errorf("create container with error %s", err.Error())
	}

	// Run the container
	err = dm.Client.StartContainer(container.ID, options.HostConfig)
	if err != nil {
		return "", fmt.Errorf("start container with error %s", err.Error())
	}

	return container.ID, nil
}

// ExecOptions specify parameters to the ExecInContainer function.
type ExecOptions struct {
	Cmd       []string
	Container string
	User      string

	// InputStream  io.Reader
	OutputStream io.Writer
	ErrorStream  io.Writer
}

func (dm *DockerManager) ExecInContainer(options ExecOptions) error {
	// Create the exec instance in the running container.
	// In order to return after the command finishes, the options must attach the stdout and stderr,
	// and set their writer stream.
	ceo := docker_client.CreateExecOptions{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          options.Cmd,
		Container:    options.Container,
	}
	exec, err := dm.Client.CreateExec(ceo)
	if err != nil {
		return fmt.Errorf("create exec instance in container %s with error %s", ceo.Container, err.Error())
	}

	// Start the exec instance
	seo := docker_client.StartExecOptions{
		ErrorStream:  options.ErrorStream,
		OutputStream: options.OutputStream,
	}
	err = dm.Client.StartExec(exec.ID, seo)
	if err != nil {
		return fmt.Errorf("start command %s in container %s with error %s", ceo.Cmd, ceo.Container, err.Error())
	}

	// Check the exit code of the exec instance
	execInspect, err := dm.Client.InspectExec(exec.ID)
	if err != nil {
		return fmt.Errorf("inspect command %s in container %s with error %s", ceo.Cmd, ceo.Container, err.Error())
	}

	if execInspect.ExitCode != 0 {
		return fmt.Errorf("command %s failed in container %s", ceo.Cmd, ceo.Container)
	}

	return nil
}

// CopyFromContainerOptions specify parameters download resources from a container.
type CopyFromContainerOptions struct {
	Container     string
	HostPath      string
	ContainerPath string
}

func (dm *DockerManager) CopyFromContainer(options CopyFromContainerOptions) error {
	var buf bytes.Buffer
	dfco := docker_client.DownloadFromContainerOptions{
		Path:         options.ContainerPath,
		OutputStream: &buf,
	}
	err := dm.Client.DownloadFromContainer(options.Container, dfco)

	if err != nil {
		err = fmt.Errorf("copy %s from container %s with error %s", options.ContainerPath, options.Container, err.Error())
		log.Error(err)
		return err
	}

	reader := bytes.NewReader(buf.Bytes())
	tarReader := tar.NewReader(reader)

	for {
		var fileHead *tar.Header
		fileHead, err = tarReader.Next()
		if err == io.EOF {
			log.Info("docker copy file finished")
			break
		} else if err != nil {
			log.Error(err)
			return err
		}

		if fileHead.FileInfo().IsDir() {
			os.Mkdir(options.HostPath+fileHead.Name, os.FileMode(fileHead.Mode))
		} else {
			var fileOutPut *os.File
			fileOutPut, err = os.OpenFile(options.HostPath+fileHead.Name,
				os.O_CREATE|os.O_WRONLY, os.FileMode(fileHead.Mode))
			if err != nil {
				log.Error(err)
				return err
			}

			if _, err := io.Copy(fileOutPut, tarReader); err != nil {
				log.Error(err)
				return err
			}
			fileOutPut.Close()
		}
	}
	return nil
}
