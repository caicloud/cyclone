/*
Copyright 2017 caicloud authors. All rights reserved.

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

package errors

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
)

// Error defines error with code
type Error struct {
	ID      int    `json:"id,omitempty"`
	Code    int    `json:"code,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	format  string
	pc      uintptr
}

// Error returns error reason
func (e *Error) Error() string {
	return e.Message
}

// Equal returns whether err.ID equal to e.ID
func (e *Error) Equal(err error) bool {
	if errx, ok := (err).(*Error); ok {
		return errx.ID == e.ID
	}
	return false
}

// Format generate an specified error
func (e *Error) Format(params ...interface{}) *Error {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		pc = 0
	}
	if len(e.format) <= 0 {
		return e
	}
	return &Error{
		e.ID,
		e.Code,
		e.Reason,
		fmt.Sprintf(e.format, params...),
		e.Detail,
		e.format,
		pc,
	}
}

// id counter
var counter int32

// NewErrorID generates an unique id in current runtime
func NewErrorID() int {
	return int(atomic.AddInt32(&counter, 1))
}

// NewStaticError creates a static error
func NewStaticError(code int, reason string, message string) *Error {
	return &Error{NewErrorID(), code, reason, message, "", "", 0}
}

// NewFormatError creates a format error
func NewFormatError(code int, reason string, format string) *Error {
	return &Error{NewErrorID(), code, reason, "", "", format, 0}
}

// ================================

const (
	prefix = " ==> "
)

func (r *Error) ErrorDetail() string {
	b := make([]byte, 1, 64)
	b[0] = '\n'
	b = r.AppendErrorDetail(b)
	r.Detail = string(b)
	return r.Detail
}

// AppendErrorDetail returns a byte slice joined err, cmd and pc in ErrorDetail.
func (r *Error) AppendErrorDetail(b []byte) []byte {
	b = append(b, prefix...)
	if r.pc != 0 {
		f := runtime.FuncForPC(r.pc)
		if f != nil {
			file, line := f.FileLine(r.pc)
			b = append(b, shortFile(file)...)
			b = append(b, ':')
			b = append(b, strconv.Itoa(line)...)
			b = append(b, ':', ' ')

			fnName := f.Name()
			fnName = fnName[strings.LastIndex(fnName, "/")+1:]
			fnName = fnName[strings.Index(fnName, ".")+1:]
			b = append(b, '[')
			b = append(b, fnName...)
			b = append(b, ']', ' ')
		}
	}

	b = append(b, r.Message...)
	b = append(b, ' ', '~', ' ')
	b = append(b, r.Reason...)
	return b
}

func appendErrorDetail(b []byte, err error) []byte {
	if e, ok := err.(*Error); ok {
		return e.AppendErrorDetail(b)
	}
	b = append(b, prefix...)
	return append(b, err.Error()...)
}

func shortFile(file string) string {
	pos := strings.LastIndex(file, "/src/")
	if pos != -1 {
		return file[pos+5:]
	}
	return file
}
