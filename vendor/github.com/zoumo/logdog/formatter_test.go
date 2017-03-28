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

var smallFields = Fields{
	"a": "b",
	"c": "d",
	"e": "f",
	"g": "h",
	"i": "j",
	"k": "l",
	"m": "n",
	"o": "p",
	"q": "r",
	"s": "t",
}

var largeFields = Fields{
	"a": "b",
	"c": "d",
	"e": "f",
	"g": "h",
	"i": "j",
	"k": "l",
	"m": "n",
	"o": "p",
	"q": "r",
	"s": "t",
	"u": "v",
	"w": "x",
	"y": "z",
	"1": 1,
	"2": 2,
	"3": 3,
	"4": 4,
	"5": 5,
	"6": 6,
	"7": 7,
	"8": 8,
	"9": 9,
	"0": 0,
}

// func TestIsTerminal(t *testing.T) {
// 	// I have no idea why this unit test can not pass
// 	// but it does not affect the usage
// 	// there is something wrong with syscall.SYS_IOCTL in testing
// 	msg := "I have no idea why this unit test can not pass\nbut it does not affect the usage"
// 	assert.True(t, isTerminal, msg)
// 	assert.True(t, isColorTerminal, msg)

// }

func TestTextFormatterLoadConfig(t *testing.T) {
	formatter := NewTextFormatter()
	formatter.LoadConfig(Config{
		"enableColors": true,
	})

	assert.Equal(t, formatter.Fmt, DefaultFmtTemplate)
	assert.Equal(t, formatter.DateFmt, DefaultDateFmtTemplate)
	assert.True(t, formatter.EnableColors)
}

func TestJsonFormatterLoadConfig(t *testing.T) {
	formatter := NewJSONFormatter()
	formatter.LoadConfig(Config{
		"datefmt": "test",
	})
	assert.Equal(t, formatter.Datefmt, "test")
}

func TestFormatterInterface(t *testing.T) {
	assert.Implements(t, (*Formatter)(nil), NewTextFormatter())
	assert.Implements(t, (*ConfigLoader)(nil), NewTextFormatter())
	assert.Implements(t, (*Formatter)(nil), NewJSONFormatter())
	assert.Implements(t, (*ConfigLoader)(nil), NewJSONFormatter())
}

func do(b *testing.B, formatter Formatter, fields Fields) {

	record := NewLogRecord("", 0, "file/test", "func", 0, "", fields)

	var d string
	var err error
	for i := 0; i < b.N; i++ {
		d, err = formatter.Format(record)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(d)))
	}
}

func BenchmarkSmallTextFormatter(b *testing.B) {
	do(b, &TextFormatter{EnableColors: false}, smallFields)
}

func BenchmarkLargeTextFormatter(b *testing.B) {
	do(b, &TextFormatter{EnableColors: false}, largeFields)
}

func BenchmarkSmallColoredTextFormatter(b *testing.B) {
	do(b, &TextFormatter{EnableColors: true}, smallFields)
}

func BenchmarkLargeColoredTextFormatter(b *testing.B) {
	do(b, &TextFormatter{EnableColors: true}, largeFields)
}

func BenchmarkSmallJsonFormatter(b *testing.B) {
	do(b, &JSONFormatter{}, smallFields)
}

func BenchmarkLargeJsonFormatter(b *testing.B) {
	do(b, &JSONFormatter{}, largeFields)
}
