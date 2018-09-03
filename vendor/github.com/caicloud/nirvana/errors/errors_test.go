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

package errors

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestExpand(t *testing.T) {
	tab := []struct {
		format string
		args   []interface{}
		result string
		data   map[string]string
	}{
		{
			"name is too short",
			nil,
			"name is too short",
			nil,
		},
		{
			"name ${name} is too short",
			[]interface{}{"hzh"},
			"name hzh is too short",
			map[string]string{"name": "hzh"},
		},
		{
			"${name1} ${name2} ${name3}",
			[]interface{}{"one", "two", "three"},
			"one two three",
			map[string]string{"name1": "one", "name2": "two", "name3": "three"},
		},
		{
			"${name1 ${name2} ${name3}",
			[]interface{}{"one", "two"},
			"one two",
			map[string]string{"name1 ${name2": "one", "name3": "two"},
		},
		{
			"$name1} ${name2} name3}",
			[]interface{}{"one"},
			"$name1} one name3}",
			map[string]string{"name2": "one"},
		},
		{
			"i have ${count} ${items} in my ${what}",
			[]interface{}{"one", "dog", "house"},
			"i have one dog in my house",
			map[string]string{
				"count": "one",
				"items": "dog",
				"what":  "house",
			},
		},
		{
			"i have $",
			[]interface{}{"one", "dog", "house"},
			"i have $",
			nil,
		},
	}

	for _, v := range tab {
		s, m := expand(v.format, v.args...)
		if !reflect.DeepEqual(v.data, m) {
			t.Fatal(m, v.data)
		}
		if s != v.result {
			t.Fatal(s, v.result)
		}
	}
}

func TestExpandPanic1(t *testing.T) {
	defer func() {
		if x := recover(); x != nil {
			if fmt.Sprint(x) != "unexpected EOF while looking for matching }" {
				t.Fatal(x)
			}
		} else {
			t.Fatal("should panic")
		}
	}()
	expand("i have ${")
}

func TestExpandPanic2(t *testing.T) {
	defer func() {
		if x := recover(); x != nil {
			if fmt.Sprint(x) != "unexpected EOF while looking for matching }" {
				t.Fatal(x)
			}
		} else {
			t.Fatal("should panic")
		}
	}()
	expand("${name")
}

func TestExpandPanic3(t *testing.T) {
	defer func() {
		if x := recover(); x != nil {
			if fmt.Sprint(x) != "not enough args" {
				t.Fatal(x)
			}
		} else {
			t.Fatal("should panic")
		}
	}()
	expand("${name1 ${name2} ${name3")
}

func TestNewRaw(t *testing.T) {
	err := NewFactory(400, "japari:NotFriend", "${kind} is not in japari park").Error("anje").(*err)
	if err.Error() != "anje is not in japari park" ||
		err.Code() != 400 ||
		!reflect.DeepEqual(err.Message(), &message{
			Reason: "japari:NotFriend",
			Data: map[string]string{
				"kind": "anje",
			},
			Message: "anje is not in japari park",
		}) {
		t.Fatal(err)
	}
}

func TestDerived(t *testing.T) {
	friendNotFound := NotFound.Build("japari:NotFriend", "${kind} is not in japari park")
	foodNotFound := NotFound.Build("japari:NotFood", "${food} is not in japari park now")
	e1 := friendNotFound.Error("anje")
	e2 := foodNotFound.Error("charlotte")
	e3 := errors.New("anje is not in japari park")

	if !friendNotFound.Derived(e1) {
		t.Fatal(e1)
	}

	if friendNotFound.Derived(e2) {
		t.Fatal(e2)
	}

	if friendNotFound.Derived(e3) {
		t.Fatal(e3)
	}
}
