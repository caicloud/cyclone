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

func TestStreamHandler(t *testing.T) {
	record := NewLogRecord(name, DebugLevel, pathname, fun, line, "%s", "debug", fields)
	record2 := NewLogRecord(name, InfoLevel, pathname, fun, line, "%s", "success", fields)
	handler := NewStreamHandler()
	// handler.Level = INFO
	err := handler.LoadConfig(Config{
		"level": InfoLevel.String(),
	})
	assert.Nil(t, err)
	assert.Equal(t, handler.Level, InfoLevel)
	assert.Equal(t, handler.Formatter, TerminalFormatter)
	assert.True(t, handler.Filter(record))
	assert.False(t, handler.Filter(record2))
	// sync on stderr stdout will fail
	assert.Error(t, handler.Flush())
	assert.Nil(t, handler.Close())

}

func TestStreamHandlerApplyOption(t *testing.T) {
	fmt := NewTextFormatter()
	hdlr := NewStreamHandler(
		OptionName(name),
		DebugLevel,
		fmt,
		OptionDiscardOutput(),
	)
	assert.Equal(t, hdlr.Name, name)
	assert.Equal(t, hdlr.Level, DebugLevel)
	assert.Equal(t, hdlr.Formatter, fmt)
	assert.Equal(t, hdlr.Output, Discard)
}

func TestFileHandler(t *testing.T) {
	record := NewLogRecord(name, DebugLevel, pathname, fun, line, "%s", "debug", fields)
	record2 := NewLogRecord(name, InfoLevel, pathname, fun, line, "%s", "success", fields)
	handler := NewFileHandler()

	err := handler.LoadConfig(Config{
		"level":     InfoLevel.String(),
		"filename":  "/dev/null",
		"formatter": "terminal",
	})
	assert.Nil(t, err)
	assert.Equal(t, handler.Path, "/dev/null")
	assert.Equal(t, handler.Formatter, TerminalFormatter)
	assert.Equal(t, handler.Level, InfoLevel)

	assert.True(t, handler.Filter(record))
	assert.False(t, handler.Filter(record2))
	assert.Nil(t, handler.Close())
}

func TestFileHandlerApplyOption(t *testing.T) {
	fmt := NewTextFormatter()
	hdlr := NewStreamHandler(
		OptionName(name),
		DebugLevel,
		fmt,
		OptionDiscardOutput(),
	)
	assert.Equal(t, hdlr.Name, name)
	assert.Equal(t, hdlr.Level, DebugLevel)
	assert.Equal(t, hdlr.Formatter, fmt)
	assert.Equal(t, hdlr.Output, Discard)
}

func TestHandlerInterface(t *testing.T) {
	assert.Implements(t, (*Handler)(nil), NewStreamHandler())
	assert.Implements(t, (*ConfigLoader)(nil), NewStreamHandler())
	assert.Implements(t, (*Handler)(nil), NewFileHandler())
	assert.Implements(t, (*ConfigLoader)(nil), NewFileHandler())
}
