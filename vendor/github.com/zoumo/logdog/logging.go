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

import "sync"

const (
	// RootLoggerName is the name of root logger
	RootLoggerName = "root"
)

var (
	mu = sync.Mutex{}
	// set default logger
	root = GetLogger(RootLoggerName, OptionHandlers(NewStreamHandler()))
)

// AddHandlers is an alias of root.AddHandler
func AddHandlers(handlers ...Handler) *Logger {
	root.AddHandlers(handlers...)
	return root
}

// ApplyOptions is an alias of root.ApplyOptions
func ApplyOptions(options ...Option) *Logger {
	root.ApplyOptions(options...)
	return root
}

// Flush ...
func Flush() error {
	return root.Flush()
}

// Debugf is an alias of root.Debugf
func Debugf(msg string, args ...interface{}) {
	root.log(DebugLevel, msg, args...)
}

// Infof is an alias of root.Infof
func Infof(msg string, args ...interface{}) {
	root.log(InfoLevel, msg, args...)
}

// Warningf is an alias of root.Warningf
func Warningf(msg string, args ...interface{}) {
	root.log(WarnLevel, msg, args...)
}

// Warnf is an alias of root.Warnf
func Warnf(msg string, args ...interface{}) {
	root.log(WarnLevel, msg, args...)
}

// Errorf is an alias of root.Errorf
func Errorf(msg string, args ...interface{}) {
	root.log(ErrorLevel, msg, args...)
}

// Noticef is an alias of root.Noticef
func Noticef(msg string, args ...interface{}) {
	root.log(NoticeLevel, msg, args...)
}

// Fatalf is an alias of root.Criticalf
func Fatalf(msg string, args ...interface{}) {
	root.log(FatalLevel, msg, args...)
}

// Panicf is an alias of root.Panicf
func Panicf(msg string, args ...interface{}) {
	root.log(FatalLevel, msg, args...)
	panic("CRITICAL")
}

// Debug is an alias of root.Debug
func Debug(args ...interface{}) {
	root.log(DebugLevel, "", args...)
}

// Info is an alias of root.Info
func Info(args ...interface{}) {
	root.log(InfoLevel, "", args...)
}

// Warning is an alias of root.Warning
func Warning(args ...interface{}) {
	root.log(WarnLevel, "", args...)
}

// Warn is an alias of root.Warn
func Warn(args ...interface{}) {
	root.log(WarnLevel, "", args...)
}

// Error is an alias of root.Error
func Error(args ...interface{}) {
	root.log(ErrorLevel, "", args...)
}

// Notice is an alias of root.Notice
func Notice(args ...interface{}) {
	root.log(NoticeLevel, "", args...)
}

// Fatal is an alias of root.Critical
func Fatal(args ...interface{}) {
	root.log(FatalLevel, "", args...)
}

// Panic an alias of root.Panic
func Panic(msg string, args ...interface{}) {
	root.log(FatalLevel, "", args...)
	panic("CRITICAL")
}
