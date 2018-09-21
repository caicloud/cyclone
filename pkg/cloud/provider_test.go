/*
Copyright 2018 caicloud authors. All rights reserved.

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

package cloud

import (
	"fmt"
	"testing"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/util/http/errors"
)

func testProviderFunc(*api.Cloud) (Provider, error) {
	return nil, nil
}

func TestRegistryCloudProvider(t *testing.T) {
	testCases := map[string]struct {
		cloudType api.CloudType
		err       error
	}{
		"first register docker": {
			api.CloudTypeDocker,
			nil,
		},
		"second register docker": {
			api.CloudTypeDocker,
			fmt.Errorf("cloud provider %s has been registried", api.CloudTypeDocker),
		},
		"register k8s": {
			api.CloudTypeKubernetes,
			nil,
		},
	}

	for d, tc := range testCases {
		err := RegistryCloudProvider(tc.cloudType, testProviderFunc)
		if err != tc.err {
			if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
				t.Errorf("%s failed: expect %v, but got %v", d, tc.err, err)
			}
		}
	}
}

func TestNewCloudProvider(t *testing.T) {
	// Init cloud providers.
	cloudProviderFactory = make(map[api.CloudType]newCloudFunc)
	if err := RegistryCloudProvider(api.CloudTypeKubernetes, testProviderFunc); err != nil {
		t.Errorf("fail to init k8s providers")
	}

	testCases := map[string]struct {
		cloudCfg *api.Cloud
		err      error
	}{
		"cloud config should not be nil": {
			nil,
			fmt.Errorf("Cloud config is nil"),
		},
		"get k8s provider": {
			&api.Cloud{
				Type: api.CloudTypeKubernetes,
			},
			nil,
		},
		"get docker provider": {
			&api.Cloud{
				Type: api.CloudTypeDocker,
			},
			errors.ErrorUnsupported.Error("cloud type", api.CloudTypeDocker),
		},
	}

	for d, tc := range testCases {
		_, err := NewCloudProvider(tc.cloudCfg)
		if err != tc.err {
			if err != nil && err.Error() != tc.err.Error() {
				t.Errorf("%s failed: expect %v, but got %v", d, err, tc.err)
			}
		}
	}
}
