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

import "os"

// SilentLogger logs nothing.
type SilentLogger struct{}

var silentLogger Logger = &SilentLogger{}

// V reports whether verbosity at the call site is at least the requested level.
// The returned value is a Verboser, which implements Info, Infof
func (l *SilentLogger) V(Level) Verboser { return l }

// Info logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (*SilentLogger) Info(...interface{}) {}

// Infof logs to the INFO log.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (*SilentLogger) Infof(string, ...interface{}) {}

// Infoln logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (*SilentLogger) Infoln(...interface{}) {}

// Warning logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (*SilentLogger) Warning(...interface{}) {}

// Warningf logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (*SilentLogger) Warningf(string, ...interface{}) {}

// Warningln logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (*SilentLogger) Warningln(...interface{}) {}

// Error logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (*SilentLogger) Error(...interface{}) {}

// Errorf logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (*SilentLogger) Errorf(string, ...interface{}) {}

// Errorln logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (*SilentLogger) Errorln(...interface{}) {}

// Fatal logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (*SilentLogger) Fatal(v ...interface{}) {
	os.Exit(1)
}

// Fatalf logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (*SilentLogger) Fatalf(string, ...interface{}) {
	os.Exit(1)
}

// Fatalln logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (*SilentLogger) Fatalln(v ...interface{}) {
	os.Exit(1)
}

// Clone clones current logger with new wrapper.
// A positive wrapper indicates how many wrappers outside the logger.
func (l *SilentLogger) Clone(wrapper int) Logger {
	return l
}
