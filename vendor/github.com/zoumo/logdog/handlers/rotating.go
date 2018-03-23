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
package handler

import (
	"bytes"
	"io"
	"os"

	"github.com/zoumo/logdog"
)

type RotatingFileHandler struct {
	logdog.FileHandler

	MaxLine int
	CurLine int

	MaxSize int
	CurSize int

	Daily bool
}

func NewRotatingFileHandler(name string, path string) {

}

func (hdlr RotatingFileHandler) shouldRollover(size int) bool {
	needed := (hdlr.MaxSize > 0 && (hdlr.CurSize+size) >= hdlr.MaxSize) ||
		(hdlr.MaxLine > 0 && (hdlr.CurLine+1) >= hdlr.MaxLine)
	return needed
}

func (hdlr RotatingFileHandler) doRollover() {

}

// Here is a faster line counter useing bytes.Count
// http://stackoverflow.com/questions/24562942/golang-how-do-i-determine-the-number-of-lines-in-a-file-efficiently
// benchmark:
// BenchmarkBuffioScan   500      6408963 ns/op     4208 B/op    2 allocs/op
// BenchmarkBytesCount   500      4323397 ns/op     8200 B/op    1 allocs/op
// BenchmarkBytes32k     500      3650818 ns/op     65545 B/op   1 allocs/op
func (hdlr RotatingFileHandler) countLine() (int, error) {
	file, err := os.Open(hdlr.Path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	buf := make([]byte, 32768)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}
