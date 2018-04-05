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

package api

import (
	"fmt"
)

// GetGitSource gets git source according from code source according to the SCM type.
func GetGitSource(codeSource *CodeSource) (*GitSource, error) {
	scmType := codeSource.Type
	var gitSource *GitSource
	switch scmType {
	case Github:
		gitSource = codeSource.Github
	case Gitlab:
		gitSource = codeSource.Gitlab
	case SVN:
		gitSource = codeSource.SVN
	default:
		return nil, fmt.Errorf("SCM type %s is not supported", scmType)
	}

	return gitSource, nil
}
