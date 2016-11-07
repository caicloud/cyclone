/*
Copyright 2016 caicloud authors. All rights reserved.

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

package filebuffer

import (
	"os"

	"github.com/djherbis/buffer"
)

// FileBuffer is a thin wrapper around buffer.Buffer. It exposes a Close method to
// close the underline file.
type FileBuffer interface {
	buffer.Buffer
	Name() string
	Stat() (fi os.FileInfo, err error)
	Close() error
}

type fileBuffer struct {
	buffer.BufferAt
	file buffer.File
}

// NewFileBuffer returns a FileBuffer.
func NewFileBuffer(N int64, file buffer.File) FileBuffer {
	return &fileBuffer{buffer.NewFile(N, file), file}
}

// Close closes the buffer file.
func (buf *fileBuffer) Close() error {
	return buf.file.Close()
}

// Name returns the name of the buffer file.
func (buf *fileBuffer) Name() string {
	return buf.file.Name()
}

// Stat returns the status of the buffer file.
func (buf *fileBuffer) Stat() (fi os.FileInfo, err error) {
	return buf.file.Stat()
}
