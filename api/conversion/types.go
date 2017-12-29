/*
Copyright 2017 caicloud authors. All rights reserved.

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

package conversion

// Config represents the config of caicloud YAML.
type Config struct {
	PreBuild    *PreBuild    `yaml:"pre_build,omitempty"`
	Build       *Build       `yaml:"build,omitempty"`
	Integration *Integration `yaml:"integration,omitempty"`
}

// PreBuild represents the config of preBuild in caicloud YAML.
type PreBuild struct {
	Image       string   `yaml:"image"`
	Environment []string `yaml:"environment,omitempty"`
	Commands    []string `yaml:"commands"`
	Outputs     []string `yaml:"outputs"`
}

// Build represents the config of Build in caicloud YAML.
type Build struct {
	ContextDir     string `yaml:"context_dir,omitempty"`
	DockerfileName string `yaml:"dockerfile_name,omitempty"`
}

// Integration represents the config of Integration in caicloud YAML.
type Integration struct {
	Services    map[string]Service `yaml:"services,omitempty"`
	Image 		string 			   `yaml:"image,omitempty"`
	Environment []string           `yaml:"environment,omitempty"`
	Commands    []string           `yaml:"commands,omitempty"`
}

// Service represents the dependent service in integration test.
type Service struct {
	// Image represents the image of service, can not be empty.
	Image       string   `yaml:"image"`
	Environment []string `yaml:"environment,omitempty"`
	Commands    []string `yaml:"commmands,omitempty"`
}
