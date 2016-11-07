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

// Package parser is an implementation of yaml parser.
package parser

import (
	"github.com/caicloud/cyclone/worker/ci/yaml"
)

// Tree is the representation of a parsed build
// configuration Yaml file.
type Tree struct {
	Root *ListNode
	// Now deploy is independent with CI.
	DeployConfig *DeployNode

	NumberContatiner int
}

// newTree allocates a new parse tree.
func newTree() *Tree {
	return &Tree{
		Root: &ListNode{NodeType: NodeList},
	}
}

// ParseString parses the Yaml build definition file
// and returns an execution Tree.
func ParseString(raw string) (*Tree, error) {
	conf, err := yaml.ParseString(raw)
	if err != nil {
		return nil, err
	}
	return Load(conf)
}

// Parse parses the Yaml build definition file
// and returns an execution Tree.
func Parse(raw []byte) (*Tree, error) {
	conf, err := yaml.Parse(raw)
	if err != nil {
		return nil, err
	}
	return Load(conf)
}

// Load loads the Yaml build definition structure
// and returns an execution Tree.
func Load(conf *yaml.Config) (*Tree, error) {
	var tree = newTree()
	var err error

	// append the prebuild step to execution Tree.
	err = tree.appendPreBuild(conf.PreBuild.Slice())
	if err != nil {
		return nil, err
	}

	// append the build step to execution Tree.
	err = tree.appendBuild(conf.Build.Slice())
	if err != nil {
		return nil, err
	}

	// append the service map to execution Tree.
	err = tree.appendServices(conf.Integration.ServiceSlice())
	if err != nil {
		return nil, err
	}

	// append the integration step to execution Tree.
	err = tree.appendIntegration(conf.Integration.Build())
	if err != nil {
		return nil, err
	}

	// append the postbuild step to execution Tree.
	err = tree.appendPostBuild(conf.PostBuild.Slice())
	if err != nil {
		return nil, err
	}

	// append deploy step to execution tree
	err = tree.appendDeploy(conf.Deploy)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func (t *Tree) appendPreBuild(prebuilds []yaml.PreBuild) error {
	for _, prebuild := range prebuilds {
		node := newPreBuildNode(NodePreBuild, prebuild)
		t.Root.append(node)
	}
	return nil
}

// appendServices appends the services to the root, the services have their
// name for service discovery.
func (t *Tree) appendServices(services []yaml.Container) error {
	for _, service := range services {
		node := newServiceNode(service, service.Name)
		t.Root.append(node)
	}
	return nil
}

// appendIntegration appends the integration node to the root.
func (t *Tree) appendIntegration(build yaml.Build) error {
	node := newBuildNode(NodeIntegration, build)
	// TODO: add the filter node just like drone,
	//       to allow cyclone skip some steps in CI.
	t.Root.append(node)
	return nil
}

// appendBuild appends the build node to the root.
func (t *Tree) appendBuild(builds []yaml.Build) error {
	for _, build := range builds {
		node := newBuildNode(NodeBuild, build)
		t.Root.append(node)
	}
	return nil
}

// appendPostBuild appends the postbuild hook node to the root.
func (t *Tree) appendPostBuild(builds []yaml.Build) error {
	for _, build := range builds {
		node := newBuildNode(NodePostBuild, build)
		t.Root.append(node)
	}
	return nil
}

// appendDeploy appends the deploy node to the root
func (t *Tree) appendDeploy(d yaml.DeployStep) error {
	node := newDeployNode(d)
	t.DeployConfig = node
	return nil
}

func max(big int, args ...int) int {
	for _, v := range args {
		if big < v {
			big = v
		}
	}
	return big
}
