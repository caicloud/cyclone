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

package utils

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

var goPaths []string

func init() {
	goPaths = strings.Split(build.Default.GOPATH, string(os.PathListSeparator))
	for i, goPath := range goPaths {
		goPaths[i] = filepath.Join(goPath, "src")
	}
}

// GoPath finds the go path for specified path. path must be a absolute file path.
func GoPath(path string) (goPath string, absPath string, err error) {
	absPath, err = filepath.Abs(path)
	if err != nil {
		return "", "", err
	}
	for _, goPath := range goPaths {
		if strings.HasPrefix(absPath, goPath) {
			return goPath, absPath, nil
		}
	}
	return "", "", fmt.Errorf("%s is not in GOPATH", absPath)
}

// PackageForPath gets package for a path.
func PackageForPath(path string) (string, error) {
	goPath, absPath, err := GoPath(path)
	if err != nil {
		return "", err
	}
	return strings.Trim(strings.TrimPrefix(absPath, goPath), "/"), nil
}
