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
	"bytes"
	"fmt"
	"path"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestStdLogger(t *testing.T) {
	tr := newTester()
	l := newStdLogger(LevelDebug, tr.buf, 0)

	data := []interface{}{"123", 456, false}
	format := "%s %d-%v"

	l.Info(data)
	tr.test(t, SeverityInfo, data)
	l.Infof(format, data)
	tr.testf(t, SeverityInfo, format, data)
	l.Infoln(data)
	tr.testln(t, SeverityInfo, data)

	l.Warning(data)
	tr.test(t, SeverityWarning, data)
	l.Warningf(format, data)
	tr.testf(t, SeverityWarning, format, data)
	l.Warningln(data)
	tr.testln(t, SeverityWarning, data)

	l.Error(data)
	tr.test(t, SeverityError, data)
	l.Errorf(format, data)
	tr.testf(t, SeverityError, format, data)
	l.Errorln(data)
	tr.testln(t, SeverityError, data)

	l.V(1).Info(data)
	tr.test(t, SeverityInfo, data)
	l.V(1).Infof(format, data)
	tr.testf(t, SeverityInfo, format, data)
	l.V(1).Infoln(data)
	tr.testln(t, SeverityInfo, data)
}

func TestDefaultLogger(t *testing.T) {
	tr := newTester()
	logger = newStdLogger(LevelDebug, tr.buf, 1)

	data := []interface{}{"123", 456, false}
	format := "%s %d-%v"

	Info(data)
	tr.test(t, SeverityInfo, data)
	Infof(format, data)
	tr.testf(t, SeverityInfo, format, data)
	Infoln(data)
	tr.testln(t, SeverityInfo, data)

	Warning(data)
	tr.test(t, SeverityWarning, data)
	Warningf(format, data)
	tr.testf(t, SeverityWarning, format, data)
	Warningln(data)
	tr.testln(t, SeverityWarning, data)

	Error(data)
	tr.test(t, SeverityError, data)
	Errorf(format, data)
	tr.testf(t, SeverityError, format, data)
	Errorln(data)
	tr.testln(t, SeverityError, data)

	V(1).Info(data)
	tr.test(t, SeverityInfo, data)
	V(1).Infof(format, data)
	tr.testf(t, SeverityInfo, format, data)
	V(1).Infoln(data)
	tr.testln(t, SeverityInfo, data)
}

type tester struct {
	// Use the buffer as logger writer.
	buf *bytes.Buffer
}

func newTester() *tester {
	return &tester{
		buf: bytes.NewBuffer(nil),
	}
}

// The method must follow output method immediately.
func (tr *tester) test(t *testing.T, severity Severity, a ...interface{}) {
	tr.match(t, tr.buf.String(), severity, fixFilePos(), fixTrailingNewline(fmt.Sprint(a...)))
	tr.buf.Reset()
}

// The method must follow output method immediately.
func (tr *tester) testf(t *testing.T, severity Severity, format string, a ...interface{}) {
	tr.match(t, tr.buf.String(), severity, fixFilePos(), fixTrailingNewline(fmt.Sprintf(format, a...)))
	tr.buf.Reset()
}

// The method must follow output method immediately.
func (tr *tester) testln(t *testing.T, severity Severity, a ...interface{}) {
	tr.match(t, tr.buf.String(), severity, fixFilePos(), fixTrailingNewline(fmt.Sprintln(a...)))
	tr.buf.Reset()
}

func (tr *tester) match(t *testing.T, result string, severity Severity, file string, data string) {
	results := strings.SplitN(result, " | ", 2)
	if len(results) != 2 {
		t.Fatalf("Can't find | from result: %s", result)
	}
	if results[1] != data {
		t.Fatalf("Data does not match: \n%s\n%s", results[1], data)
	}
	re := fmt.Sprintf(`^%-5s (\d{4}-\d{2}:\d{2}:\d{2}\.\d{3}[+-]\d{2}) %s$`, severity, file)
	reg, err := regexp.Compile(re)
	if err != nil {
		t.Fatalf("Regexp is wrong: %s, %v", re, err)
	}
	t.Log(results[0])
	matches := reg.FindStringSubmatch(results[0])
	if len(matches) != 2 {
		t.Fatalf("Can't match log header: %s, %s", results[0], re)
	}
	_, err = time.Parse("0102-15:04:05.000-07", matches[1])
	if err != nil {
		t.Fatalf("Time format is wrong: %s, %v", matches[1], err)
	}
}

func fixFilePos() string {
	_, file, number, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		number = 0
	} else {
		file = path.Base(file)
		number--
	}
	return fmt.Sprintf("%s:%d", file, number)
}

func fixTrailingNewline(data string) string {
	if data[len(data)-1] != '\n' {
		data += "\n"
	}
	return data
}
