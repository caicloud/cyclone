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

package worker

import (
	"reflect"
	"testing"

	"github.com/caicloud/cyclone/pkg/api"
)

func TestConvertPerformStageSet(t *testing.T) {
	testCases := map[string]struct {
		stages   []api.PipelineStageName
		stageSet map[api.PipelineStageName]struct{}
	}{
		"only code checkout stage": {
			stages: []api.PipelineStageName{api.CodeCheckoutStageName},
			stageSet: map[api.PipelineStageName]struct{}{
				api.CodeCheckoutStageName: struct{}{},
			},
		},
		"two stages": {
			stages: []api.PipelineStageName{api.CodeCheckoutStageName, api.IntegrationTestStageName},
			stageSet: map[api.PipelineStageName]struct{}{
				api.CodeCheckoutStageName:    struct{}{},
				api.IntegrationTestStageName: struct{}{},
			},
		},
		"unorder stages": {
			stages: []api.PipelineStageName{api.PackageStageName, api.CodeCheckoutStageName, api.IntegrationTestStageName},
			stageSet: map[api.PipelineStageName]struct{}{
				api.PackageStageName:         struct{}{},
				api.CodeCheckoutStageName:    struct{}{},
				api.IntegrationTestStageName: struct{}{},
			},
		},
		"full stages": {
			stages: []api.PipelineStageName{api.CodeCheckoutStageName, api.UnitTestStageName, api.PackageStageName,
				api.CodeScanStageName, api.ImageBuildStageName, api.IntegrationTestStageName, api.ImageReleaseStageName},
			stageSet: map[api.PipelineStageName]struct{}{
				api.CodeCheckoutStageName:    struct{}{},
				api.UnitTestStageName:        struct{}{},
				api.PackageStageName:         struct{}{},
				api.CodeScanStageName:        struct{}{},
				api.ImageBuildStageName:      struct{}{},
				api.IntegrationTestStageName: struct{}{},
				api.ImageReleaseStageName:    struct{}{},
			},
		},
	}

	for d, tc := range testCases {
		result := convertPerformStageSet(tc.stages)
		if !reflect.DeepEqual(result, tc.stageSet) {
			t.Errorf("fail to convert %s: expect %v, but got %v", d, tc.stageSet, result)
		}
	}
}
