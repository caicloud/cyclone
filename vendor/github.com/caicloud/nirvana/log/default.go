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

// default logger. Don't use the logger directly.
var logger Logger = newStderrLogger(LevelDebug, 1)

// DefaultLogger returns default logger.
func DefaultLogger() Logger {
	return logger.Clone(-1)
}

// SetDefaultLogger sets default logger.
func SetDefaultLogger(l Logger) {
	if l == nil {
		logger = &SilentLogger{}
	} else {
		logger = l.Clone(1)
	}
}

type defaultVerboser struct {
	verboser Verboser
}

// Info logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (v *defaultVerboser) Info(a ...interface{}) {
	v.verboser.Info(a...)
}

// Infof logs to the INFO log.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (v *defaultVerboser) Infof(format string, a ...interface{}) {
	v.verboser.Infof(format, a...)
}

// Infoln logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (v *defaultVerboser) Infoln(a ...interface{}) {
	v.verboser.Infoln(a...)
}

// V reports whether verbosity at the call site is at least the requested level.
// The returned value is a Verboser, which implements Info, Infof
// and Infoln. These methods will write to the Info log if called.
func V(v Level) Verboser {
	return &defaultVerboser{logger.V(v)}
}

// Info logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Info(a ...interface{}) {
	logger.Info(a...)
}

// Infof logs to the INFO log.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Infof(format string, a ...interface{}) {
	logger.Infof(format, a...)
}

// Infoln logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Infoln(a ...interface{}) {
	logger.Infoln(a...)
}

// Warning logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Warning(a ...interface{}) {
	logger.Warning(a...)
}

// Warningf logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Warningf(format string, a ...interface{}) {
	logger.Warningf(format, a...)
}

// Warningln logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Warningln(a ...interface{}) {
	logger.Warningln(a...)
}

// Error logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Error(a ...interface{}) {
	logger.Error(a...)
}

// Errorf logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Errorf(format string, a ...interface{}) {
	logger.Errorf(format, a...)
}

// Errorln logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Errorln(a ...interface{}) {
	logger.Errorln(a...)
}

// Fatal logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Fatal(a ...interface{}) {
	logger.Fatal(a...)
}

// Fatalf logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Fatalf(format string, a ...interface{}) {
	logger.Fatalf(format, a...)
}

// Fatalln logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Fatalln(a ...interface{}) {
	logger.Fatalln(a...)
}
