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

package service

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/caicloud/nirvana/definition"
)

type vc2 struct {
	vc
	contentType string
	data        string
}

func (v *vc2) Body() (reader io.ReadCloser, contentType string, ok bool) {
	return &file{[]byte(v.data), 0}, v.contentType, true
}

func TestConsumer(t *testing.T) {
	const data = `{"value":"test body"}`
	types := []string{
		definition.MIMEText,
		definition.MIMEJSON,
		definition.MIMEXML,
		definition.MIMEOctetStream,
		definition.MIMEURLEncoded,
		definition.MIMEFormData,
	}
	targets := []reflect.Type{
		reflect.TypeOf(""),
		reflect.TypeOf(([]byte)(nil)),
	}
	defaults := []interface{}{
		"",
		[]byte{},
	}
	g := &BodyParameterGenerator{}
	for _, ct := range types {
		for i, target := range targets {
			def := defaults[i]
			if err := g.Validate("test", def, target); err != nil {
				t.Fatal(err)
			}
			result, err := g.Generate(
				context.Background(),
				&vc2{
					contentType: ct,
					data:        data,
				},
				AllConsumers(),
				"test",
				target,
			)
			if err != nil {
				t.Fatal(err)
			}
			switch r := result.(type) {
			case string:
				if r != data {
					t.Fatalf("Generate wrong data: %v", r)
				}
			case []byte:
				if string(r) != data {
					t.Fatalf("Generate wrong data: %v", r)
				}
			}
		}
	}

}

func TestProducer(t *testing.T) {
	const data = `{"value":"test body"}`
	types := []string{
		definition.MIMEText,
		definition.MIMEJSON,
		definition.MIMEXML,
		definition.MIMEOctetStream,
	}
	values := []interface{}{
		data,
		[]byte(data),
	}
	for _, at := range types {
		producer := ProducerFor(at)
		if producer == nil {
			t.Fatalf("Can't find producer for accept type: %s", at)
		}
		for _, v := range values {
			w := bytes.NewBuffer(nil)
			if err := producer.Produce(w, v); err != nil {
				t.Fatal(err)
			}
			if data != w.String() {
				t.Fatalf("Producer %s writed wrong data: %s", at, w.Bytes())
			}
		}
	}
}

func TestConverterFor(t *testing.T) {
	tests := []struct {
		tpy     reflect.Type
		data    []string
		want    interface{}
		pointer bool
	}{
		{
			reflect.TypeOf(bool(false)),
			[]string{"true", "false"},
			true,
			false,
		},
		{
			reflect.TypeOf(int(0)),
			[]string{"1", "2"},
			1,
			false,
		},
		{
			reflect.TypeOf(int8(0)),
			[]string{"1", "2"},
			int8(1),
			false,
		},
		{
			reflect.TypeOf(int32(0)),
			[]string{"1", "2"},
			int32(1),
			false,
		},
		{
			reflect.TypeOf(int64(0)),
			[]string{"1", "2"},
			int64(1),
			false,
		},
		{
			reflect.TypeOf(uint(0)),
			[]string{"1", "2"},
			uint(1),
			false,
		},
		{
			reflect.TypeOf(uint8(0)),
			[]string{"1", "2"},
			uint8(1),
			false,
		},
		{
			reflect.TypeOf(uint16(0)),
			[]string{"1", "2"},
			uint16(1),
			false,
		},
		{
			reflect.TypeOf(uint64(0)),
			[]string{"1", "2"},
			uint64(1),
			false,
		},
		{
			reflect.TypeOf(float32(0)),
			[]string{"1.2", "2"},
			float32(1.2),
			false,
		},
		{
			reflect.TypeOf(float64(0)),
			[]string{"1.2", "2"},
			float64(1.2),
			false,
		},
		{
			reflect.TypeOf(string("")),
			[]string{"1", "2"},
			"1",
			false,
		},
		{
			reflect.TypeOf(new(bool)),
			[]string{"true", "2"},
			true,
			true,
		},
		{
			reflect.TypeOf(new(int)),
			[]string{"1", "2"},
			1,
			true,
		},
		{
			reflect.TypeOf(new(int8)),
			[]string{"1", "2"},
			int8(1),
			true,
		},
		{
			reflect.TypeOf(new(int16)),
			[]string{"1", "2"},
			int16(1),
			true,
		},
		{
			reflect.TypeOf(new(int32)),
			[]string{"1", "2"},
			int32(1),
			true,
		},
		{
			reflect.TypeOf(new(int64)),
			[]string{"1", "2"},
			int64(1),
			true,
		},
		{
			reflect.TypeOf(new(uint)),
			[]string{"1", "2"},
			uint(1),
			true,
		},
		{
			reflect.TypeOf(new(uint8)),
			[]string{"1", "2"},
			uint8(1),
			true,
		},

		{
			reflect.TypeOf(new(uint16)),
			[]string{"1", "2"},
			uint16(1),
			true,
		},

		{
			reflect.TypeOf(new(uint32)),
			[]string{"1", "2"},
			uint32(1),
			true,
		},

		{
			reflect.TypeOf(new(uint64)),
			[]string{"1", "2"},
			uint64(1),
			true,
		},
		{
			reflect.TypeOf(new(float32)),
			[]string{"1.2", "2"},
			float32(1.2),
			true,
		},
		{
			reflect.TypeOf(new(float64)),
			[]string{"1.2", "2"},
			float64(1.2),
			true,
		},
		{
			reflect.TypeOf(new(string)),
			[]string{"1.2", "2"},
			"1.2",
			true,
		},
		{
			reflect.TypeOf([]bool{}),
			[]string{"true", "false"},
			[]bool{true, false},
			false,
		},
		{
			reflect.TypeOf([]int{}),
			[]string{"1", "2"},
			[]int{1, 2},
			false,
		},
		{
			reflect.TypeOf([]float64{}),
			[]string{"1.2", "2.2"},
			[]float64{1.2, 2.2},
			false,
		},
		{
			reflect.TypeOf([]string{}),
			[]string{"1.2", "2.2"},
			[]string{"1.2", "2.2"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			converter := ConverterFor(tt.tpy)
			got, err := converter(context.TODO(), tt.data)

			if err != nil {
				t.Errorf("ConvertToXXX() type %v, error = %v, ", tt.tpy, err)
				return
			}
			if !tt.pointer && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertToXXX() type %v, got %v, want %v", tt.tpy, got, tt.want)
			}
			if tt.pointer {
				value := reflect.ValueOf(got).Elem().Interface()
				if !reflect.DeepEqual(value, tt.want) {
					t.Errorf("ConvertToXXX() type %v, got %v, want %v", tt.tpy, value, tt.want)
				}
			}
		})
	}

}
