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
	"io"
	"os"
	"reflect"
)

// Option is an interface which is used to set options for the target
type Option interface {
	applyOption(target interface{}) bool
}

// optFuncWraper wraps a function so it satisfies the Option interface.
type optFuncWraper func(interface{}) bool

func (f optFuncWraper) applyOption(target interface{}) bool {
	ret := f(target)
	if !ret {
		fmt.Fprintf(os.Stderr, "target[%T] does not support this option\n", target)
	}
	return ret
}

// makes Level satisfies the Option interface.
// used in every target which has fields named `Level`
func (l Level) applyOption(target interface{}) bool {
	v := reflect.ValueOf(target).Elem()
	if level := v.FieldByName("Level"); level.IsValid() {
		level.Set(reflect.ValueOf(l))
		return true
	}
	return false
}

func (tf *TextFormatter) applyOption(target interface{}) bool {
	v := reflect.ValueOf(target).Elem()
	if f := v.FieldByName("Formatter"); f.IsValid() {
		f.Set(reflect.ValueOf(tf))
		return true
	}
	return false
}

func (fh *JSONFormatter) applyOption(target interface{}) bool {
	v := reflect.ValueOf(target).Elem()
	if f := v.FieldByName("Formatter"); f.IsValid() {
		f.Set(reflect.ValueOf(fh))
		return true
	}
	return false
}

// OptionName is an option
// used in every target which has fields named `Name`
func OptionName(name string) Option {
	return optFuncWraper(func(target interface{}) bool {
		v := reflect.ValueOf(target).Elem()
		if n := v.FieldByName("Name"); n.IsValid() {
			n.SetString(name)
			return true
		}
		return false
	})
}

// OptionCallerStackDepth is an option.
// used in every target which has fields named `CallerStackDepth`
func OptionCallerStackDepth(depth int) Option {
	return optFuncWraper(func(target interface{}) bool {
		v := reflect.ValueOf(target).Elem()
		if f := v.FieldByName("CallerStackDepth"); f.IsValid() {
			f.SetInt(int64(depth))
			return true
		}
		return false
	})
}

// OptionEnableRuntimeCaller is an option useed in :
// used in every target which has fields named `EnableRuntimeCaller`
func OptionEnableRuntimeCaller(enable bool) Option {
	return optFuncWraper(func(target interface{}) bool {
		v := reflect.ValueOf(target).Elem()
		if f := v.FieldByName("EnableRuntimeCaller"); f.IsValid() {
			f.SetBool(enable)
			return true
		}
		return false
	})
}

// OptionHandlers is an option
// used in every target which has fields named `Handlers`
func OptionHandlers(handlers ...Handler) Option {
	return optFuncWraper(func(target interface{}) bool {
		v := reflect.ValueOf(target).Elem()
		if f := v.FieldByName("Handlers"); f.IsValid() {
			if f.Kind() == reflect.Slice {
				// f.Set(reflect.AppendSlice(f, reflect.ValueOf(handlers)))
				f.Set(reflect.ValueOf(handlers))
				return true
			}
		}
		return false
	})
}

// OptionOutput is an option
// used in every target which has fields named `Output`
func OptionOutput(out io.WriteCloser) Option {
	return optFuncWraper(func(target interface{}) bool {
		v := reflect.ValueOf(target).Elem()
		if f := v.FieldByName("Output"); f.IsValid() {
			f.Set(reflect.ValueOf(out))
			return true
		}
		return false
	})
}

// OptionDiscardOutput is an option
// used in every target which has fields named `Output`
// and make all Read | Write | Close calls succeed without doing anything.
func OptionDiscardOutput() Option {
	return optFuncWraper(func(target interface{}) bool {
		v := reflect.ValueOf(target).Elem()
		if f := v.FieldByName("Output"); f.IsValid() {
			f.Set(reflect.ValueOf(Discard))
			return true
		}
		return false
	})
}
