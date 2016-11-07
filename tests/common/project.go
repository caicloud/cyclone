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

package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/cyclone/api"
)

// CreateProject creates a project with given configurations.
func CreateProject(userID string, project *api.Project, response *api.ProjectCreationResponse) error {
	buf, err := json.Marshal(project)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/projects", BaseURL, userID), bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// SetProject set a project with given configurations.
func SetProject(userID string, projectID string, project *api.Project, response *api.ProjectSetResponse) error {
	buf, err := json.Marshal(project)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s/projects/%s", BaseURL, userID, projectID),
		bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// GetProject retrieves a project from user ID and project ID.
func GetProject(userID, projectID string, response *api.ProjectGetResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/projects/%s", BaseURL, userID, projectID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// ListProjects list projects by user ID.
func ListProjects(userID string, response *api.ProjectListResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/projects", BaseURL, userID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// DeleteProject delete a project from user ID and project ID.
func DeleteProject(userID, projectID string, response *api.ProjectDelResponse) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s/projects/%s", BaseURL, userID, projectID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// CreateProjectVersion creates a project version with given configurations.
func CreateProjectVersion(userID string, projectVersion *api.ProjectVersion,
	response *api.ProjectVersionCreationResponse) error {
	buf, err := json.Marshal(projectVersion)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s/versions_project", BaseURL, userID),
		bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// GetProjectVersion retrieves a project version from user ID and project version ID.
func GetProjectVersion(userID, projectVersionID string, response *api.ProjectVersionGetResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/projectversions/%s", BaseURL, userID, projectVersionID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}

// ListProjectVersions list project versions by user ID.
func ListProjectVersions(userID, projectID string, response *api.ProjectVersionListResponse) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/projects/%s/versions", BaseURL, userID, projectID), nil)
	if err != nil {
		return err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respBody, response)
	if err != nil {
		return err
	}
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}
	return nil
}
