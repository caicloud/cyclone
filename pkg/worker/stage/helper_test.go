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

package stage

import (
	"errors"
	"fmt"
	"testing"

	"github.com/caicloud/cyclone/pkg/api"
)

func TestGenerateStageStartLog(t *testing.T) {
	testCases := map[string]struct {
		stage       api.PipelineStageName
		expectedLog string
	}{
		"code checkout stage": {
			api.CodeCheckoutStageName,
			"Stage: Code checkout status: start\n",
		},
		"package stage": {
			api.PackageStageName,
			"Stage: Package status: start\n",
		},
		"build image stage": {
			api.ImageBuildStageName,
			"Stage: Build image status: start\n",
		},
		"error stage": {
			api.PipelineStageName("errorState"),
			"",
		},
	}

	for d, tc := range testCases {
		result := generateStageStartLog(tc.stage)
		if result != tc.expectedLog {
			t.Errorf("%s fails to generate start log: expected %s; but got %s", d, tc.expectedLog, result)
		}
	}
}

func TestGenerateStageFinishLog(t *testing.T) {
	testCases := map[string]struct {
		stage       api.PipelineStageName
		err         error
		expectedLog string
	}{
		"code checkout stage": {
			api.CodeCheckoutStageName,
			nil,
			"Stage: Code checkout status: finish\n",
		},
		"package stage": {
			api.PackageStageName,
			fmt.Errorf("output is not found"),
			"Stage: Package status: fail with error: output is not found\n",
		},
		"build image stage": {
			api.ImageBuildStageName,
			errors.New("dockerfile is not found"),
			"Stage: Build image status: fail with error: dockerfile is not found\n",
		},
		"error stage": {
			api.PipelineStageName("errorState"),
			fmt.Errorf("stage not correct"),
			"",
		},
	}

	for d, tc := range testCases {
		result := generateStageFinishLog(tc.stage, tc.err)
		if result != tc.expectedLog {
			t.Errorf("%s fails to generate start log: expected %s; but got %s", d, tc.expectedLog, result)
		}
	}
}

func TestUpdateRecordStageStatus(t *testing.T) {
	pr := &api.PipelineRecord{
		StageStatus: &api.StageStatus{
			CodeCheckout: &api.CodeCheckoutStageStatus{},
			Package:      &api.GeneralStageStatus{},
		},
	}

	if err := updateRecordStageStatus(pr, api.CodeCheckoutStageName, api.Running, nil); err == nil {
		if pr.StageStatus.CodeCheckout.Status != api.Running {
			t.Errorf("expect code checkout stage status to be Running, but get %s", pr.StageStatus.CodeCheckout.Status)
		}
	} else {
		t.Errorf("expect error to be nil, but got %v", err)
	}

	if err := updateRecordStageStatus(pr, api.PackageStageName, api.Failed, fmt.Errorf("meets error")); err == nil {
		if pr.StageStatus.Package.Status != api.Failed {
			t.Errorf("expect code checkout stage status to be Running, but get %s", pr.StageStatus.CodeCheckout.Status)
		}
		if pr.StageStatus.Package.EndTime.IsZero() {
			t.Errorf("expect end time of package stage status not be empty")
		}
		if pr.ErrorMessage == "" {
			t.Errorf("expect error message of record not be empty")
		}
	} else {
		t.Errorf("expect error to be nil, but got %v", err)
	}
}
