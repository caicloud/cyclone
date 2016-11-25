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

// Package runner is an implementation of job runner.
package runner

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/caicloud/cyclone/api"

	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/filebuffer"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/worker/ci/parser"
	docker_client "github.com/fsouza/go-dockerclient"
)

const (
	// BindTemplate is the template of binds.
	BindTemplate = "%s:%s"

	// BuiltImage is the flag of running a container which use the built image as a service.
	BuiltImage = "BUILT_IMAGE"
)

// getNameInNetwork gets the name of the container in network.
func getNameInNetwork(dn *parser.DockerNode, b *Build) string {
	return fmt.Sprintf("%s", dn.Name)
}

// toServiceContainerConfig creates CreateContainerOptions from ServiceNode.
func toServiceContainerConfig(dn *parser.DockerNode, b *Build) *docker_client.CreateContainerOptions {

	name := getNameInNetwork(dn, b)

	// If the image of service config to be "BUILT_IMAGE",
	// Cyclone will use the image built during the "build" step to run a service container.
	if dn.Image == BuiltImage {
		imageName, ok := b.event.Data["image-name"]
		tagName, ok2 := b.event.Data["tag-name"]

		if !ok || !ok2 {
			return nil
		}

		log.InfoWithFields("About to run service container for integration. ",
			log.Fields{"image": imageName, "tag": tagName})
		dn.Image = fmt.Sprintf("%s:%s", imageName.(string), tagName.(string))
	}

	config := &docker_client.Config{
		Image:      dn.Image,
		Env:        dn.Environment,
		Cmd:        dn.Command,
		Entrypoint: dn.Entrypoint,
	}
	hostConfig := &docker_client.HostConfig{
		Privileged:       dn.Privileged,
		MemorySwappiness: -1,
	}

	if len(dn.Command) == 0 {
		config.Cmd = nil
	}
	if len(dn.ExtraHosts) > 0 {
		hostConfig.ExtraHosts = dn.ExtraHosts
	}
	if len(dn.DNS) != 0 {
		hostConfig.DNS = dn.DNS
	}

	// Parse the Volumes from "%s:%s", then set them into hostConfig.
	for _, path := range dn.Volumes {
		if strings.Index(path, ":") == -1 {
			continue
		}
		_, path = mockDotReference(b, path)
		// TODO: log the failure.
		if validatePath(b, path) {
			hostConfig.Binds = append(hostConfig.Binds, path)
		}
	}

	// Parse the Devices from "%s:%s", then set them into hostConfig.
	for _, path := range dn.Devices {
		if strings.Index(path, ":") == -1 {
			continue
		}
		parts := strings.Split(path, ":")
		device := docker_client.Device{
			PathOnHost:        parts[0],
			PathInContainer:   parts[1],
			CgroupPermissions: "rwm",
		}
		hostConfig.Devices = append(hostConfig.Devices, device)
	}

	createContainerOptions := &docker_client.CreateContainerOptions{
		Name:       name,
		Config:     config,
		HostConfig: hostConfig,
	}
	return createContainerOptions
}

// toContainerConfig creates CreateContainerOptions from BuildNode.
func toBuildContainerConfig(dn *parser.DockerNode, b *Build, nodetype parser.NodeType) *docker_client.CreateContainerOptions {
	config := &docker_client.Config{
		Image:      dn.Image,
		Env:        dn.Environment,
		Cmd:        dn.Command,
		Entrypoint: dn.Entrypoint,
	}
	hostConfig := &docker_client.HostConfig{
		Privileged:       dn.Privileged,
		MemorySwappiness: -1,
	}

	if len(dn.Entrypoint) == 0 {
		config.Entrypoint = nil
	}
	if len(dn.Command) == 0 {
		config.Cmd = nil
	}
	if len(dn.ExtraHosts) > 0 {
		hostConfig.ExtraHosts = dn.ExtraHosts
	}
	if len(dn.DNS) != 0 {
		hostConfig.DNS = dn.DNS
	}

	defaultFlag := true

	// Parse the Volumes from "%s:%s", then set them into hostConfig.
	for _, path := range dn.Volumes {
		if strings.Index(path, ":") == -1 {
			continue
		}
		parts := strings.Split(path, ":")
		// Set the current directory as workingDir.
		if parts[0] == "." {
			defaultFlag = false
			config.WorkingDir = parts[1]
		}
		_, path = mockDotReference(b, path)
		// TODO: log the failure.
		if validatePath(b, path) {
			hostConfig.Binds = append(hostConfig.Binds, path)
		}
	}

	if defaultFlag == true {
		// Set the workingDir, default is the contextDir
		path := fmt.Sprintf("%s:%s", b.contextDir, b.contextDir)
		hostConfig.Binds = append(hostConfig.Binds, path)
		config.WorkingDir = b.contextDir
	}

	// b.dockerManager.EndPoint would return unix://xxxx,
	// so get the substr of the endpoint.
	// Trick: bind the docker sock file to container to support
	// docker operation in the container.
	enterpoint := []byte(b.dockerManager.EndPoint)[7:]
	log.Infof("enterpoint is %s", string(enterpoint))
	pathenterpoint := fmt.Sprintf("%s:%s", string(enterpoint), "/var/run/docker.sock")
	hostConfig.Binds = append(hostConfig.Binds, pathenterpoint)

	// Parse the Devices from "%s:%s", then set them into hostConfig.
	for _, path := range dn.Devices {
		if strings.Index(path, ":") == -1 {
			continue
		}
		parts := strings.Split(path, ":")
		device := docker_client.Device{
			PathOnHost:        parts[0],
			PathInContainer:   parts[1],
			CgroupPermissions: "rwm",
		}
		hostConfig.Devices = append(hostConfig.Devices, device)
	}

	createContainerOptions := &docker_client.CreateContainerOptions{
		Config:     config,
		HostConfig: hostConfig,
	}
	return createContainerOptions
}

