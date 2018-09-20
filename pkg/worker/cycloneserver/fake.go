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

package cycloneserver

import (
	"net/http"
	"strings"

	"github.com/caicloud/cyclone/pkg/api"
)

type fakeClient struct {
	baseURL string
	client  *http.Client
}

// NewFakeClient new fake client to communicate with Cyclone server for test.
func NewFakeClient(cycloneServer string) CycloneServerClient {
	baseURL := strings.TrimRight(cycloneServer, "/")
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}

	return &fakeClient{
		baseURL: baseURL,
		client:  http.DefaultClient,
	}
}

func (c *fakeClient) SendEvent(event *api.Event) error {
	return nil
}

func (c *fakeClient) GetEvent(id string) (*api.Event, error) {
	return nil, nil
}

func (c *fakeClient) PushLogStream(project, pipeline, recordID string, stage api.PipelineStageName, task string, filePath string, close chan struct{}) error {
	return nil
}

func (c *fakeClient) SendJUnitFile(project, pipeline, recordID string, path string) error {
	return nil
}
