/*
Copyright 2018 Caicloud Authors

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
	"encoding/json"
	"encoding/xml"
)

// DataType is the type of raw data.
type DataType string

const (
	// DataTypeJSON corresponds to content type "application/json".
	DataTypeJSON DataType = "json"
	// DataTypeXML corresponds to content type "application/xml".
	DataTypeXML DataType = "xml"
	// DataTypePlain indicates there is a plain error message.
	DataTypePlain DataType = ""
)

// ExternalError describes an error interface for client error.
type ExternalError interface {
	error
	// Code returns status code of the error.
	Code() int
	// Reason returns reason string of the error.
	Reason() string
	// Data returns data map of the error.
	Data() map[string]string
}

type externalError struct {
	message
	code int
}

// Code returns status code of the error.
func (e *externalError) Code() int {
	return e.code
}

// Reason returns reason of the error.
func (e *externalError) Reason() string {
	return string(e.message.Reason)
}

// Data returns data map of the error.
func (e *externalError) Data() map[string]string {
	return e.message.Data
}

// Message returns detailed message of the error.
func (e *externalError) Message() interface{} {
	return &e.message
}

// Error returns error description.
func (e *externalError) Error() string {
	return e.message.Message
}

// ParseError parse error from raw data.
func ParseError(code int, dt DataType, data []byte) (ExternalError, error) {
	e := &externalError{
		code: code,
	}
	switch dt {
	case DataTypeJSON:
		return e, json.Unmarshal(data, &e.message)
	case DataTypeXML:
		return e, xml.Unmarshal(data, &e.message)
	}
	e.message.Message = string(data)
	return e, nil
}
