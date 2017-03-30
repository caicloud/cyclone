// Copyright 2016 Jim Zhang (jim.zoumo@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logdog

import "testing"
import "github.com/stretchr/testify/assert"

func TestOptionsInterface(t *testing.T) {
	assert.Implements(t, (*Option)(nil), NoticeLevel)
	assert.Implements(t, (*Option)(nil), NewTextFormatter())
	assert.Implements(t, (*Option)(nil), NewJSONFormatter())
	assert.Implements(t, (*Option)(nil), OptionCallerStackDepth(1))
	assert.Implements(t, (*Option)(nil), OptionEnableRuntimeCaller(true))
	assert.Implements(t, (*Option)(nil), OptionHandlers())
	assert.Implements(t, (*Option)(nil), OptionOutput(devNull(0)))
	assert.Implements(t, (*Option)(nil), OptionDiscardOutput())
}
