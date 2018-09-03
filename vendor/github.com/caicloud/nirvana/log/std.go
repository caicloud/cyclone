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

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"time"
)

// stdLogger writes logs to a stream.
type stdLogger struct {
	level   Level
	stream  io.Writer
	wrapper int
}

// NewStdLogger creates a stdandard logger for logging to stderr.
func NewStdLogger(level Level) Logger {
	return newStderrLogger(level, 0)
}

func newStderrLogger(level Level, wrapper int) *stdLogger {
	return newStdLogger(level, os.Stderr, wrapper)
}

func newStdLogger(level Level, stream io.Writer, wrapper int) *stdLogger {
	return &stdLogger{
		level:   level,
		stream:  stream,
		wrapper: wrapper,
	}
}

var _ Logger = &stdLogger{}

// V reports whether verbosity at the call site is at least the requested level.
// The returned value is a Verboser, which implements Info, Infof
// and Infoln. These methods will write to the Info log if called.
func (l *stdLogger) V(v Level) Verboser {
	if v > l.level {
		return silentLogger
	}
	return l
}

// Info logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *stdLogger) Info(a ...interface{}) {
	l.output(SeverityInfo, a...)
}

// Infof logs to the INFO log.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (l *stdLogger) Infof(format string, a ...interface{}) {
	l.outputf(SeverityInfo, format, a...)
}

// Infoln logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *stdLogger) Infoln(a ...interface{}) {
	l.outputln(SeverityInfo, a...)
}

// Warning logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (l *stdLogger) Warning(a ...interface{}) {
	l.output(SeverityWarning, a...)
}

// Warningf logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (l *stdLogger) Warningf(format string, a ...interface{}) {
	l.outputf(SeverityWarning, format, a...)
}

// Warningln logs to the WARNING logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *stdLogger) Warningln(a ...interface{}) {
	l.outputln(SeverityWarning, a...)
}

// Error logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (l *stdLogger) Error(a ...interface{}) {
	l.output(SeverityError, a...)
}

// Errorf logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (l *stdLogger) Errorf(format string, a ...interface{}) {
	l.outputf(SeverityError, format, a...)
}

// Errorln logs to the ERROR logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *stdLogger) Errorln(a ...interface{}) {
	l.outputln(SeverityError, a...)
}

// Fatal logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func (l *stdLogger) Fatal(a ...interface{}) {
	l.output(SeverityFatal, a...)
}

// Fatalf logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func (l *stdLogger) Fatalf(format string, a ...interface{}) {
	l.outputf(SeverityFatal, format, a...)
}

// Fatalln logs to the FATAL logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func (l *stdLogger) Fatalln(a ...interface{}) {
	l.outputln(SeverityFatal, a...)
}

// Clone clones current logger with new wrapper.
// A positive wrapper indicates how many wrappers outside the logger.
func (l *stdLogger) Clone(wrapper int) Logger {
	return &stdLogger{
		level:   l.level,
		stream:  l.stream,
		wrapper: l.wrapper + wrapper,
	}
}

func (l *stdLogger) output(severity Severity, a ...interface{}) {
	l.write(prefix(severity, l.wrapper+2) + fmt.Sprint(a...))
	l.exitIfFatal(severity)
}

func (l *stdLogger) outputf(severity Severity, format string, a ...interface{}) {
	l.write(prefix(severity, l.wrapper+2) + fmt.Sprintf(format, a...))
	l.exitIfFatal(severity)
}

func (l *stdLogger) outputln(severity Severity, a ...interface{}) {
	l.write(prefix(severity, l.wrapper+2) + fmt.Sprintln(a...))
	l.exitIfFatal(severity)
}

func (l *stdLogger) exitIfFatal(severity Severity) {
	if severity == SeverityFatal {
		os.Exit(1)
	}
}

func (l *stdLogger) write(data string) {
	if data[len(data)-1] != '\n' {
		data += "\n"
	}
	fmt.Fprint(l.stream, data)
}

func prefix(severity Severity, depth int) string {
	return fmt.Sprintf("%-5s %s %s | ", severity, time.Now().Format("0102-15:04:05.000-07"), file(1+depth))
}

func file(depth int) string {
	_, file, number, ok := runtime.Caller(1 + depth)
	if !ok {
		file = "???"
		number = 0
	} else {
		file = path.Base(file)
	}
	return fmt.Sprintf("%s:%d", file, number)
}
