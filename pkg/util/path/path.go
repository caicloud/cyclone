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

package pathutil

import (
	"os"
	"path"
)

// EnsureParentDir makes sure parent directory exists for the given directory;
// if not, it will be created. 'name' can be either a file name or a directory
// name. If it's a directory, the directory will NOT be created, the function
// only makes sure parent directory exists.
func EnsureParentDir(name string, perm os.FileMode) error {
	_, err := os.Stat(name)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(path.Dir(name), perm)
}
