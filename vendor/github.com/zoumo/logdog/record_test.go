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

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	fields = Fields{
		"a": 1,
		"b": "2",
		"c": 1.2,
		"d": errors.New("test fields"),
	}
	name     = "test"
	pathname = "test/record"
	fun      = "test/test.record"
	level    = DebugLevel
	line     = 1
)

func TestLogRecord(t *testing.T) {

	record := NewLogRecord(name, level, pathname, fun, line, "%s", "right fields", fields)

	assert.Equal(t, name, record.Name)
	assert.Equal(t, level, record.Level)
	assert.Equal(t, "record", record.FileName)
	assert.Equal(t, "test.record", record.FuncName)
	assert.Equal(t, "record", record.ShortFuncName)
	assert.Equal(t, fields, record.Fields)

	t.Log(record.GetMessage())

}

func TestLogRecordError(t *testing.T) {
	// fields should be the last one
	record := NewLogRecord(name, level, pathname, fun, line, "%s", fields, "error fields")
	assert.Nil(t, record.Fields)

}
