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

package method

import (
	"reflect"
	"testing"
)

type TestInterface interface {
	Number() int
}

type TestStruct struct {
	a int
}

func (t *TestStruct) Number() int {
	return t.a
}

type TestStruct2 struct {
	a int
}

func (t *TestStruct2) Number() int {
	return t.a * 2
}

func TestContainer(t *testing.T) {
	var ts TestInterface = &TestStruct2{100}
	cases := []struct {
		typ      interface{}
		instance interface{}
		method   string
		result   int
	}{
		{&TestStruct{}, &TestStruct{100}, "Number", 100},
		{(*TestInterface)(nil), &TestStruct{100}, "Number", 100},
		{(*TestInterface)(nil), TestInterface(&TestStruct{100}), "Number", 100},
		{(*TestInterface)(nil), ts, "Number", 200},
	}
	for _, c := range cases {
		f := Get(c.typ, c.method)
		if defaultContainer.typeOf(c.typ).Kind() == reflect.Interface {
			PutInterface(c.typ, c.instance)
		} else {
			Put(c.instance)
		}
		result := f.(func() int)()
		if result != c.result {
			t.Fatalf("Function result should be %d, but got: %d", c.result, result)
		}
	}
}

func BenchmarkContainer(b *testing.B) {
	c := NewContainer()
	f := c.Get(&TestStruct{}, "Number")
	c.Put(&TestStruct{100})
	fn := reflect.ValueOf(f)
	for i := 0; i < b.N; i++ {
		fn.Call([]reflect.Value{})
	}
}

func BenchmarkDirect(b *testing.B) {
	fn := reflect.ValueOf(&TestStruct{100}).MethodByName("Number")
	for i := 0; i < b.N; i++ {
		fn.Call([]reflect.Value{})
	}
}
