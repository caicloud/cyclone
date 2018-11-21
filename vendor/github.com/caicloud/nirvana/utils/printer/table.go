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

package printer

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/text/width"
)

// Table prints a table for lines.
type Table struct {
	lines       [][]string
	columes     []int
	maxColWidth int
}

// NewTable creates a table printer.
func NewTable(maxColWidth int) *Table {
	return &Table{
		lines:       make([][]string, 0),
		columes:     make([]int, 0),
		maxColWidth: maxColWidth,
	}
}

// AddRow adds a line to current table.
func (p *Table) AddRow(values ...interface{}) {
	line := make([]string, 0, len(values))
	for i, v := range values {
		str := fmt.Sprint(v)
		length := MaxWidthForLines(str)
		if i >= len(p.columes) {
			p.columes = append(p.columes, length)
		} else if length > p.columes[i] {
			p.columes[i] = length
		}
		line = append(line, str)
	}
	p.lines = append(p.lines, line)
}

// String returns table string.
func (p *Table) String() string {
	buf := bytes.NewBuffer(nil)
	for _, line := range p.lines {
		cells := make([]*cell, 0, len(line))
		for _, s := range line {
			cells = append(cells, newCell(s))
		}
		hasData := true
		for hasData {
			hasData = false
			for i, cell := range cells {
				colWidth := p.columes[i] + 2
				if colWidth > p.maxColWidth {
					colWidth = p.maxColWidth
				}
				str := cell.NextLine(colWidth - 2)
				if _, err := buf.WriteString(str); err != nil {
					panic(err)
				}
				for i := LengthForString(str); i < colWidth; i++ {
					if err := buf.WriteByte(' '); err != nil {
						panic(err)
					}
				}
				if !hasData && !cell.Empty() {
					hasData = true
				}
			}
			if err := buf.WriteByte('\n'); err != nil {
				panic(err)
			}
		}
	}
	return buf.String()
}

// cell represents a table cell.
type cell struct {
	str   []rune
	start int
}

func newCell(str string) *cell {
	return &cell{
		str: []rune(str),
	}
}

// Empty returns whether current cell is empty or not.
func (c *cell) Empty() bool {
	return c.start >= len(c.str)
}

// NextLine returns next line limited by width.
func (c *cell) NextLine(width int) string {
	if c.Empty() || width <= 0 {
		return ""
	}
	end := c.start
	for ; end < len(c.str); end++ {
		length := LengthForRune(c.str[end])
		if length == 0 {
			// Every contol char can cause breaking.
			break
		}
		width -= length
		if width < 0 {
			break
		}
	}
	result := c.str[c.start:end]
	// Ignore trailing control chars.
	for ; end < len(c.str); end++ {
		if LengthForRune(c.str[end]) != 0 {
			break
		}
	}
	c.start = end
	return string(result)
}

// MaxWidthForLines gets max line width for a multi-lines string.
func MaxWidthForLines(str string) int {
	lines := strings.Split(str, "\n")
	maxLength := 0
	for _, line := range lines {
		length := LengthForString(line)
		if length > maxLength {
			maxLength = length
		}
	}
	return maxLength
}

// LengthForString gets length for a string.
func LengthForString(str string) int {
	length := 0
	for _, r := range str {
		length += LengthForRune(r)
	}
	return length
}

// LengthForRune gets char length for a rune.
// All control chars are treated as zero-length.
func LengthForRune(r rune) int {
	if r < 32 || r == 127 {
		return 0
	}
	switch width.LookupRune(r).Kind() {
	case width.EastAsianFullwidth, width.EastAsianWide:
		return 2
	}
	return 1
}
