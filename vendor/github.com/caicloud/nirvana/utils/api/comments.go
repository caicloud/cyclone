/*
Copyright 2018 Caicloud Authors

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

package api

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	// CommentsOptionDescriptors is the option name of descriptors.
	CommentsOptionDescriptors = "descriptors"
	// CommentsOptionModifiers is the option name of modifiers.
	CommentsOptionModifiers = "modifiers"
	// CommentsOptionAlias is the option name of alias.
	CommentsOptionAlias = "alias"
	// CommentsOptionOrigin is the option name of original name.
	CommentsOptionOrigin = "origin"
)

// Comments is parsed from go comments.
type Comments struct {
	lines   []string
	options map[string][]string
}

var optionsRegexp = regexp.MustCompile(`^[ \t]*\+nirvana:api[ \t]*=(.*)$`)
var options = []string{CommentsOptionDescriptors, CommentsOptionModifiers, CommentsOptionAlias}

// ParseComments parses comments and extracts nirvana options.
func ParseComments(comments string) *Comments {
	c := &Comments{
		options: map[string][]string{},
	}
	comments = strings.TrimSpace(comments)
	lines := strings.Split(comments, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			matches := optionsRegexp.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) != 2 {
					continue
				}
				tag := reflect.StructTag(match[1])
				for _, option := range options {
					value := tag.Get(option)
					if value != "" {
						opt := c.options[option]
						if opt == nil {
							opt = []string{}
						}
						opt = append(opt, value)
						c.options[option] = opt
					}
				}
			}
			if len(matches) > 0 {
				continue
			}
			c.lines = append(c.lines, line)
		} else if len(c.lines) > 0 && c.lines[len(c.lines)-1] != "" {
			c.lines = append(c.lines, line)
		}
	}
	if len(c.lines) > 0 && c.lines[len(c.lines)-1] == "" {
		c.lines = c.lines[:len(c.lines)-1]
	}
	return c
}

// BlockComments returns block style comments.
func (c *Comments) BlockComments() string {
	buf := NewBuffer()
	buf.Write("/*\n")
	for _, line := range c.lines {
		buf.Writef("%s\n", line)
	}
	if len(c.options) > 0 {
		buf.Writeln()
	}
	for option, values := range c.options {
		for _, value := range values {
			buf.Writef("+nirvana:api=%s:%s\n", option, strconv.Quote(value))
		}
	}
	buf.Write("*/")
	return buf.String()
}

// LineComments returns line style comments.
func (c *Comments) LineComments() string {
	if len(c.lines) <= 0 {
		return ""
	}
	buf := NewBuffer()
	for _, line := range c.lines {
		buf.Writef("// %s\n", line)
	}
	if len(c.options) > 0 {
		buf.Write("//\n")
	}
	for option, values := range c.options {
		for _, value := range values {
			buf.Writef("// +nirvana:api=%s:%s\n", option, strconv.Quote(value))
		}
	}
	return buf.String()
}

// Rename replaces the first word of this comments. If its first word is
// not the same as origin, the method returns false.
func (c *Comments) Rename(origin, target string) bool {
	if len(c.lines) > 0 {
		line := c.lines[0]
		if strings.HasPrefix(line, origin) {
			line = target + line[len(origin):]
			c.lines[0] = line
			return true
		}
	}
	return false
}

// Options returns all options.
func (c *Comments) Options() map[string][]string {
	return c.options
}

// Option returns values of an option.
func (c *Comments) Option(name string) []string {
	return c.options[name]
}

// AddOption adds an option.
func (c *Comments) AddOption(name, value string) {
	c.options[name] = append(c.options[name], value)
}

// CleanOptions removes all options.
func (c *Comments) CleanOptions() {
	c.options = map[string][]string{}
}

// String returns comments.
func (c *Comments) String() string {
	return strings.Join(c.lines, "\n")
}

// Buffer provides a buffer to write data. The buffer will panic if
// an error occurs.
type Buffer struct {
	buf *bytes.Buffer
}

// NewBuffer creates a buffer.
func NewBuffer() *Buffer {
	return &Buffer{bytes.NewBuffer(nil)}
}

// Write writes data to this buffer.
func (b *Buffer) Write(a ...interface{}) {
	_, err := fmt.Fprint(b.buf, a...)
	if err != nil {
		panic(err)
	}
}

// Writef writes data with format to this buffer.
func (b *Buffer) Writef(format string, a ...interface{}) {
	_, err := fmt.Fprintf(b.buf, format, a...)
	if err != nil {
		panic(err)
	}
}

// Writeln writes data with a newline to this buffer.
func (b *Buffer) Writeln(a ...interface{}) {
	_, err := fmt.Fprintln(b.buf, a...)
	if err != nil {
		panic(err)
	}
}

// Bytes returns a slice of length b.Len() holding the unread portion of the buffer.
func (b *Buffer) Bytes() []byte {
	return b.buf.Bytes()
}

// String returns the contents of the unread portion of the buffer
// as a string. If the Buffer is a nil pointer, it returns "<nil>".
func (b *Buffer) String() string {
	return b.buf.String()
}
