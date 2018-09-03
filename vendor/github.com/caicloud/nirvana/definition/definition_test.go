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

package definition

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestParameter(t *testing.T) {
	d := Parameter{
		Source:      Path,
		Name:        "test",
		Default:     1,
		Description: "For test",
	}
	u := PathParameterFor(d.Name, d.Description)
	u.Default = d.Default
	if !reflect.DeepEqual(d, u) {
		t.Fatalf("Parameters is not equal: %+v, %+v", d, u)
	}

	ds := []Parameter{
		{Source: Path, Name: "test", Description: "For test"},
		{Source: Query, Name: "test", Description: "For test"},
		{Source: Header, Name: "test", Description: "For test"},
		{Source: Form, Name: "test", Description: "For test"},
		{Source: File, Name: "test", Description: "For test"},
		{Source: Body, Description: "For test"},
		{Source: Auto, Description: "For test"},
		{Source: Prefab, Name: "test", Description: "For test"},
	}
	us := []Parameter{
		PathParameterFor("test", "For test"),
		QueryParameterFor("test", "For test"),
		HeaderParameterFor("test", "For test"),
		FormParameterFor("test", "For test"),
		FileParameterFor("test", "For test"),
		BodyParameterFor("For test"),
		AutoParameterFor("For test"),
		PrefabParameterFor("test", "For test"),
	}
	for i, d := range ds {
		u := us[i]
		if !reflect.DeepEqual(d, u) {
			t.Fatalf("Parameters is not equal: %+v, %+v", d, u)
		}
	}
}

func TestResult(t *testing.T) {
	d := Result{
		Destination: Data,
		Description: "For test",
	}

	u := DataResultFor(d.Description)

	if !reflect.DeepEqual(d, u) {
		t.Fatalf("Results is not equal: %+v, %+v", d, u)
	}

	ds := []Result{
		{Destination: Meta, Description: "For test"},
		{Destination: Data, Description: "For test"},
		{Destination: Error},
	}
	us := []Result{
		MetaResultFor("For test"),
		DataResultFor("For test"),
		ErrorResult(),
	}
	for i, d := range ds {
		u := us[i]
		if !reflect.DeepEqual(d, u) {
			t.Fatalf("Results is not equal: %+v, %+v", d, u)
		}
	}
}

func TestOperator(t *testing.T) {
	kind := "test"
	test := func(o Operator) {
		if o.Kind() != kind {
			t.Fatalf("Operator kind is not correct: %s", o.Kind())
		}
		if o.In() != reflect.TypeOf(int(0)) {
			t.Fatalf("Operator in type is not correct: %s", o.Kind())
		}
		if o.Out() != reflect.TypeOf("") {
			t.Fatalf("Operator in type is not correct: %s", o.Kind())
		}
		result, err := o.Operate(context.Background(), "test", 1)
		if err != nil {
			t.Fatal(err)
		}
		if r, ok := result.(string); !ok {
			t.Fatalf("Operator result type is not correct: %s", reflect.TypeOf(result))
		} else if r != "test:1" {
			t.Fatalf("Operator result is not correct: %s", r)
		}
	}
	o := NewOperator(kind, reflect.TypeOf(int(0)), reflect.TypeOf(""),
		func(ctx context.Context, field string, object interface{}) (interface{}, error) {
			return fmt.Sprintf("%s:%d", field, object), nil
		})
	test(o)
	o = OperatorFunc(kind, func(ctx context.Context, field string, object int) (string, error) {
		return fmt.Sprintf("%s:%d", field, object), nil
	})
	test(o)
}
