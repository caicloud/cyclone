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

package main

import (
	"fmt"
	"io"

	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/examples/stream/api/v1"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

type SomeDataProducer struct {
}

// ContentType returns a HTTP MIME type.
func (p *SomeDataProducer) ContentType() string {
	return "application/somedata"
}

// Produce marshals v to data and write to w.
func (p *SomeDataProducer) Produce(w io.Writer, v interface{}) error {
	// The result must implement a reader.
	source, ok := v.(io.Reader)
	if !ok {
		return fmt.Errorf("type of data for %s must be a reader", p.ContentType())
	}
	_, err := io.Copy(w, source)
	return err
}

func main() {
	// Register a producer for content type 'application/somedata'
	if err := service.RegisterProducer(&SomeDataProducer{}); err != nil {
		panic(err)
	}

	cmd := config.NewDefaultNirvanaCommand()
	if err := cmd.Execute(v1.Descriptor()); err != nil {
		log.Fatal(err)
	}
}
