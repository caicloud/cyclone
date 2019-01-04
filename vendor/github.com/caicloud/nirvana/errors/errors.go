/*
Copyright 2017 Caicloud Authors

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
	"encoding/xml"
	"fmt"
)

// Reason is an enumeration of possible failure causes. Each Reason
// must map to a format which is a string containing ${formatArgu1}.
//
// Following format is recommended:
//   MuduleName[:SubmoduleName]:ShortErrorDescription
// Examples:
//   Reason "Nirvana:KindNotFound" may map to format "${kindName} was not found".
//   Reason "Nirvana:SomeoneIsSleeping" may map to format "${name} is sleeping now"
type Reason string

// Factory can create error from a fixed format.
type Factory interface {
	// Error generates an error from v.
	Error(v ...interface{}) error
	// Derived checks if an error was derived from current factory.
	Derived(e error) bool
}

// dataMap is a wrapper for marshalling map into XML. Standard package can't
// generate XML for map.
type dataMap map[string]string

// MarshalXML marshals data map into XML.
func (dm dataMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	tokens := []xml.Token{start}
	for key, value := range dm {
		t := xml.StartElement{
			Name: xml.Name{
				Local: key,
			},
		}
		tokens = append(tokens, t, xml.CharData(value), xml.EndElement{Name: t.Name})
	}
	tokens = append(tokens, xml.EndElement{Name: start.Name})

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	return e.Flush()
}

// message can be marshaled for transferring.
//
// Example:
// {
//   "message": "name of something is to short",
//   "reason": "SomeModule:NameTooShort",
//   "data": {
//     "name": "something"
//   }
// }
type message struct {
	// Reason is a unique key for an error in global environment.
	Reason Reason `json:"reason,omitempty" xml:",omitempty"`
	// Message contains the detailed description of an error.
	Message string `json:"message"`
	// Data is used for i18n.
	Data dataMap `json:"data,omitempty" xml:",omitempty"`
}

// Error returns error description.
func (m *message) Error() string {
	return m.Message
}

// err implements error interface.
type err struct {
	message
	factory *factory
}

// Code returns status code of the error.
func (e *err) Code() int {
	return e.factory.code
}

// Message returns detailed message of the error.
func (e *err) Message() interface{} {
	return &e.message
}

// Error returns error description.
func (e *err) Error() string {
	return e.message.Message
}

// factory is an error factory.
type factory struct {
	code   int
	reason Reason
	format string
}

// Code returns code of current factory.
func (f *factory) Code() int {
	return f.code
}

// Reason returns reason of current factory.
func (f *factory) Reason() Reason {
	return f.reason
}

// Error generates an error from v.
func (f *factory) Error(v ...interface{}) error {
	msg := message{Reason: f.reason}
	msg.Message, msg.Data = expand(f.format, v...)
	return &err{
		message: msg,
		factory: f,
	}
}

// Derived checks if an error was derived from current factory.
// If an error is not derived by current factory but implements
// ExternalError as well as code and reason are matched,
// this method also returns true.
func (f *factory) Derived(e error) bool {
	origin, ok := e.(*err)
	if ok {
		return origin.factory == f
	}
	external, ok := e.(ExternalError)
	return ok && external.Code() == f.code &&
		external.Reason() == string(f.reason)
}

// expand expands a format string like "name ${name} is too short" to "name japari is too short"
// by replacing ${} with v... one by one.
// Note that if len(v) < count of ${}, it will panic.
func expand(format string, v ...interface{}) (msg string, data map[string]string) {
	n := 0
	var m map[string]string
	buf := make([]byte, 0, len(format))

	for i := 0; i < len(format); {
		if format[i] == '$' && (i+1) < len(format) && format[i+1] == '{' {
			b := make([]byte, 0, len(format)-i)
			if i+2 == len(format) { // check "...${"
				panic("unexpected EOF while looking for matching }")
			}
			ii := i + 2
			for ii < len(format[i+2:])+i+2 {
				if format[ii] != '}' {
					b = append(b, format[ii])
				} else {
					break
				}
				ii++
				if ii == len(format[i+2:])+i+2 { // check "...${..."
					panic("unexpected EOF while looking for matching }")
				}
			}
			i = ii + 1
			if n == len(v) {
				panic("not enough args")
			}
			if m == nil {
				m = map[string]string{}
			}
			m[string(b)] = fmt.Sprint(v[n])
			buf = append(buf, fmt.Sprint(v[n])...)
			n++
		} else {
			buf = append(buf, format[i])
			i++
		}
	}
	return string(buf), m
}
