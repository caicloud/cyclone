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

package rest

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	restful "github.com/emicklei/go-restful"
)

const (
	// DockerfilePath dockerfile templates path
	DockerfilePath = "/templates/dockerfile"
	// YamlPath dockerfile templates path
	YamlPath = "/templates/yaml"
)

var (
	dockerfiles = make(map[string]string)
	yamlfiles   = make(map[string]string)
)

// walkDockerfiles is a WalkFunc to deal with dockerfile templates
func walkDockerfiles(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		// skip dir
		return nil
	}

	filename := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	dockerfiles[filename] = string(b)
	return nil
}

// walkYamlfiles is a WalkFunc to deal with yaml templates
func walkYamlfiles(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		// skip dir
		return nil
	}

	filename := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	yamlfiles[filename] = string(b)
	return nil
}

// listYamlfiles list all yaml templates' name
func listYamlfiles(request *restful.Request, response *restful.Response) {

	keys := make([]string, len(yamlfiles))
	i := 0
	for k := range yamlfiles {
		keys[i] = k
		i++
	}

	resp := map[string]interface{}{
		"templates": keys,
	}
	response.WriteAsJson(resp)
}

// getYamlfile get one yaml file content
func getYamlfile(request *restful.Request, response *restful.Response) {
	filename := request.PathParameter("yamlfile")

	fileStr, ok := yamlfiles[filename]
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"data": fileStr,
	}

	response.WriteAsJson(resp)

}

// listDockerfiles list all dockerfile templates' name
func listDockerfiles(request *restful.Request, response *restful.Response) {
	keys := make([]string, len(dockerfiles))
	i := 0
	for k := range dockerfiles {
		keys[i] = k
		i++
	}
	resp := map[string]interface{}{
		"templates": keys,
	}
	response.WriteAsJson(resp)
}

// getDockerfile get one dockerfile content
func getDockerfile(request *restful.Request, response *restful.Response) {
	filename := request.PathParameter("dockerfile")

	fileStr, ok := dockerfiles[filename]
	if !ok {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"data": fileStr,
	}

	response.WriteAsJson(resp)

}
