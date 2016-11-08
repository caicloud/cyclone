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

import "gopkg.in/yaml.v2"

// Parse parses a Yaml configuration file.
func Parse(in []byte) (*Config, error) {
	c := Config{}
	e := yaml.Unmarshal(in, &c)
	return &c, e
}

// ParseString parses a Yaml configuration file
// in string format.
func ParseString(in string) (*Config, error) {
	return Parse([]byte(in))
}
