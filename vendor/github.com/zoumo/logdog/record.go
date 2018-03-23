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
	"fmt"
	"path"
	"sort"
	"strings"
	"time"
)

// Use simple []byte instead of bytes.Buffer to avoid large dependency.
type buffer []byte

func (b *buffer) Write(p []byte) (int, error) {
	*b = append(*b, p...)
	return len(p), nil
}

// Fields is an alias to man[string]interface{}
type Fields map[string]interface{}

// String convert Fields to string
func (f Fields) String() string {
	return fmt.Sprintf("%#v", (map[string]interface{})(f))
}

// ToKVString convert Fields to string likes k1=v1 k2=v2
func (f Fields) ToKVString(color, endColor string) string {

	if len(f) == 0 {
		return ""
	}

	b := &buffer{}
	fmt.Fprint(b, " | ")

	sorted := make([]string, 0, len(f))
	for k := range f {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	first := true
	for _, k := range sorted {
		v := f[k]
		// auto format time to RFC3339
		if vv, ok := v.(time.Time); ok {
			v = vv.Format(time.RFC3339)
		}

		if first {
			fmt.Fprintf(b, "%s=%s%+v%s", k, color, v, endColor)
			first = false
		} else {
			fmt.Fprintf(b, " %s=%s%+v%s", k, color, v, endColor)
		}
	}

	return string(*b)
}

// LogRecord defines a real log record should be
type LogRecord struct {
	Name          string
	Level         Level
	LevelName     string
	PathName      string
	FileName      string
	FuncName      string
	ShortFuncName string
	Line          int
	Time          time.Time
	// msg could be ""
	Msg  string
	Args []interface{}
	// extract fields from args
	Fields Fields
}

// NewLogRecord returns a new log record
func NewLogRecord(name string, level Level, pathname string, funcname string, line int, msg string, args ...interface{}) *LogRecord {
	record := LogRecord{
		Name:     name,
		Level:    level,
		PathName: pathname,
		Line:     line,
		Msg:      msg,
		Args:     args,
		Time:     time.Now(),
	}
	// level name
	record.LevelName = level.String()

	// file name
	_, filename := path.Split(pathname)
	record.FileName = filename

	// func name
	i := strings.LastIndex(funcname, "/")
	record.FuncName = funcname[i+1:]
	j := strings.LastIndex(funcname[i+1:], ".")
	record.ShortFuncName = record.FuncName[j+1:]

	// split args and fields
	record.ExtractFieldsFromArgs()

	return &record
}

// GetMessage formats record message by msg and args
func (lr LogRecord) GetMessage() string {
	msg := lr.Msg
	buf := &buffer{}
	if msg == "" {
		fmt.Fprintln(buf, lr.Args...)
		msg = string(*buf)
		msg = msg[:len(msg)-1]
	} else {
		msg = fmt.Sprintf(lr.Msg, lr.Args...)
	}
	return msg
}

// ExtractFieldsFromArgs extracts fields (Fields) from args
// Fields must be the last element in args
func (lr *LogRecord) ExtractFieldsFromArgs() {
	argsLen := len(lr.Args)
	if argsLen == 0 {
		return
	}

	if fields, ok := lr.Args[argsLen-1].(Fields); ok {
		lr.Args = lr.Args[:argsLen-1]
		lr.Fields = fields
	}

}
