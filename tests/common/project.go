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
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/cyclone/pkg/api"
)

type Metadata struct {
	Total int `bson:"total,omitempty" json:"total,omitempty"`
}

type ListResponse struct {
	Metadata Metadata      `bson:"metadata,omitempty" json:"metadata,omitempty"`
	Items    []api.Project `bson:"items,omitempty" json:"items,omitempty"`
}

// CreateProject creates a project with given configurations.
func CreateProject(project *api.Project, response *api.Project, errResp *api.ErrorResponse) (int, error) {
	buf, err := json.Marshal(project)
	if err != nil {
		return 500, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/projects", BaseURL), bytes.NewBuffer(buf))
	if err != nil {
		return 500, err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 500, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode/100 == 2 {
		err = json.Unmarshal(respBody, response)
		if err != nil {
			fmt.Println("unmarshl error :", err)
			return resp.StatusCode, err
		}
	} else {
		err = json.Unmarshal(respBody, errResp)
		if err != nil {
			fmt.Println("unmarshl p error :", err)
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

// SetProject set a project with given configurations.
func SetProject(project *api.Project, response *api.Project, errResp *api.ErrorResponse) (int, error) {
	buf, err := json.Marshal(project)
	if err != nil {
		return 500, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/projects/%s", BaseURL, project.Name), bytes.NewBuffer(buf))
	if err != nil {
		return 500, err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 500, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode/100 == 2 {
		err = json.Unmarshal(respBody, response)
		if err != nil {
			fmt.Println("unmarshl error :", err)
			return resp.StatusCode, err
		}
	} else {
		err = json.Unmarshal(respBody, errResp)
		if err != nil {
			fmt.Println("unmarshl p error :", err)
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

// GetProject retrieves a project from user ID and project ID.
func GetProject(name string, response *api.Project, errResp *api.ErrorResponse) (int, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/projects/%s", BaseURL, name), nil)
	if err != nil {
		return 500, err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 500, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode/100 == 2 {
		err = json.Unmarshal(respBody, response)
		if err != nil {
			fmt.Println("unmarshl error :", err)
			return resp.StatusCode, err
		}
	} else {
		err = json.Unmarshal(respBody, errResp)
		if err != nil {
			fmt.Println("unmarshl p error :", err)
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

//// ListProjects list projects by user ID.
func ListProjects(response *ListResponse, errResp *api.ErrorResponse) (int, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/projects", BaseURL), nil)
	if err != nil {
		return 500, err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 500, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode/100 == 2 {
		err = json.Unmarshal(respBody, response)
		if err != nil {
			fmt.Println("unmarshl error :", err)
			return resp.StatusCode, err
		}
	} else {
		err = json.Unmarshal(respBody, errResp)
		if err != nil {
			fmt.Println("unmarshl p error :", err)
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}

// DeleteProject delete a project from user ID and project ID.
func DeleteProject(name string, response *api.Project, errResp *api.ErrorResponse) (int, error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/projects/%s", BaseURL, name), nil)
	if err != nil {
		return 500, err
	}
	generateRequestWithToken(req, "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 500, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode/100 == 2 {
		return resp.StatusCode, nil
	} else {
		err = json.Unmarshal(respBody, errResp)
		if err != nil {
			fmt.Println("unmarshl p error :", err)
			return resp.StatusCode, err
		}
	}

	return resp.StatusCode, nil
}
