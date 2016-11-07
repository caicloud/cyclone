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

// NodeType identifies the type of a parse tree node.
type NodeType uint

// Type returns itself and provides an easy default
// implementation for embedding in a Node. Embedded
// in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeList NodeType = 1 << iota
	NodeIntegration
	NodePreBuild
	NodeBuild
	NodePostBuild
	NodeService
	NodeDeploy
)

// Nodes.

// Node is the interface of a generic node.
type Node interface {
	Type() NodeType
}

// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Nodes []Node // nodes executed in lexical order.
}

// DockerNode represents a Docker container that
// should be launched as part of the build process.
type DockerNode struct {
	// NodeType defines the type of the DockerNode.
	NodeType

	//
	// Fields only for ServiceNode
	//

	// Name of the node, when the node is running, the name is used to
	// service discovery.
	Name string

	//
	// Fields only for BuildNode and PreBuildNode
	//

	// Commands is only for BuildNode and PreBuildNode, to support command list.
	Commands []string

	//
	// Common fields
	//

	Image          string
	Pull           bool
	Privileged     bool
	Environment    []string
	Entrypoint     []string
	Command        []string
	Volumes        []string
	Devices        []string
	ExtraHosts     []string
	Net            string
	DNS            []string
	AuthConfig     yaml.AuthConfig
	Memory         int64
	CPUSetCPUs     string
	OomKillDisable bool
	Outputs        []string
	DockerfilePath string
	DockerfileName string
	Vargs          map[string]interface{}
}

// DeployNode is the type for deploy section in yml.
type DeployNode struct {
	// NodeType defines the type of the DockerNode.
	NodeType

	Applications []yaml.Application
}

func newDockerNode(typ NodeType, c yaml.Container) *DockerNode {
	return &DockerNode{
		NodeType:       typ,
		Image:          c.Image,
		Pull:           c.Pull,
		Privileged:     c.Privileged,
		Environment:    c.Environment.Slice(),
		Entrypoint:     c.Entrypoint.Slice(),
		Command:        c.Command.Slice(),
		Volumes:        c.Volumes,
		Devices:        c.Devices,
		ExtraHosts:     c.ExtraHosts,
		Net:            c.Net,
		DNS:            c.DNS.Slice(),
		AuthConfig:     c.AuthConfig,
		Memory:         c.Memory,
		CPUSetCPUs:     c.CPUSetCPUs,
		OomKillDisable: c.OomKillDisable,
	}
}

// newServiceNode returns a new ServiceNode, different from the other nodes,
// ServiceNode has the name for service discovery.
func newServiceNode(c yaml.Container, name string) *DockerNode {
	node := newDockerNode(NodeService, c)
	node.Name = name
	return node
}

// newBuildNode returns a new BuildNode, different from the other nodes,
// BuildNode has the commands for running.
func newBuildNode(typ NodeType, b yaml.Build) *DockerNode {
	node := newDockerNode(typ, b.Container)
	node.DockerfilePath = b.DockerfilePath
	node.DockerfileName = b.DockerfileName
	node.Commands = b.Commands
	return node
}

func newPreBuildNode(typ NodeType, b yaml.PreBuild) *DockerNode {
	node := newDockerNode(typ, b.Container)
	node.Commands = b.Commands
	node.Outputs = b.Outputs
	node.DockerfilePath = b.DockerfilePath
	node.DockerfileName = b.DockerfileName
	return node
}

// newDeployNode returns a new DeployNode. A DeployNode represents an action deploying
// a version to an application.
func newDeployNode(d yaml.DeployStep) *DeployNode {
	return &DeployNode{
		NodeType:     NodeDeploy,
		Applications: d.Applications,
	}
}

// Append appends a node to the list.
func (l *ListNode) append(n ...Node) {
	l.Nodes = append(l.Nodes, n...)
}
