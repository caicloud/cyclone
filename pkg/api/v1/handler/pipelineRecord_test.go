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

package handler

import (
	"testing"

	"github.com/caicloud/cyclone/pkg/api"
)

func TestAbortPipelineRecord(t *testing.T) {

	testCases := map[string]struct {
		record *api.PipelineRecord
		pass   bool
	}{
		"CodeCheckout Running": {
			&api.PipelineRecord{
				Status: api.Running,
				StageStatus: &api.StageStatus{

					CodeCheckout: &api.CodeCheckoutStageStatus{
						GeneralStageStatus: api.GeneralStageStatus{
							Status: api.Running,
						},
					},
					UnitTest: &api.GeneralStageStatus{
						Status: api.Running,
					},
					Package: &api.GeneralStageStatus{
						Status: api.Success,
					},
					ImageRelease: &api.ImageReleaseStageStatus{
						GeneralStageStatus: api.GeneralStageStatus{
							Status: api.Success,
						},
					},
				},
			},
			true,
		},
	}

	for k, tc := range testCases {
		abortPipelineRecord(tc.record)
		if tc.record.StageStatus.CodeCheckout.Status != api.Aborted {
			t.Errorf("%s failed as error : Expect result %v equals to running",
				k, tc.record.StageStatus.CodeCheckout.Status)
		}

		if tc.record.StageStatus.UnitTest.Status != api.Aborted {
			t.Errorf("%s failed as error : Expect result %v equals to running",
				k, tc.record.StageStatus.UnitTest.Status)
		}

		if tc.record.StageStatus.Package.Status != api.Success {
			t.Errorf("%s failed as error : Expect result %v equals to success",
				k, tc.record.StageStatus.Package.Status)
		}

		if tc.record.StageStatus.ImageRelease.Status != api.Success {
			t.Errorf("%s failed as error : Expect result %v equals to success",
				k, tc.record.StageStatus.ImageRelease.Status)
		}
	}
}