// start a container with the given CreateContainerOptions.
func start(b *Build, cco *docker_client.CreateContainerOptions) (*docker_client.Container, error) {
	log.InfoWithFields("About to inspect the image.", log.Fields{"image": cco.Config.Image})
	result, err := b.dockerManager.IsImagePresent(cco.Config.Image)
	if err != nil {
		return nil, err
	}
	if result == false {
		log.InfoWithFields("About to pull the image.", log.Fields{"image": cco.Config.Image})
		err := b.dockerManager.PullImage(cco.Config.Image)
		if err != nil {
			return nil, err
		}
		log.InfoWithFields("Successfully pull the image.", log.Fields{"image": cco.Config.Image})
	}

	log.InfoWithFields("About to create the container.", log.Fields{"config": *cco})
	client := b.dockerManager.Client
	container, err := client.CreateContainer(*cco)
	if err != nil {
		return nil, err
	}
	err = client.StartContainer(container.ID, cco.HostConfig)
	if err != nil {
		// TODO: Check the error.
		client.RemoveContainer(docker_client.RemoveContainerOptions{
			ID: container.ID,
		})
		return nil, err
	}
	log.InfoWithFields("Successfully create the container.", log.Fields{"config": *cco})
	// Notice that the container wouldn't be removed before the return. So it should
	// be done at the runner.Build.TearDown().
	return container, nil
}

// run a container with the given CreateContainerOptions, currently it
// involves: start the container, wait it to stop and record the log
// into output.
func run(b *Build, cco *docker_client.CreateContainerOptions,
	outPutFiles []string, outPutPath string, nodetype parser.NodeType, output filebuffer.FileBuffer) (*docker_client.Container, error) {
	// Fetches the container information.
	client := b.dockerManager.Client
	container, err := start(b, cco)
	if err != nil {
		return nil, err
	}

	// Ensures the container is always stopped
	// and ready to be removed.
	defer func() {
		if nodetype == parser.NodeIntegration {
			for _, ID := range b.ciServiceContainers {
				b.dockerManager.StopAndRemoveContainer(ID)
			}

			//number := len(b.ciServiceContainers)
		}
		client.StopContainer(container.ID, 5)
		client.RemoveContainer(docker_client.RemoveContainerOptions{
			ID: container.ID,
		})
	}()

	// channel listening for errors while the
	// container is running async.
	errc := make(chan error, 1)
	containerc := make(chan *docker_client.Container, 1)
	go func() {
		// Options to fetch the stdout and stderr logs
		// by tailing the output.
		logOptsTail := &docker_client.LogsOptions{
			Follow:       true,
			Stdout:       true,
			Stderr:       true,
			Container:    container.ID,
			OutputStream: output,
			ErrorStream:  output,
		}

		// It's possible that the docker logs endpoint returns before the container
		// is done, we'll naively resume up to 5 times if when the logs unblocks
		// the container is still reported to be running.
		for attempts := 0; attempts < 5; attempts++ {
			if attempts > 0 {
				// When resuming the stream, only grab the last line when starting
				// the tailing.
				logOptsTail.Tail = "1"
			}

			// Blocks and waits for the container to finish
			// by streaming the logs (to /dev/null). Ideally
			// we could use the `wait` function instead
			err := client.Logs(*logOptsTail)
			if err != nil {
				log.Errorf("Error tailing %s. %s\n", cco.Config.Image, err)
				errc <- err
				return
			}

			info, err := client.InspectContainer(container.ID)
			if err != nil {
				log.Errorf("Error getting exit code for %s. %s\n", cco.Config.Image, err)
				errc <- err
				return
			}

			if info.State.Running != true {
				containerc <- info
				return
			}
		}

		errc <- errors.New("Maximum number of attempts made while tailing logs.")
	}()

	select {
	case info := <-containerc:
		err = CopyOutPutFiles(b.dockerManager, container.ID, outPutFiles, outPutPath)
		if nil != err {
			return container, err
		}
		return info, nil
	case err := <-errc:
		log.InfoWithFields("Run the container failed.", log.Fields{"config": cco})
		return container, err
	}
}

