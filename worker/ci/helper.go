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
	"io/ioutil"

	"github.com/caicloud/cyclone/worker/ci/parser"
)

// fetchAndParseYaml fetches caicloud.yml from the repo, then
// parse yaml to execution Tree.
func fetchAndParseYaml(directFilePath string) (*parser.Tree, error) {
	raw, err := ioutil.ReadFile(directFilePath)
	if err != nil {
		return nil, err
	}
	tree, err := parser.Parse(raw)
	if err != nil {
		return nil, err
	}

	return tree, nil
}
