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

import "github.com/zoumo/register"

var (
	formatters   = register.NewRegister(nil)
	handlers     = register.NewRegister(nil)
	constructors = register.NewRegister(nil)
	loggers      = register.NewRegister(nil)
	levels       = register.NewRegister(nil)
)

// Constructor is a function which returns an ConfigLoader
type Constructor func() ConfigLoader

// GetConstructor returns an Constructor registered with the given name
// if not, returns nil
func GetConstructor(name string) Constructor {
	v, ok := constructors.Get(name)
	if !ok {
		return nil
	}
	return v.(Constructor)
}

// RegisterConstructor binds name and Constructor
func RegisterConstructor(name string, c Constructor) {
	constructors.Register(name, c)
}

// RegisterFormatter binds name and Formatter
func RegisterFormatter(name string, formatter Formatter) {
	formatters.Register(name, formatter)
}

// GetFormatter returns an Formatter registered with the given name
func GetFormatter(name string) Formatter {
	v, ok := formatters.Get(name)
	if !ok {
		return nil
	}
	return v.(Formatter)
}

// RegisterHandler binds name and Handler
func RegisterHandler(name string, handler Handler) {
	handlers.Register(name, handler)
}

// GetHandler returns a Handler registered with the given name
func GetHandler(name string) Handler {
	v, ok := handlers.Get(name)
	if !ok {
		return nil
	}
	return v.(Handler)
}

// GetLogger returns an logger by name
// if not, create one and add it to logger register
func GetLogger(name string, options ...Option) *Logger {
	if name == "" {
		name = RootLoggerName
	}

	v, ok := loggers.Get(name)
	if ok {
		return v.(*Logger)
	}

	options = append(options, OptionName(name))
	logger := NewLogger(options...)

	// check twice
	// maybe sb. adds logger when this logger is creating
	v, ok = loggers.Get(name)
	if ok {
		return v.(*Logger)
	}

	loggers.Register(name, logger)
	return logger
}

// GetLevel returns a Level registered with the given name
func GetLevel(name string) Level {
	v, ok := levels.Get(name)
	if !ok {
		return Level(-1)
	}
	return v.(Level)
}

// RegisterLevel binds name and level
func RegisterLevel(name string, level Level) {
	levels.Register(name, level)
	// add custom levels name
	levelNames[level] = name
}

// DisableExistingLoggers closes all existing loggers and unregister them
func DisableExistingLoggers() {
	// close all existing logger
	loggers.Lock()
	for _, logger := range loggers.Iter() {
		_logger := logger.(*Logger)
		_logger.Close()
	}
	loggers.Unlock()

	loggers.Clear()

	// reset root
	root = GetLogger(RootLoggerName)
	root.ApplyOptions(OptionHandlers(NewStreamHandler()))
}
