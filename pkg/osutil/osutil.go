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

package osutil

import (
	"os"
	"os/user"
	"strconv"

	"github.com/caicloud/cyclone/pkg/log"
)

// GetHomeDir gets current user's home directory.
func GetHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir
}

// GetStringEnv get evironment value of 'name', and return provided
// default value if not found.
func GetStringEnv(name, def string) string {
	val := os.Getenv(name)
	if val == "" {
		log.Infof("Env variant %s not found, using default value: %s", name, def)
		return def
	}
	log.Infof("Env variant %s found, using env value: %s", name, val)
	return val
}

// GetIntEnv get evironment value of 'name', and return provided
// default value if not found.
func GetIntEnv(name string, def int) int {
	val, err := strconv.Atoi(os.Getenv(name))
	if err != nil {
		log.Infof("Env variant %s is not numeric, using default value: %d", name, def)
		return def
	}
	log.Infof("Env variant %s found, using env value: %d", name, val)
	return val
}

// GetFloat64Env get evironment value of 'name', and return provided
// default value if not found.
func GetFloat64Env(name string, def float64) float64 {
	val, err := strconv.ParseFloat(os.Getenv(name), 64)
	if err != nil {
		log.Infof("Env variant %s is not numeric, using default value: %f", name, def)
		return def
	}
	log.Infof("Env variant %s found, using env value: %f", name, val)
	return val
}

// OpenFile opens file from 'name', and create one if not exist.
func OpenFile(fileName string, flag int, perm os.FileMode) (*os.File, error) {
	var file *os.File
	var err error

	file, err = os.OpenFile(fileName, flag, perm)
	if err != nil && os.IsNotExist(err) {
		file, err = os.Create(fileName)
		if err != nil {
			return nil, err
		}
	}

	return file, err
}

// IsFileExists returns true if the file exists.
func IsFileExists(fileName string) bool {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return false
	}
	return true
}
