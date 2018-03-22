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
	initStatsDetails(statistics, 1521345720, 1522468920)
	if len(statistics.Details) != 14 {
		t.Errorf("Expect result %d equals to 14", len(statistics.Details))
	}

	statistics.Details = []*api.StatsDetail{}
	initStatsDetails(statistics, 1521345720, 1522468919)
	if len(statistics.Details) != 14 {
		t.Errorf("Expect result %d equals to 14", len(statistics.Details))
	}

	statistics.Details = []*api.StatsDetail{}
	initStatsDetails(statistics, 1521345720, 1522468921)
	if len(statistics.Details) != 14 {
		t.Errorf("Expect result %d equals to 14", len(statistics.Details))
	}

	statistics.Details = []*api.StatsDetail{}
	initStatsDetails(statistics, 1521345720, 1521345720)
	if len(statistics.Details) != 1 {
		t.Errorf("Expect result %d equals to 14", len(statistics.Details))
	}

	statistics.Details = []*api.StatsDetail{}
	initStatsDetails(statistics, 1521345720, 1521345719)
	if len(statistics.Details) != 0 {
		t.Errorf("Expect result %d equals to 14", len(statistics.Details))
	}

}
