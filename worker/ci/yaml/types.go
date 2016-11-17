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

// Package yaml is an implementation of yaml compiler.
package yaml

import (
	"fmt"
	"strings"

	"github.com/flynn/go-shlex"
	"gopkg.in/yaml.v2"
)

// Config is a typed representation of the
// Yaml configuration file.
type Config struct {
	Integration IntegrationStep `yaml:"integration"`
	PreBuild    PreBuildStep    `yaml:"pre_build"`
	Build       BuildStep       `yaml:"build"`
	PostBuild   BuildStep       `yaml:"post_build"`
	Deploy      DeployStep      `yaml:",inline"`
}

// Container is a typed representation of a
// docker step in the Yaml configuration file.
type Container struct {
	// Name is used by service block.
	Name           string        `yaml:"name"`
	Image          string        `yaml:"image"`
	Pull           bool          `yaml:"pull"`
	Privileged     bool          `yaml:"privileged"`
	Environment    MapEqualSlice `yaml:"environment"`
	Entrypoint     Command       `yaml:"entrypoint"`
	Command        Command       `yaml:"command"`
	ExtraHosts     []string      `yaml:"extra_hosts"`
	Volumes        []string      `yaml:"volumes"`
	Devices        []string      `yaml:"devices"`
	Net            string        `yaml:"net"`
	DNS            Stringorslice `yaml:"dns"`
	AuthConfig     AuthConfig    `yaml:"auth_config"`
	Memory         int64         `yaml:"mem_limit"`
	CPUSetCPUs     string        `yaml:"cpuset"`
	OomKillDisable bool          `yaml:"oom_kill_disable"`
}

// Containerslice is a slice of Containers with a custom
// Yaml unarmshal function to preserve ordering.
type Containerslice struct {
	parts []Container
}

// PreBuild is a typed representation of the pre_build
// step in the Yaml configuration file.
type PreBuild struct {
	DockerfilePath string `yaml:"context_dir"`
	DockerfileName string `yaml:"dockerfile_name"`

	Container `yaml:",inline"`

	Commands []string `yaml:"commands"`
	Outputs  []string `yaml:"outputs"`
}

// PreBuildStep holds the pre_build step configuration using a custom
// Yaml unarmshal function to preserve ordering.
type PreBuildStep struct {
	parts []PreBuild
}

// IntegrationStep holds the integration step configuration using
// a custom Yaml unmarshal function to preserve ordering.
type IntegrationStep struct {
	services Containerslice `yaml:"services"`
	build    Build
}

// Build is a typed representation of the build
// step in the Yaml configuration file.
type Build struct {
	DockerfilePath string `yaml:"context_dir"`
	DockerfileName string `yaml:"dockerfile_name"`

	Container `yaml:",inline"`

	Commands []string `yaml:"commands"`
}

// BuildStep holds the build step configuration using a custom
// Yaml unmarshal function to preserve ordering.
type BuildStep struct {
	parts []Build
}

// DeployStep deploys version to multiple applications.
type DeployStep struct {
	Applications []Application `yaml:"deploy"`
}

// Application contains information that helps locating a application, and
// deploying image to its containers.
type Application struct {
	ClusterType    string   `yaml:"type"` // kubernetes, caicloud_claas, mesos
	ClusterHost    string   `yaml:"host"`
	ClusterToken   string   `yaml:"token"`
	ClusterName    string   `yaml:"cluster"`
	NamespaceName  string   `yaml:"namespace"`
	DeploymentName string   `yaml:"deployment"`
	Containers     []string `yaml:"containers"`
}

// MapEqualSlice is the type for env map slice.
type MapEqualSlice struct {
	parts []string
}

// Command is the type for command field in yml config.
type Command struct {
	parts []string
}

// Stringorslice represents a string or an array of strings.
// TODO use docker/docker/pkg/stringutils.StrSlice once 1.9.x is released.
type Stringorslice struct {
	parts []string
}

// AuthConfig is the type for auth for Docker Image Registry.
type AuthConfig struct {
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	Email         string `yaml:"email"`
	RegistryToken string `yaml:"registry_token"`
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringType string
	err := unmarshal(&stringType)
	if err == nil {
		s.parts, err = shlex.Split(stringType)
		return err
	}

	var sliceType []string
	err = unmarshal(&sliceType)
	if err == nil {
		s.parts = sliceType
		return nil
	}

	return err
}

// Slice gets the parts of the Slice as a Slice of string.
func (s *Command) Slice() []string {
	return s.parts
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MapEqualSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&s.parts)
	if err == nil {
		return nil
	}

	var mapType map[string]string

	err = unmarshal(&mapType)
	if err != nil {
		return err
	}

	for k, v := range mapType {
		s.parts = append(s.parts, strings.Join([]string{k, v}, "="))
	}

	return nil
}

// Slice gets the parts of the MapEqualSlice as a Slice of string.
func (s *MapEqualSlice) Slice() []string {
	return s.parts
}

// MarshalYAML implements the Marshaller interface.
func (s Stringorslice) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []string
	err := unmarshal(&sliceType)
	if err == nil {
		s.parts = sliceType
		return nil
	}

	var stringType string
	err = unmarshal(&stringType)
	if err == nil {
		sliceType = make([]string, 0, 1)
		s.parts = append(sliceType, string(stringType))
		return nil
	}
	return err
}

// Len returns the number of parts of the Stringorslice.
func (s *Stringorslice) Len() int {
	if s == nil {
		return 0
	}
	return len(s.parts)
}

