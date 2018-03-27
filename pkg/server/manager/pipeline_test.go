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

package manager

import (
	"testing"
	"time"

	"github.com/caicloud/cyclone/pkg/api"
)

func TestInitStatsDetails(t *testing.T) {
	statistics := &api.PipelineStatusStats{
		Overview: api.StatsOverview{
			Total:        0,
			SuccessRatio: "0.00%",
		},
		Details: []*api.StatsDetail{},
	}
	// 1521345720 2018/3/18 12:2:0
	// 1522468920 2018/3/31 12:2:0

	testCases := map[string]struct {
		start int64
		end   int64
		pass  int
	}{
		"end = start + n*86400": {
			1521345720,
			1522468920,
			14,
		},
		"end-1 = start + n*86400": {
			1521345720,
			1522468919,
			14,
		},
		"end+1 = start + n*86400": {
			1521345720,
			1522468921,
			14,
		},
		"end = start": {
			1521345720,
			1521345720,
			1,
		},
		"end < start": {
			1521345720,
			1521345719,
			0,
		},
	}

	for d, tc := range testCases {
		initStatsDetails(statistics, time.Unix(tc.start, 0), time.Unix(tc.end, 0))
		if len(statistics.Details) != tc.pass {
			t.Errorf("%s failed as error : Expect result %d equals to %d", d, len(statistics.Details), tc.pass)
		}
		statistics.Details = []*api.StatsDetail{}
	}
}
