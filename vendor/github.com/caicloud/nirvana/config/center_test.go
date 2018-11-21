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

package config

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

var config = `
nirvana:
  ip: 0.0.0.0
  port: 9999

custom:
  bool: true
  int: 
    int8: 127
    int16: 32767
    int32: 2147483647
    int64: 9223372036854775807
  uint:
    uint8: 255
    uint16: 65535
    uint32: 4294967295
    uint64: 18446744073709551615
  float:
    float32: 3.4e38
    float64: 1.7e308
  strings:
  - a
  - b
  - c
`

type Field struct {
	Key   string
	Value interface{}
	Func  interface{}
}

func TestConfigCenter(t *testing.T) {
	v.SetConfigType("yaml")
	if err := v.MergeConfig(strings.NewReader(config)); err != nil {
		t.Fatal(err)
	}
	fields := []Field{
		{"nirvana.ip", "0.0.0.0", GetString},
		{"nirvana.port", 9999, GetInt},
		{"custom.bool", true, GetBool},
		{"custom.int.int8", int8(math.MaxInt8), GetInt8},
		{"custom.int.int16", int16(math.MaxInt16), GetInt16},
		{"custom.int.int32", int32(math.MaxInt32), GetInt32},
		{"custom.int.int64", int64(math.MaxInt64), GetInt64},
		{"custom.int.int64", int(math.MaxInt64), GetInt},
		{"custom.uint.uint8", uint8(math.MaxUint8), GetUint8},
		{"custom.uint.uint16", uint16(math.MaxUint16), GetUint16},
		{"custom.uint.uint32", uint32(math.MaxUint32), GetUint32},
		{"custom.uint.uint64", uint64(math.MaxUint64), GetUint64},
		{"custom.uint.uint64", uint(math.MaxUint64), GetUint},
		{"custom.float.float32", float32(3.4e38), GetFloat32},
		{"custom.float.float64", float64(1.7e308), GetFloat64},
		{"custom.strings", []string{"a", "b", "c"}, GetStringSlice},
	}
	for _, f := range fields {
		if err := testGet(f.Key, f.Value, f.Func); err != nil {
			t.Fatal(err)
		}
	}

	if !IsSet("nirvana.ip") {
		t.Fatal("nirvana.ip is set but got wrong")
	}
	if IsSet("nirvana.ip2") {
		t.Fatal("nirvana.ip2 is not set but got wrong")
	}
	if Get("nirvana.ip") == nil {
		t.Fatal("nirvana.ip has value but got wrong")
	}
	if Get("nirvana.ip2") != nil {
		t.Fatal("nirvana.ip2 has no value but got wrong")
	}
	Set("nirvana.ip2", "1.2.3.4")
	if !reflect.DeepEqual(Get("nirvana.ip2"), "1.2.3.4") {
		t.Fatal("nirvana.ip2 has been set but got wrong")
	}
}

func testGet(key string, desired interface{}, f interface{}) error {
	fv := reflect.ValueOf(f)
	result := fv.Call([]reflect.Value{reflect.ValueOf(key)})[0].Interface()
	if !reflect.DeepEqual(result, desired) {
		return fmt.Errorf("Wrong result for %s: %v", key, result)
	}
	return nil
}
