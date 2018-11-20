/*
Copyright 2018 Caicloud Authors

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

package swagger

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Contact describes a project maintainer.
type Contact struct {
	// Name is maintainer's name.
	Name string
	// Email is maintainer's email.
	Email string
	// Description describes the dutis of this maintainer.
	Description string
}

// Version describes an API version.
type Version struct {
	// Name is version number. SemVer is recommended.
	Name string
	// Description describes this version.
	Description string
	// Schemes override the same field in config.
	Schemes []string
	// Hosts override the same field in config.
	Hosts []string
	// Contacts override the same field in config.
	Contacts []Contact
	// PathRules contains a list of regexp rules to match path.
	PathRules []string
}

// Config describes configurations of swagger
type Config struct {
	// Project is project name.
	Project string
	// Description describes this project.
	Description string
	// Schemes contains all schemes of APIs.
	// Values must be in "http", "https", "ws", "wss".
	Schemes []string
	// Hosts contains the host address to access APIs.
	// It's values can be "domain:port" or "ip:port".
	Hosts []string
	// Contacts contains maintainers of this project.
	Contacts []Contact
	// Versions describes API versions.
	Versions []Version
}

// LoadConfig loads config from yaml file.
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}
