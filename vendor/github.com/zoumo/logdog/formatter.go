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
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"sync"
	"syscall"

	"github.com/zoumo/logdog/pkg/pythonic"
	"github.com/zoumo/logdog/pkg/when"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	// ForceColor forces formatter use color output
	ForceColor = false
)

// Formatter is an interface which can convert a LogRecord to string
type Formatter interface {
	Format(*LogRecord) (string, error)
	Option
}

// FormatTime returns the creation time of the specified LogRecord as formatted text.
func FormatTime(record *LogRecord, datefmt string) string {
	if datefmt == "" {
		datefmt = DefaultDateFmtTemplate
	}
	return when.Strftime(&record.Time, datefmt)
}

// TextFormatter is the default formatter used to convert a LogRecord to text.
//
// The Formatter can be initialized with a format string which makes use of
// knowledge of the LogRecord attributes - e.g. the default value mentioned
// above makes use of the fact that the user's message and arguments are pre-
// formatted into a LogRecord's message attribute. Currently, the useful
// attributes in a LogRecord are described by:
//
// %(name)            Name of the logger (logging channel)
// %(levelno)         Numeric logging level for the message (DEBUG, INFO,
//                    WARNING, ERROR, CRITICAL)
// %(levelname)       Text logging level for the message ("DEBUG", "INFO",
//                    "WARNING", "ERROR", "CRITICAL")
// %(pathname)        Full pathname of the source file where the logging
//                    call was issued (if available) or maybe ??
// %(filename)        Filename portion of pathname
// %(lineno)          Source line number where the logging call was issued
//                    (if available)
// %(funcname)        Function name of caller or maybe ??
// %(time)            Textual time when the LogRecord was created
// %(message)         The result of record.getMessage(), computed just as
//                    the record is emitted
// %(color)           Print color
// %(endColor)        Reset color
type TextFormatter struct {
	fieldSequence []string
	fmtTeplate    string
	Fmt           string
	DateFmt       string
	EnableColors  bool
	mu            sync.Mutex
	ConfigLoader
}

const (
	// DefaultFmtTemplate is the default log string format value for TextFormatter
	DefaultFmtTemplate = "%(time) %(color)%(levelname)%(endColor) %(filename):%(lineno) | %(message)"
	// DefaultDateFmtTemplate is the default log time string format value for TextFormatter
	DefaultDateFmtTemplate = "%Y-%m-%d %H:%M:%S"
	// colors
	blue      = 34
	green     = 32
	yellow    = 33
	red       = 31
	darkGreen = 36
	white     = 37
)

var (
	// LogRecordFieldRegexp is the field regexp
	// for example, I will replace %(name) of real record name
	// TODO support %[(name)][flags][width].[precision]typecode
	LogRecordFieldRegexp = regexp.MustCompile(`\%\(\w+\)`)
	// DefaultFormatter is the default formatter of TextFormatter without color
	DefaultFormatter = &TextFormatter{
		Fmt:     DefaultFmtTemplate,
		DateFmt: DefaultDateFmtTemplate,
	}
	// TerminalFormatter is an TextFormatter with color
	TerminalFormatter = &TextFormatter{
		Fmt:          DefaultFmtTemplate,
		DateFmt:      DefaultDateFmtTemplate,
		EnableColors: true,
	}

	// ColorHash describes colors of different log level
	// you can add new color for your own log level
	ColorHash = map[Level]int{
		DebugLevel:  blue,
		InfoLevel:   green,
		WarnLevel:   yellow,
		ErrorLevel:  red,
		NoticeLevel: darkGreen,
		FatalLevel:  red,
	}

	// check if stderr is terminal, sometimes it is redirected to a file
	// isTerminal      = terminal.IsTerminal(syscall.Stderr)
	isTerminal      = terminal.IsTerminal(syscall.Stderr)
	isColorTerminal = isTerminal && (runtime.GOOS != "windows")
)

//IsColorTerminal return isTerminal and isColorTerminal
func IsColorTerminal() (bool, bool) {
	return isTerminal, isColorTerminal
}

// colorHash returns color for deferent level, default is white
func colorHash(level Level) (string, string) {
	// http://blog.csdn.net/acmee/article/details/6613060
	color, ok := ColorHash[level]
	if !ok {
		color = white // white
	}
	return fmt.Sprintf("\033[%dm", color), "\033[0m"
}

