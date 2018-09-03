/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package errors

import (
	"fmt"
	"sync/atomic"
)

// Error defines error with code
type Error struct {
	ID      int    `json:"-"`
	Code    int    `json:"code"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Detail  string `json:"detail, omitempty"`
	format  string
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
	return &Error{NewErrorID(), code, reason, message, "", ""}
}

// NewFormatError creates a format error
func NewFormatError(code int, reason string, format string) *Error {
	return &Error{NewErrorID(), code, reason, "", "", format}
}
