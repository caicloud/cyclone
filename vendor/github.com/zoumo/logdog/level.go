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

import "fmt"

const (
	// NothingLevel log level only used in filter
	NothingLevel Level = 0
	// DebugLevel log level
	DebugLevel Level = 1 //0x00000001
	// InfoLevel log level
	InfoLevel Level = 2 //0x00000010
	// WarnLevel log level
	WarnLevel Level = 4 //0x00000100
	// WarningLevel is alias of WARN
	WarningLevel Level = 4 //0x00000100
	// ErrorLevel log level
	ErrorLevel Level = 8 //0x00001000
	// NoticeLevel log level
	NoticeLevel Level = 16 //0x00010000
	// FatalLevel log level
	FatalLevel Level = 32 //0x00100000
	// AllLevel log level only used in filter
	AllLevel Level = 255 //0x11111111
)

var (
	// levelNames store level's name
	levelNames = make(map[Level]string)
)

// Level is a logging priority.
// Note that Level satisfies the Option interface
type Level int

func (l Level) String() string {
	if name, ok := levelNames[l]; ok {
		return name
	}
	return fmt.Sprintf("Level %d", l)
}

func init() {
	RegisterLevel("NOTHING", NothingLevel)
	RegisterLevel("DEBUG", DebugLevel)
	RegisterLevel("INFO", InfoLevel)
	RegisterLevel("WARN", WarnLevel)
	RegisterLevel("ERROR", ErrorLevel)
	RegisterLevel("NOTICE", NoticeLevel)
	RegisterLevel("FATAL", FatalLevel)
	RegisterLevel("ALL", AllLevel)

}
