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

package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// DefaultProjectFileName is the default project file name of a nirvana project.
const DefaultProjectFileName = "nirvana.yaml"

// Contact describes a project maintainer.
type Contact struct {
	// Name is maintainer's name.
	Name string `yaml:"name"`
	// Email is maintainer's email.
	Email string `yaml:"email"`
	// Description describes the dutis of this maintainer.
	Description string `yaml:"description"`
}

// PathRule describes a path rule.
type PathRule struct {
	// Prefix indicates a prefix of path.
	Prefix string `yaml:"prefix"`
	// Regexp is a regular expression. Prefix and Regexp are mutually exclusive.
	// If a prefix is specified, this field will be ignored.
	Regexp string `yaml:"regexp"`
	regexp *regexp.Regexp
	// Replacement is used to replace path parts matched by Prefix or Regexp.
	// In a concrete case, such as we need to export our apis with a gateway. The gateway
	// redirects "/component-name/v1" to "/api/v1". Then we should generate api docs and
	// clients with prefix "/component-name/v1". Then set Prefix to "/api/v1" and
	// Replacement to "/component-name/v1".
	Replacement string `yaml:"replacement"`
}

// Check checks if a path is matched by this rule.
func (r *PathRule) Check(path string) bool {
	if r.Prefix != "" {
		return strings.HasPrefix(path, r.Prefix)
	}
	if r.Regexp == "" {
		return false
	}
	if r.regexp == nil {
		r.regexp = regexp.MustCompile(r.Regexp)
	}
	return r.regexp.MatchString(path)
}

// Replace replaces a path with replacement. If path is not matched, returns empty string.
func (r *PathRule) Replace(path string) string {
	if !r.Check(path) {
		return ""
	}
	if r.Replacement == "" {
		return path
	}
	if r.Prefix != "" {
		return r.Replacement + path[len(r.Prefix):]
	}
	return r.regexp.ReplaceAllString(path, r.Replacement)
}

// Validate validates this rule.
func (r *PathRule) Validate() error {
	if r.Regexp == "" {
		return nil
	}
	if r.regexp != nil {
		return nil
	}
	exp, err := regexp.Compile(r.Regexp)
	if err != nil {
		return err
	}
	r.regexp = exp
	return nil
}

// Version describes an API version.
type Version struct {
	// Name is version number. SemVer is recommended.
	Name string `yaml:"name"`
	// Description describes this version.
	Description string `yaml:"description"`
	// Schemes overrides the same field in config.
	Schemes []string `yaml:"schemes"`
	// Hosts overrides the same field in config.
	Hosts []string `yaml:"hosts"`
	// Contacts overrides the same field in config.
	Contacts []Contact `yaml:"contacts"`
	// PathRules contains a list of regexp rules to match path.
	PathRules []PathRule `yaml:"rules"`
}

// Validate validates this version.
func (v *Version) Validate() error {
	for _, r := range v.PathRules {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Config describes configurations of a project.
type Config struct {
	// Root is the directory of this config.
	Root string `yaml:"-"`
	// Project is project name.
	Project string `yaml:"project"`
	// Description describes this project.
	Description string `yaml:"description"`
	// Schemes contains all schemes of APIs.
	// Values must be in "http", "https", "ws", "wss".
	Schemes []string `yaml:"schemes"`
	// Hosts contains the host address to access APIs.
	// It's values can be "domain:port" or "ip:port".
	Hosts []string `yaml:"hosts"`
	// Contacts contains maintainers of this project.
	Contacts []Contact `yaml:"contacts"`
	// Versions describes API versions.
	Versions []Version `yaml:"versions"`
}

// Validate validates this config.
func (c *Config) Validate() error {
	for _, v := range c.Versions {
		if err := v.Validate(); err != nil {
			return err
		}
	}
	return nil
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
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	config.Root = filepath.Dir(absPath)
	return config, config.Validate()
}

// LoadDefaultProjectFile finds the path of nirvana.yaml and loads it.
// It will find the directory itself and its parents recursively.
func LoadDefaultProjectFile(directory string) (*Config, error) {
	path, err := FindDefaultProjectFile(directory)
	if err != nil {
		return nil, err
	}
	return LoadConfig(path)
}

// FindDefaultProjectFile finds the path of nirvana.yaml.
// It will find the path itself and its parents recursively.
func FindDefaultProjectFile(directory string) (string, error) {
	return FindProjectFile(directory, DefaultProjectFileName)
}

// FindProjectFile finds the path of project file.
// It will find the path itself and its parents recursively.
func FindProjectFile(directory string, fileName string) (string, error) {
	goPath, absPath, err := GoPath(directory)
	if err != nil {
		return "", err
	}
	for len(absPath) > len(goPath) {
		path := filepath.Join(absPath, fileName)
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			return path, nil
		}
		absPath = filepath.Dir(absPath)
	}
	return "", fmt.Errorf("can't find nirvana.yaml")
}