// CopyOutPutFiles copy output files from container
func CopyOutPutFiles(dockerManager *docker.Manager, cid string,
	outputs []string, outPutPath string) error {
	os.Mkdir(outPutPath, 0755)
	for _, output := range outputs {
		log.Infof("copy file(%s) to %s\n", output, outPutPath)
		err := CopyFromContainer(dockerManager.Client, cid, output, outPutPath)
		if nil != err {
			return err
		}
	}
	return nil
}

// CopyFromContainer copies from sourcefile in container to dstPath in host.
func CopyFromContainer(client *docker_client.Client,
	cid string, sourcefile string, dstPath string) error {
	var buf bytes.Buffer

	err := client.CopyFromContainer(docker_client.CopyFromContainerOptions{
		Container:    cid,
		Resource:     sourcefile,
		OutputStream: &buf,
	})

	if err != nil {
		log.Errorf("Error while copying from %s: %s\n", cid, err)
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
			os.Mkdir(dstPath+fileHead.Name, os.FileMode(fileHead.Mode))
		} else {
			var fileOutPut *os.File
			fileOutPut, err = os.OpenFile(dstPath+fileHead.Name,
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

// PreBuildByDockerfile prebuilds bin by Dockerfile.
func preBuildByDockerfile(output filebuffer.FileBuffer, dockerManager *docker.Manager, event *api.Event,
	dockerfilePath string, dockerfileName string, outPutFiles []string, outPutPath string) error {
	contextdir, ok := event.Data["context-dir"]
	if !ok {
		return fmt.Errorf("Unable to retrieve name and context directory from Event %#+v: %t",
			event, ok)
	}
	contextDir := contextdir.(string)
	if "" != dockerfilePath {
		contextDir = contextDir + "/" + dockerfilePath
	}

	if "" == dockerfileName {
		dockerfileName = "Dockerfile"
	}

	imageName := "output:" + event.Version.VersionID
	log.InfoWithFields("About to build docker image for prebuild.", log.Fields{"image": imageName})

	opt := docker_client.BuildImageOptions{
		Name:           imageName,
		Dockerfile:     dockerfileName,
		ContextDir:     contextDir,
		OutputStream:   output,
		RmTmpContainer: true,
		AuthConfigs:    dockerManager.GetAuthOpts(),
	}

	client := dockerManager.Client
	err := client.BuildImage(opt)
	if err != nil {
		log.Errorf("prebuild build images err: %v", err)
		return err
	}

	cco := &docker_client.CreateContainerOptions{
		Config: &docker_client.Config{
			Image: imageName,
		},
		HostConfig: &docker_client.HostConfig{},
	}

	container, errCreate := client.CreateContainer(*cco)
	if errCreate != nil {
		log.Errorf("prebuild create container err: %v", errCreate)
		return errCreate
	}
	errStart := client.StartContainer(container.ID, cco.HostConfig)
	if errStart != nil {
		log.Errorf("prebuild start container err: %v", errStart)
		return errStart
	}

	// Ensures the container is always stopped
	// and ready to be removed.
	defer func() {
		client.StopContainer(container.ID, 5)
		client.RemoveContainer(docker_client.RemoveContainerOptions{
			ID: container.ID,
		})
		client.RemoveImage(cco.Config.Image)
	}()

	errCopy := CopyOutPutFiles(dockerManager, container.ID, outPutFiles, outPutPath)
	if nil != errCopy {
		log.Errorf("prebuild copy from dockfile container err: %v", errCopy)
		return errCopy
	}

	return nil
}

// mockDotReference transfers ".:XXX" and "./XXX:YYY" to "$(pwd):XXX" and "$(pwd)/XXX:YYY"
// There are some implementations to slove dot reference.
// For example, the daocloud uses "before_script",
// run mv & cd to move the golang repo to the correct path; the drone just add a
// block called "clone", at the clone step, just clone the repo to the specific
// path.
// We use a trick, when the user add a ".:xxx" bind, cyclone will link the repo to
// the specific path, otherwise cyclone just links the repo to the same path as in
// the host.
func mockDotReference(b *Build, path string) (bool, string) {
	parts := strings.Split(path, ":")
	source := parts[0]
	if source == "." {
		// Mock ".:XXX"
		transferredPath := fmt.Sprintf("%s:%s", b.contextDir, parts[1])
		return true, transferredPath
	} else if source[0:2] == "./" {
		// Mock "./XXX:YYY"
		source = fmt.Sprintf("%s/%s", b.contextDir, source[2:])
		transferredPath := fmt.Sprintf("%s:%s", source, parts[1])
		return true, transferredPath
	}
	return false, path
}

// validatePath checks the path, if the path is not "b.contextDir/xxx/yyy",
// return false, else true.
func validatePath(b *Build, path string) bool {
	parts := strings.Split(path, ":")
	source := parts[0]
	if strings.HasPrefix(source, b.contextDir) && !strings.Contains(source, "..") {
		return true
	}
	return false
}
