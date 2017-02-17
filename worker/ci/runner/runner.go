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
	"fmt"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/worker/ci/parser"
	"github.com/caicloud/cyclone/worker/clair"
	steplog "github.com/caicloud/cyclone/worker/log"
)

// BuildStatus represents the type for status of build.
type BuildStatus uint

// Build is a typed representation of a build job.
type Build struct {
	contextDir          string
	event               *api.Event
	dockerManager       *docker.Manager
	tree                *parser.Tree
	flags               parser.NodeType
	ciServiceContainers []string
	status              BuildStatus
}

const (
	pushImageSuccess BuildStatus = 1 << iota
)

// Load loads the tree to the build job.
func Load(contextDir string, event *api.Event, dockerManager *docker.Manager, tree *parser.Tree) *Build {
	return &Build{
		contextDir:    contextDir,
		event:         event,
		dockerManager: dockerManager,
		tree:          tree,
	}
}

// Setup environments.
func (b *Build) Setup() error {
	return nil
}

// Teardown tears down the build.
func (b *Build) Teardown() error {
	return nil
}

// RunNode walks through the tree, run the build job.
func (b *Build) RunNode(flags parser.NodeType) error {
	b.flags = flags
	return b.walk(b.tree.Root)
}

// walk through the tree, recursively.
func (b *Build) walk(node parser.Node) (err error) {
	outPutPath := b.event.Data["context-dir"].(string) + "/"
	switch node := node.(type) {
	case *parser.ListNode:
		for _, node := range node.Nodes {
			err = b.walk(node)
			if err != nil {
				return err
			}
		}

	case *parser.DockerNode:
		if shouldSkip(b.flags, node.NodeType) {
			break
		}
		if isLackOfCriticalConfig(node) {
			break
		}

		switch node.Type() {
		case parser.NodeService:
			// Record image name
			createContainerOptions := toServiceContainerConfig(node, b)

			// Run the docker container.
			container, err := start(b, createContainerOptions)
			if err != nil {
				return err
			}
			// Set the container ID, to stop and remove the containers at
			// post hook function.
			b.ciServiceContainers = append(b.ciServiceContainers, container.ID)

		case parser.NodeIntegration:
			// Record image name
			createContainerOptions := toBuildContainerConfig(node, b, parser.NodeIntegration)
			// Encode the commands to one line script.
			Encode(createContainerOptions, node)

			// Run the docker container.
			container, err := run(b, createContainerOptions, node.Outputs, outPutPath, node.Type(), steplog.Output)
			if err != nil {
				return err
			}

			// Check the exitcode from container
			if container.State.ExitCode != 0 {
				return fmt.Errorf("container meets error")
			}

		case parser.NodeBuild:
			log.Info("Build with Dockerfile path: ", node.DockerfilePath,
				" ", node.DockerfileName)
			if err := b.dockerManager.BuildImageSpecifyDockerfile(b.event,
				node.DockerfilePath, node.DockerfileName, steplog.Output); err != nil {
				return err
			}

		case parser.NodePreBuild:
			// Record image name
			steplog.InsertStepLog(b.event, steplog.PreBuild, steplog.Start, nil)
			if "" != node.DockerfilePath || "" != node.DockerfileName {
				log.Info("Pre_build with Dockerfile path: ", node.DockerfilePath,
					" ", node.DockerfileName)
				errDockerfile := preBuildByDockerfile(steplog.Output, b.dockerManager,
					b.event, node.DockerfilePath, node.DockerfileName, node.Outputs,
					outPutPath)
				if nil != errDockerfile {
					steplog.InsertStepLog(b.event, steplog.PreBuild, steplog.Stop, errDockerfile)
					return errDockerfile
				}
			} else {
				createContainerOptions := toBuildContainerConfig(node, b, parser.NodePreBuild)
				// Encode the commands to one line script.
				Encode(createContainerOptions, node)

				// Run the docker container.
				container, err := run(b, createContainerOptions, node.Outputs, outPutPath, node.Type(), steplog.Output)
				if err != nil {
					steplog.InsertStepLog(b.event, steplog.PreBuild, steplog.Stop, err)
					return err
				}
				// Check the exitcode from container
				if container.State.ExitCode != 0 {
					errExit := fmt.Errorf("container meets error")
					steplog.InsertStepLog(b.event, steplog.PreBuild, steplog.Stop, errExit)
					return errExit
				}
			}
			steplog.InsertStepLog(b.event, steplog.PreBuild, steplog.Finish, nil)

		case parser.NodePostBuild:
			// Record image name
			steplog.InsertStepLog(b.event, steplog.PostBuild, steplog.Start, nil)
			// create the container with default network.
			createContainerOptions := toBuildContainerConfig(node, b, parser.NodePostBuild)
			// Encode the commands to one line script.
			Encode(createContainerOptions, node)

			// Run the docker container.
			container, err := run(b, createContainerOptions, node.Outputs, outPutPath, node.Type(), steplog.Output)
			if err != nil {
				steplog.InsertStepLog(b.event, steplog.PostBuild, steplog.Stop, err)
				return err
			}
			// Check the exitcode from container
			if container.State.ExitCode != 0 {
				errExit := fmt.Errorf("container meets error")
				steplog.InsertStepLog(b.event, steplog.PostBuild, steplog.Stop, errExit)
				return errExit
			}
			steplog.InsertStepLog(b.event, steplog.PostBuild, steplog.Finish, nil)
		}
	}
	return nil
}

// PublishImage publish image to registry.
func (b *Build) PublishImage() (err error) {
	steplog.InsertStepLog(b.event, steplog.PushImage, steplog.Start, nil)
	if err := b.dockerManager.PushImage(b.event, steplog.Output); err != nil {
		steplog.InsertStepLog(b.event, steplog.PushImage, steplog.Stop, err)
		return err
	}

	// Now image is pushed to registry successfully.
	b.status |= pushImageSuccess

	clair.Analysis(b.event, b.dockerManager)
	steplog.InsertStepLog(b.event, steplog.PushImage, steplog.Finish, nil)
	return nil
}

// IsPushImageSuccess gets if image is pushed successfully.
func (b *Build) IsPushImageSuccess() bool {
	if b.event.Version.Operation == api.DeployOperation {
		return true
	}
	return (b.status & pushImageSuccess) != 0
}

// shouldSkip is a helper function that returns true if
// node execution should be skipped. This happens when
// the build is executed for a subset of build steps.
func shouldSkip(flags parser.NodeType, nodeType parser.NodeType) bool {
	return flags != 0 && flags&nodeType == 0
}

// isLackOfCriticalConfig func checks whether there is lack of critical config.
func isLackOfCriticalConfig(node *parser.DockerNode) bool {
	// build node didn't neet any critical config
	if parser.NodeBuild != node.Type() && 0 == len(node.Image) &&
		0 == len(node.DockerfilePath) &&
		0 == len(node.DockerfileName) {
		return true
	}
	return false
}

// GetEvent returns the event.
func (b Build) GetEvent() *api.Event {
	return b.event
}
