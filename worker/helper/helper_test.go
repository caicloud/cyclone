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

package helper

import (
	"testing"

	"github.com/caicloud/cyclone/utils"
	"github.com/caicloud/cyclone/worker/ci/yaml"
)

const (
	mockUserID = "mock-user"
	imageName  = "mock-image-name"
)

// TestUpdateContainerInClusterWithYaml tests updateContainerInClusterWithYaml function.
func TestUpdateContainerInClusterWithYaml(t *testing.T) {
	app := yaml.Application{
		ClusterType: "NOT_K8S",
	}
	if err := updateContainerInClusterWithYaml(utils.DeployUID, imageName, app); err != nil {
		t.Errorf("Expected err %v to be nil", err)
	}
}