// NewTextFormatter return a new TextFormatter with default config
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		Fmt:          DefaultFmtTemplate,
		DateFmt:      DefaultDateFmtTemplate,
		EnableColors: false,
	}
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (tf *TextFormatter) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return err
	}

	tf.Fmt = config.MustGetString("fmt", DefaultFmtTemplate)
	tf.DateFmt = config.MustGetString("datefmt", DefaultDateFmtTemplate)
	tf.EnableColors = config.MustGetBool("enableColors", false)

	return nil

}

// parse shoule run only once
func (tf *TextFormatter) parse() {
	if tf.fmtTeplate != "" && len(tf.fieldSequence) > 0 {
		return
	}

	tf.mu.Lock()
	defer tf.mu.Unlock()
	if tf.Fmt == "" {
		tf.Fmt = DefaultFmtTemplate
	}

	// append fields to Fmt no matter what it is
	tf.Fmt += "%(fields)"

	// replace %(field) with %s && add field name to sequence
	// e.g. covert %(name) %(message) to %s %s
	tf.fmtTeplate = LogRecordFieldRegexp.ReplaceAllStringFunc(tf.Fmt, func(match string) string {
		// match : %(field)
		field := match[2 : len(match)-1]
		tf.fieldSequence = append(tf.fieldSequence, field)
		return "%s"
	})

}

func (tf *TextFormatter) getColor(record *LogRecord) (string, string) {
	color, endColor := "", ""
	if ForceColor || (isColorTerminal && tf.EnableColors) {
		color, endColor = colorHash(record.Level)
	}
	return color, endColor
}

// Format converts the specified record to string.
// bench mark with 10 fields
// go template            33153 ns/op
// ReplaceAllStringFunc    8420 ns/op
// field sequence          5046 ns/op
func (tf *TextFormatter) Format(record *LogRecord) (string, error) {

	if tf.Fmt == "" {
		// Don't open color printing by default
		tf.EnableColors = false
		tf.Fmt = DefaultFmtTemplate
	}

	tf.parse()

	color, endColor := tf.getColor(record)

	sequnce := make([]interface{}, 0, 20)

	for _, field := range tf.fieldSequence {
		switch field {
		case "name":
			sequnce = append(sequnce, record.Name)
		case "time":
			sequnce = append(sequnce, FormatTime(record, tf.DateFmt))
		case "levelno":
			sequnce = append(sequnce, fmt.Sprintf("%d", record.Level))
		case "levelname":
			sequnce = append(sequnce, fmt.Sprintf("%6s", record.LevelName))
		case "pathname":
			sequnce = append(sequnce, record.PathName)
		case "filename":
			sequnce = append(sequnce, record.FileName)
		case "funcname":
			sequnce = append(sequnce, record.ShortFuncName)
		case "lineno":
			sequnce = append(sequnce, fmt.Sprintf("%d", record.Line))
		case "message":
			sequnce = append(sequnce, record.GetMessage())
		case "color":
			sequnce = append(sequnce, color)
		case "endColor":
			sequnce = append(sequnce, endColor)
		case "fields":
			sequnce = append(sequnce, record.Fields.ToKVString(color, endColor))
		}
	}
	return fmt.Sprintf(tf.fmtTeplate, sequnce...), nil
}

// JSONFormatter can convert LogRecord to json text
type JSONFormatter struct {
	Datefmt string
	ConfigLoader
}

// NewJSONFormatter returns a JSONFormatter with default config
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		Datefmt: DefaultDateFmtTemplate,
	}
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (jf *JSONFormatter) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return err
	}

	jf.Datefmt = config.MustGetString("datefmt", DefaultDateFmtTemplate)
	return nil
}

// Format converts the specified record to json string.
func (jf *JSONFormatter) Format(record *LogRecord) (string, error) {
	fields := make(Fields, len(record.Fields)+4)
	for k, v := range record.Fields {
		fields[k] = v
	}
	// jf.formatFields(fields)
	data := make(map[string]interface{})

	data["time"] = FormatTime(record, jf.Datefmt)
	data["message"] = record.GetMessage()
	data["file"] = record.FileName
	data["line"] = record.Line
	data["level"] = record.LevelName
	if len(fields) > 0 {
		data["_fields"] = fields
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Marashal fields to Json failed, [%v]", err)
	}

	return string(jsonBytes), nil
}

func init() {
	RegisterConstructor("TextFormatter", func() ConfigLoader {
		return NewTextFormatter()
	})
	RegisterConstructor("JsonFormatter", func() ConfigLoader {
		return NewJSONFormatter()
	})

	RegisterFormatter("default", DefaultFormatter)
	RegisterFormatter("terminal", TerminalFormatter)
}
