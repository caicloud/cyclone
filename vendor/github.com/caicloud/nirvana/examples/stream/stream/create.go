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

package stream

import (
	"bytes"
	"context"
	"io"
)

func Stream(ctx context.Context) (io.Reader, error) {
	// Why 2049?
	// HTTP response writer has a buffer for caching data.
	// The capacity is 2048 bytes.
	// If the data overflow the buffer, It automatically set
	// HTTP header `Transfer-Encoding` to chunked and transfer
	// data in chunked mode.
	data := make([]byte, 2049)
	for i := 0; i < len(data); i++ {
		data[i] = 'a' + byte(i%26)
	}
	return bytes.NewReader(data), nil
}
