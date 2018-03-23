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
	"sync"

	"github.com/zoumo/logdog/pkg/pythonic"
)

var (
	// Discard is an io.ReadWriteCloser on which all Read | Write | Close calls succeed
	// without doing anything.
	Discard = devNull(0)
)

type flusher interface {
	Sync() error
}

type flushWriter interface {
	io.Writer
	flusher
}

type flushWriteCloser interface {
	io.WriteCloser
	flusher
}

type devNull int

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

func (devNull) Read(p []byte) (n int, err error) {
	return len(p), nil
}

func (devNull) Sync() error {
	return nil
}

func (devNull) Close() error {
	return nil
}

// Handler specifies how to write a LoadConfig, appropriately formatted, to output.
type Handler interface {
	// Filter checks if handler should filter the specified record
	Filter(*LogRecord) bool
	// Emit log record to output - e.g. stderr or file
	Emit(*LogRecord)
	// Flush flushes the file system's in-memory copy of recently written data to disk.
	// Typically, calls the file.Sync()
	Flush() error
	// Close output stream, if not return error
	Close() error
}

// NullHandler is an example handler doing nothing
type NullHandler struct {
	Name string
}

// NewNullHandler returns a NullHandler
func NewNullHandler() *NullHandler {
	return &NullHandler{}
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (hdlr *NullHandler) LoadConfig(config map[string]interface{}) error {
	return nil
}

// Filter checks if handler should filter the specified record
func (hdlr *NullHandler) Filter(*LogRecord) bool {
	return true
}

// Emit log record to output - e.g. stderr or file
func (hdlr *NullHandler) Emit(*LogRecord) {
	// do nothing
}

// Flush flushes in-memory data to disk
func (hdlr *NullHandler) Flush() error {
	return nil
}

// Close output stream, if not return error
func (hdlr *NullHandler) Close() error {
	return nil
}

// StreamHandler is a handler which writes logging records,
// appropriately formatted, to a stream.
// Note that this handler does not close the stream,
// as os.Stdout or os.Stderr may be used.
type StreamHandler struct {
	Name      string
	Level     Level
	Formatter Formatter
	Output    flushWriter
	mu        sync.Mutex
}

// NewStreamHandler returns a new StreamHandler fully initialized
func NewStreamHandler(options ...Option) *StreamHandler {
	hdlr := &StreamHandler{
		Name:      "",
		Output:    os.Stderr,
		Formatter: TerminalFormatter,
		Level:     NothingLevel,
	}

	hdlr.ApplyOptions(options...)

	return hdlr
}

// ApplyOptions applys all option to StreamHandler
func (hdlr *StreamHandler) ApplyOptions(options ...Option) *StreamHandler {
	for _, opt := range options {
		opt.applyOption(hdlr)
	}
	return hdlr
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (hdlr *StreamHandler) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return err
	}

	hdlr.Name = config.MustGetString("name", "")

	hdlr.Level = GetLevel(config.MustGetString("level", "NOTHING"))

	_formatter := config.MustGetString("formatter", "terminal")
	formatter := GetFormatter(_formatter)
	if formatter == nil {
		return fmt.Errorf("can not find formatter: %s", _formatter)
	}
	hdlr.Formatter = formatter

	return nil
}

// Emit log record to output - e.g. stderr or file
func (hdlr *StreamHandler) Emit(record *LogRecord) {
	if hdlr.Output == nil || hdlr.Formatter == nil {
		panic("you should set output and fomatter before use this handler")
	}

	if hdlr.Filter(record) {
		return
	}

	hdlr.mu.Lock()
	defer hdlr.mu.Unlock()

	msg, err := hdlr.Formatter.Format(record)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format record failed, [%v]\n", err)
		return
	}
	fmt.Fprintln(hdlr.Output, msg)
}

// Filter checks if handler should filter the specified record
func (hdlr *StreamHandler) Filter(record *LogRecord) bool {
	return record.Level < hdlr.Level
}

// Flush flushes the file system's in-memory copy to disk
func (hdlr *StreamHandler) Flush() error {
	return hdlr.Output.Sync()
}

// Close output stream, if not return error
func (hdlr *StreamHandler) Close() error {
	hdlr.Output.Sync()
	return nil
}

// FileHandler is a handler similar to SteamHandler
// if specified file and it will close the file
type FileHandler struct {
	Name      string
	Level     Level
	Formatter Formatter
	Output    flushWriteCloser
	Path      string
	mu        sync.Mutex
}

// NewFileHandler returns a new FileHandler fully initialized
func NewFileHandler(options ...Option) *FileHandler {
	fh := &FileHandler{
		Output:    Discard,
		Name:      "",
		Level:     NothingLevel,
		Formatter: DefaultFormatter,
	}

	fh.ApplyOptions(options...)

	return fh
}

// ApplyOptions applys all option to StreamHandler
func (hdlr *FileHandler) ApplyOptions(options ...Option) *FileHandler {
	for _, opt := range options {
		opt.applyOption(hdlr)
	}
	return hdlr
}

// LoadConfig loads config from its input and
// stores it in the value pointed to by c
func (hdlr *FileHandler) LoadConfig(c map[string]interface{}) error {
	config, err := pythonic.DictReflect(c)
	if err != nil {
		return err
	}

	// get name
	hdlr.Name = config.MustGetString("name", "")

	// get path and file
	path := config.MustGetString("filename", "")
	hdlr.SetPath(path)

	// get level
	hdlr.Level = GetLevel(config.MustGetString("level", "NOTHING"))

	// get formatter
	_formatter := config.MustGetString("formatter", "default")
	formatter := GetFormatter(_formatter)
	if formatter == nil {
		return fmt.Errorf("can not find formatter: %s", _formatter)
	}
	hdlr.Formatter = formatter

	return nil
}

// SetPath opens file located in the path, if not, create it
func (hdlr *FileHandler) SetPath(path string) *FileHandler {
	if path == "" {
		panic("Should provide a valid file path")
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		panic(fmt.Sprintf("Can not open file %s", path))
	}

	hdlr.Path = path
	hdlr.Output = file

	return hdlr
}

// Emit log record to file
func (hdlr *FileHandler) Emit(record *LogRecord) {
	if hdlr.Output == nil || hdlr.Formatter == nil {
		panic("you should set output and fomatter before use this handler")
	}

	if hdlr.Filter(record) {
		return
	}

	hdlr.mu.Lock()
	defer hdlr.mu.Unlock()

	msg, err := hdlr.Formatter.Format(record)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Format record failed, [%v]\n", err)
		return
	}

	fmt.Fprintln(hdlr.Output, msg)
}

// Filter checks if handler should filter the specified record
func (hdlr *FileHandler) Filter(record *LogRecord) bool {
	return record.Level < hdlr.Level
}

// Flush flushes the file system's in-memory copy
// of recently written data to disk.
func (hdlr *FileHandler) Flush() error {
	if hdlr.Output == nil {
		return nil
	}
	return hdlr.Output.Sync()
}

// Close file, if not return error
func (hdlr *FileHandler) Close() error {
	if hdlr.Output == nil {
		return nil
	}
	return hdlr.Output.Close()
}

func init() {
	RegisterConstructor("NullHandler", func() ConfigLoader {
		return NewNullHandler()
	})
	RegisterConstructor("StreamHandler", func() ConfigLoader {
		return NewStreamHandler()
	})
	RegisterConstructor("FileHandler", func() ConfigLoader {
		return NewFileHandler()
	})

}
