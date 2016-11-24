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

// Package ci is an implementation of ci manager.
package ci

import (
	"errors"
	"fmt"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/docker"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/worker/ci/parser"
	"github.com/caicloud/cyclone/worker/ci/runner"
	steplog "github.com/caicloud/cyclone/worker/log"
)

const (
	// yamlName defines the name of CI config file.
	yamlName = "caicloud.yml"
)

var (
	// List of errors about yaml.

	// ErrYamlNotExist defines a error caused by no yaml found.
	ErrYamlNotExist = errors.New("yaml not found")
	// ErrCustomYamlNotExist defines a error caused by custom yaml not found.
	ErrCustomYamlNotExist = errors.New("custom yaml not found")
)

// Manager manages all CI operations.
type Manager struct {
	// client creates docker containers to run build jobs.
	dockerManager *docker.Manager
}

// NewManager creates a new CI manager.
func NewManager(dockerManager *docker.Manager) (*Manager, error) {
	if dockerManager != nil {
		return &Manager{
			dockerManager: dockerManager,
		}, nil
	}
	return nil, fmt.Errorf("docekermanager is nil , can't new Manager")
}

// ExecIntegration executes the 'integration' section in yaml file
func (cm *Manager) ExecIntegration(r *runner.Build) error {
	var err error
	log.Info("About to run integration event.")
	steplog.InsertStepLog(r.GetEvent(), steplog.Integration, steplog.Start, nil)
	err = r.RunNode(parser.NodeService)
	if err != nil {
		log.Info("run integration event failed.")
		steplog.InsertStepLog(r.GetEvent(), steplog.Integration, steplog.Stop, err)
		return err
	}
	err = r.RunNode(parser.NodeIntegration)
	if err != nil {
		log.Info("run integration event failed.")
		steplog.InsertStepLog(r.GetEvent(), steplog.Integration, steplog.Stop, err)
		return err
	}
	steplog.InsertStepLog(r.GetEvent(), steplog.Integration, steplog.Finish, nil)
	return nil
}

// ExecPreBuild executes the 'pre build' section in yaml file
func (cm *Manager) ExecPreBuild(r *runner.Build) error {
	log.Info("About to run prebuild event.")
	err := r.RunNode(parser.NodePreBuild)
	if err != nil {
		log.Info("Run pre build failed.")
		return err
	}
	return nil
}

// ExecBuild executes the 'build' section in yaml file
func (cm *Manager) ExecBuild(r *runner.Build) error {
	log.Info("About to run build event.")
	err := r.RunNode(parser.NodeBuild)
	if err != nil {
		log.Info("Run build failed.")
		return err
	}
	return nil
}

// ExecPublish executes publish image to registry.
func (cm *Manager) ExecPublish(r *runner.Build) error {
	log.Info("About to run publish event.")

	err := r.PublishImage()
	if err != nil {
		log.Info("Run publish failed.")
		return err
	}
	return nil
}

// ExecPostBuild executes the 'postbuild' section in yaml file
func (cm *Manager) ExecPostBuild(r *runner.Build) error {
	log.Info("About to run post build event.")
	err := r.RunNode(parser.NodePostBuild)
	if err != nil {
		log.Info("run post post build failed.")
		return err
	}
	return nil
}

// Parse the yaml to node tree
func (cm *Manager) Parse(event *api.Event) (*parser.Tree, error) {
	var directFilePath string
	var errYaml error
	steplog.InsertStepLog(event, steplog.ParseYaml, steplog.Start, nil)
	contextdir, ok := event.Data["context-dir"]
	if !ok {
		err := fmt.Errorf("Unable to retrieve name and context directory from Event %#+v: %t",
			event, ok)
		steplog.InsertStepLog(event, steplog.ParseYaml, steplog.Stop, err)
		return nil, err
	}
	contextDir := contextdir.(string)
	errYaml = ErrYamlNotExist
	// Use the filename of config file if given.
	if event.Service.YAMLConfigName != "" {
		directFilePath = fmt.Sprintf("%s/%s", contextDir, event.Service.YAMLConfigName)
		// Set the err to ErrCustomYamlNotExist.
		errYaml = ErrCustomYamlNotExist
	} else {
		directFilePath = fmt.Sprintf("%s/%s", contextDir, yamlName)
	}
	if osutil.IsFileExists(directFilePath) != true {
		fmt.Fprintf(steplog.Output, "Error: %v\n", errYaml)
		steplog.InsertStepLog(event, steplog.ParseYaml, steplog.Stop, errYaml)
		return nil, errYaml
	}

	// Fetch and parse caicloud.yml from the repo.
	tree, err := fetchAndParseYaml(directFilePath)
	if err != nil {
		steplog.InsertStepLog(event, steplog.ParseYaml, steplog.Stop, err)
		return nil, err
	}
	steplog.InsertStepLog(event, steplog.ParseYaml, steplog.Finish, nil)
	return tree, nil
}

// LoadTree loads the tree and returns the runner.Build.
func (cm *Manager) LoadTree(event *api.Event, tree *parser.Tree) (*runner.Build, error) {
	contextDir, ok := event.Data["context-dir"]
	if !ok {
		return nil, fmt.Errorf("Unable to retrieve name and context directory from Event %#+v: %t",
			event, ok)
	}
	return runner.Load(contextDir.(string), event, cm.dockerManager, tree), nil
}

// Setup sets up the networks and volumes
func (cm *Manager) Setup(r *runner.Build) error {
	return r.Setup()
}

// Teardown the networks and volumes
func (cm *Manager) Teardown(r *runner.Build) error {
	return r.Teardown()
}
