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

package event

import (
	"testing"

	"github.com/caicloud/circle/api"
)

// TestToBuildContainerConfig tests the BuildContainerConfig function.
func TestToBuildContainerConfig(t *testing.T) {
	var cpu int64
	var memory int64
	var eventID api.EventID

	eventID = "unit-test-eventid"
	cpu = 1
	memory = 1024

	option := toBuildContainerConfig(eventID, cpu, memory)
	if option.HostConfig.Memory != 1024 || option.HostConfig.CPUShares != cpu {
		t.Errorf("Expect memory equals to %d, cpu equals to %d", memory, cpu)
	}
}
