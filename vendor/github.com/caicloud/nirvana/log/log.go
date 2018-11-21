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

package log

import "math"

// Level is log level of verboser. We strongly recommend you to
// follow the rules:
// 1. Use LevelDebug for debug log.
// 2. Use 1 for important but not frequent log.
// 3. Use 2 for important and frequent log.
// 4. Use 3 for not important and not frequent log.
// 5. Use 4 for not important but frequent log.
// 6. Use [5, LevelDebug) only if you want to custom log levels.
type Level int32

const (
	// LevelDebug is for debug info.
	LevelDebug Level = math.MaxInt32
)

// Severity has four classes to correspond with log situation.
type Severity string

const (
	// SeverityInfo is for usual log.
	SeverityInfo Severity = "INFO"
	// SeverityWarning is for warning.
	SeverityWarning Severity = "WARN"
	// SeverityError is for error.
	SeverityError Severity = "ERROR"
	// SeverityFatal is for panic error. The severity means that
	// a logger will output the error and exit immediately.
	// It can't be recovered.
	SeverityFatal Severity = "FATAL"
)

// Verboser is an interface type that implements Info(f|ln) .
// See the documentation of V for more information.
type Verboser interface {
	// Info logs to the INFO log.
	// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
	Info(...interface{})
	// Infof logs to the INFO log.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Infof(string, ...interface{})
	// Infoln logs to the INFO log.
	// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
	Infoln(...interface{})
}

// Logger provides a set of methods to output log.
type Logger interface {
	Verboser
	// V reports whether verbosity at the call site is at least the requested level.
	// The returned value is a Verboser, which implements Info, Infof
	// and Infoln. These methods will write to the Info log if called.
	V(Level) Verboser
	// Warning logs to the WARNING logs.
	// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
	Warning(...interface{})
	// Warningf logs to the WARNING logs.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Warningf(string, ...interface{})
	// Warningln logs to the WARNING logs.
	// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
	Warningln(...interface{})
	// Error logs to the ERROR logs.
	// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
	Error(...interface{})
	// Errorf logs to the ERROR logs.
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Errorf(string, ...interface{})
	// Errorln logs to the ERROR logs.
	// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
	Errorln(...interface{})
	// Fatal logs to the FATAL logs, then calls os.Exit(1).
	// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
	Fatal(...interface{})
	// Fatalf logs to the FATAL logs, then calls os.Exit(1).
	// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
	Fatalf(string, ...interface{})
	// Fatalln logs to the FATAL logs, then calls os.Exit(1).
	// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
	Fatalln(...interface{})
	// Clone clones current logger with new wrapper.
	// A positive wrapper indicates how many wrappers outside the logger.
	// A negative wrapper indicates how many wrappers should be stripped.
	Clone(wrapper int) Logger
}
