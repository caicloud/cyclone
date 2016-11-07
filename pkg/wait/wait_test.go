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

package wait

import (
	"testing"
	"time"
)

// TestPoll tests the poll function.
func TestPoll(t *testing.T) {
	result := 0
	err := Poll(3*time.Second, 120*time.Second, func() (bool, error) {
		result = result + 1
		return true, nil
	})
	if err != nil {
		t.Error("Expect error to be nil")
	}

	if result != 1 {
		t.Errorf("Expect result %d equals to 1", result)
	}
}