// Slice gets the parts of the StrSlice as a Slice of string.
func (s *Stringorslice) Slice() []string {
	if s == nil {
		return nil
	}
	return s.parts
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Containerslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// unmarshal the yaml into the generic
	// mapSlice type to preserve ordering.
	obj := yaml.MapSlice{}
	err := unmarshal(&obj)
	if err != nil {
		return err
	}

	// unarmshals each item in the mapSlice,
	// unmarshal and append to the slice.
	return unmarshalYaml(obj, func(key string, val []byte) error {
		ctr := Container{}
		err := yaml.Unmarshal(val, &ctr)
		if err != nil {
			return err
		}
		if len(ctr.Image) == 0 {
			ctr.Image = key
		}
		s.parts = append(s.parts, ctr)
		return nil
	})
}

// Slice gets the parts of the Containerslice as a Slice of string.
func (s *Containerslice) Slice() []Container {
	return s.parts
}

// IsPrebuildArray checks if there is a array of prebuild step.
func IsPrebuildArray(prebuild PreBuild) bool {
	if prebuild.Image == "" &&
		prebuild.DockerfilePath == "" &&
		prebuild.DockerfileName == "" {
		return true
	}
	return false
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *PreBuildStep) UnmarshalYAML(unmarshal func(interface{}) error) error {
	prebuild := PreBuild{}
	err := unmarshal(&prebuild)
	if err != nil {
		return err
	}

	if !IsPrebuildArray(prebuild) {
		s.parts = append(s.parts, prebuild)
		return nil
	}

	// unmarshal the yaml into the generic
	// mapSlice type to preserve ordering.
	obj := yaml.MapSlice{}
	if err := unmarshal(&obj); err != nil {
		return err
	}

	// unarmshals each item in the mapSlice,
	// unmarshal and append to the slice.
	return unmarshalYaml(obj, func(key string, val []byte) error {
		prebuild := PreBuild{}
		err := yaml.Unmarshal(val, &prebuild)
		if err != nil {
			return err
		}
		s.parts = append(s.parts, prebuild)
		return nil
	})
}

// Slice gets the parts of the PreBuildStep as a Slice of string.
func (s *PreBuildStep) Slice() []PreBuild {
	return s.parts
}

// UnmarshalYAML implements the Unmarshaller interface.
func (i *IntegrationStep) UnmarshalYAML(unmarshal func(interface{}) error) error {
	slice := yaml.MapSlice{}
	if err := unmarshal(&slice); err != nil {
		return err
	}

	// Ugly code. I don't know how to parse the yaml elegantly. It should parse the
	// yaml just like a compiler's parser.
	var sliceWithoutService = slice
	for index, v := range slice {
		if v.Key == "services" {
			// Parse the services block.
			yml, err := yaml.Marshal(&v.Value)
			if err != nil {
				return err
			}
			servicesBuf := yaml.MapSlice{}
			services := Containerslice{}
			if err := yaml.Unmarshal(yml, &servicesBuf); err != nil {
				return err
			}
			for _, service := range servicesBuf {
				yml, err := yaml.Marshal(&service.Value)
				if err != nil {
					return err
				}
				ctr := Container{}
				if err := yaml.Unmarshal(yml, &ctr); err != nil {
					return err
				}
				ctr.Name = service.Key.(string)
				if len(ctr.Image) == 0 {
					ctr.Image = service.Key.(string)
				}
				services.parts = append(services.parts, ctr)
			}
			i.services = services

			// Remove the services block from slice.
			// Notice: The services block must be the first block.
			sliceWithoutService = slice[index+1:]
		}
	}

	// Parse the build block.
	yml, err := yaml.Marshal(sliceWithoutService)
	if err != nil {
		return err
	}
	build := Build{}
	if err := yaml.Unmarshal(yml, &build); err != nil {
		return err
	}
	i.build = build
	return nil
}

// ServiceSlice gets the parts of service in the IntegrationStep as a Slice of string.
func (i *IntegrationStep) ServiceSlice() []Container {
	return i.services.parts
}

// Build gets the build of IntegrationStep.
func (i *IntegrationStep) Build() Build {
	return i.build
}

// IsBuildArray checks if there is a build array.
func IsBuildArray(build Build) bool {
	if build.Image == "" &&
		build.DockerfilePath == "" &&
		build.DockerfileName == "" {
		return true
	}
	return false
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *BuildStep) UnmarshalYAML(unmarshal func(interface{}) error) error {
	build := Build{}
	err := unmarshal(&build)
	if err != nil {
		return err
	}
	if !IsBuildArray(build) {
		s.parts = append(s.parts, build)
		return nil
	}

	// unmarshal the yaml into the generic
	// mapSlice type to preserve ordering.
	obj := yaml.MapSlice{}
	if err := unmarshal(&obj); err != nil {
		return err
	}

	// unarmshals each item in the mapSlice,
	// unmarshal and append to the slice.
	return unmarshalYaml(obj, func(key string, val []byte) error {
		build := Build{}
		err := yaml.Unmarshal(val, &build)
		if err != nil {
			return err
		}
		s.parts = append(s.parts, build)
		return nil
	})
}

// Slice gets the parts of the BuildStep as a Slice of string.
func (s *BuildStep) Slice() []Build {
	return s.parts
}

// emitter defines the callback function used for
// generic yaml parsing. It emits back a raw byte
// slice for custom unmarshalling into a structure.
type unmarshal func(string, []byte) error

// unmarshalYaml is a helper function that removes
// some of the boilerplate from unmarshalling
// complex map slices.
func unmarshalYaml(v yaml.MapSlice, emit unmarshal) error {
	for _, vv := range v {
		// re-marshal the interface{} back to
		// a raw yaml value
		val, err := yaml.Marshal(&vv.Value)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%v", vv.Key)

		// unmarshal the raw value using the
		// callback function.
		if err := emit(key, val); err != nil {
			return err
		}
	}
	return nil
}
