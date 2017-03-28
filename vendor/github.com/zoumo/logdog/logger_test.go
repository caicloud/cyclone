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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	logger := GetLogger(name)
	logger.ApplyOptions(
		OptionHandlers(
			NewStreamHandler(OptionDiscardOutput()),
		),
		InfoLevel,
		OptionCallerStackDepth(3),
	)

	assert.Equal(t, name, logger.Name)
	assert.Len(t, logger.Handlers, 1)
	assert.Equal(t, logger.Level, InfoLevel)
	assert.Equal(t, 3, logger.CallerStackDepth)

	logger.ApplyOptions(
		OptionCallerStackDepth(2),
		OptionEnableRuntimeCaller(true),
	)
	logger.Debug("test debug") // filtered

	logger.ApplyOptions(NothingLevel)

	logger.Debug("who is your daddy", Fields{"who": "jim"})
	logger.Info("logdog is useful", Fields{"agree": "yes"})
	logger.Warn("warning warning", Fields{"x": "man"})
	logger.Notice("this notice is impotant", Fields{"x": "man"})
	logger.Error("error error..", Fields{"x": "man"})
	logger.Fatal("I have no idea !", Fields{"x": "man"})

	logger2 := NewLogger(
		OptionName("test2"),
		OptionEnableRuntimeCaller(true),
		Level(InfoLevel),
		OptionCallerStackDepth(2),
	)

	assert.Equal(t, "test2", logger2.Name)
	assert.True(t, logger2.EnableRuntimeCaller)
}

func TestJsonLogger(t *testing.T) {
	logger := GetLogger("json").AddHandlers(
		NewStreamHandler(NewJSONFormatter(), OptionDiscardOutput()),
	)

	logger.Info("this is json formatter1")
	logger.Notice("this is json formatter2", fields)

}

func TestLoggerInterface(t *testing.T) {
	assert.Implements(t, (*ConfigLoader)(nil), NewLogger())
}
