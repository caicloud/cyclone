package pythonic

import (
	"bytes"
	"errors"
	"log"
	"reflect"
)

// https://golang.org/doc/faq#nil_error
// Under the covers, interfaces are implemented as two elements, a type and a value.
// The value, called the interface's dynamic value,
// is an arbitrary concrete value and the type is that of the value.
// For the int value 3, an interface value contains, schematically, (int, 3).
//
// Support for dynamic parameters
// only 0 or 1 arg is allowed, args should be array or slice
// if len(args) == 0, return zero value of Type,
// Type Kind | Go Zero value
// bool      | (bool, false)
// int       | (int, 0)
// float     | (float64, 0)
// string    | (string, "")
// map       | (map, nil)
// slice     | (slice, nil)
// interface | (nil, nil)
// Dict      | (Dict, nil)
// if len(args) == 1. return args[0]
// if len(args) > 1, panic it
func getDefault(t reflect.Type, args interface{}) *reflect.Value {
	value := reflect.ValueOf(args)
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		// pass
	default:
		log.Panic("args should be array or slice")
	}

	var def reflect.Value
	switch value.Len() {
	case 0:
		def = reflect.Zero(t)
	case 1:
		def = value.Index(0)
	default:
		log.Panicf("only 1 arg is allowed, receive too many arguments %d", value.Len())
	}

	return &def
}

// Dict is a Python like Dict type
type Dict map[interface{}]interface{}

// NewDict make a Dict
func NewDict() Dict {
	return make(Dict)
}

// DictReflect reflect a map into Dict
// v should be map type
func DictReflect(v interface{}) (Dict, error) {
	if reflect.TypeOf(v).Kind() != reflect.Map {
		return nil, errors.New("interface should be type of map")
	}
	dict := Dict{}
	value := reflect.ValueOf(v)
	keys := value.MapKeys()

	for _, k := range keys {
		_k := k.Interface()
		dict[_k] = value.MapIndex(k).Interface()
	}
	return dict, nil

}

// String returns a string value of Dcit
func (dict Dict) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("{")
	for k, v := range dict {
		buf.WriteString(spprint(k))
		buf.WriteString(": ")
		buf.WriteString(spprint(v))
		buf.WriteString(", ")
	}
	buf.WriteString("}")
	return buf.String()
}

// Keys returns Dict keys
func (dict Dict) Keys() []interface{} {
	keys := make([]interface{}, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	return keys
}

// HasKey checks if dict has given key
func (dict Dict) HasKey(key interface{}) bool {
	_, ok := dict[key]
	return ok
}

// SetDefault returns the value of key
// if the key is not exist, set the value and return it
func (dict Dict) SetDefault(key, def interface{}) interface{} {
	if value, ok := dict[key]; ok {
		return value
	}

	dict[key] = def
	return def

}

// Delete key and value from dict
func (dict Dict) Delete(key interface{}) {
	delete(dict, key)
}

// Update the dict with another dict
func (dict Dict) Update(other Dict) {
	for key, value := range other {
		dict[key] = value
	}
}

// Get try to get the value of key
func (dict Dict) Get(key interface{}) interface{} {
	return dict[key]

}

// Set key, value int Dict
func (dict Dict) Set(key, value interface{}) {
	dict[key] = value
}

// MustGet the value of key
// if key not exists, return default
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGet(key) == MustGet(key, nil)
// get the first item in args as default value
func (dict Dict) MustGet(key interface{}, args ...interface{}) interface{} {
	if len(args) > 1 {
		panic("only one arg allowed")
	}

	if v, ok := dict[key]; ok {
		return v
	}

	// var t interface{}
	// 空的interface的反射类型是nil, 暂时没想到拿到反射后的类型为interface的办法, 先用数组代替
	var t []int
	var def = getDefault(reflect.TypeOf(t), args)
	return def.Interface()

}

// MustGetBool must get the value of key, and the type of value must be bool
// if key not exists or type of value is not bool, return default
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGetBool(key) == MustGetBool(key, false)
// get the first item in args as default value
func (dict Dict) MustGetBool(key interface{}, args ...bool) bool {
	if len(args) > 1 {
		panic("only one arg allowed")
	}

	if v, ok := dict[key]; ok {
		if _v, _ok := v.(bool); _ok {
			return _v
		}
	}

	var def = getDefault(reflect.TypeOf(false), args)
	return def.Bool()
}

// MustGetInt must get the value of key, and the type of value must be int
// if key not exists or type of value is not int, return default
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGetInt(key) == MustGetInt(key, 0)
// get the first item in args as default value
func (dict Dict) MustGetInt(key interface{}, args ...int) int {
	if len(args) > 1 {
		panic("only one arg allowed")
	}

	if v, ok := dict[key]; ok {
		_v, err := Int64(v)
		if err == nil {
			return int(_v)
		}
	}

	var def = getDefault(reflect.TypeOf(0), args)
	return int(def.Int())
}

// MustGetInt64 must get the value of key, and the type of value must be int64
// if key not exists or type of value is not int64, return default
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGetInt64(key) == MustGetInt64(key, int64(0))
// get the first item in args as default value
func (dict Dict) MustGetInt64(key interface{}, args ...int64) int64 {
	if len(args) > 1 {
		panic("only one arg allowed")
	}

	if v, ok := dict[key]; ok {
		_v, err := Int64(v)
		if err == nil {
			return _v
		}
	}

	var def = getDefault(reflect.TypeOf(int64(0)), args)
	return def.Int()
}

// MustGetFloat64 must get the value of key, and the type of value must be float64
// if key not exists or type of value is not float64, return default value
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGetFloat64(key) == MustGetFloat64(key, float64(0))
// get the first item in args as default value
func (dict Dict) MustGetFloat64(key interface{}, args ...float64) float64 {
	if len(args) > 1 {
		panic("only one arg allowed")
	}

	if v, ok := dict[key]; ok {
		_v, err := Float64(v)
		if err == nil {
			return _v
		}
	}

	var def = getDefault(reflect.TypeOf(1.0), args)
	return def.Float()
}

// MustGetString must get the value of key, and the type of value must be string
// if key not exists or type of value is not string, return default value
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGetString(key) == MustGetString(key, "")
// get the first item in args as default value
func (dict Dict) MustGetString(key interface{}, args ...string) string {
	if len(args) > 1 {
		panic("only one arg allowed")
	}

	if v, ok := dict[key]; ok {
		if _v, _ok := v.(string); _ok {
			return _v
		}
	}

	var def = getDefault(reflect.TypeOf(""), args)
	return def.String()
}

// MustGetArray must get the value of key, and the type of value must be slice
// if key not exists or type of value is not slice, return default value
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGetArray(key) == MustGetArray(key, []interface{}(nil))
// get the first item in args as default value
func (dict Dict) MustGetArray(key interface{}, args ...[]interface{}) []interface{} {
	if len(args) > 1 {
		panic("only one arg allowed")
	}
	// https://github.com/golang/go/wiki/InterfaceSlice
	if v, ok := dict[key]; ok {
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
			rv := reflect.ValueOf(v)
			data := make([]interface{}, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				data[i] = rv.Index(i).Interface()
			}
			return data
		}
	}

	var t []interface{}
	var def = getDefault(reflect.TypeOf(t), args)
	return def.Interface().([]interface{})
}

// MustGetStringArray must get the value of key, and the type of value must be string slice
// if key not exists or type of value is not string slice, return default value
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGetStringArray(key) == MustGetStringArray(key, []string(nil))
// get the first item in args as default value
func (dict Dict) MustGetStringArray(key interface{}, args ...[]string) []string {
	if len(args) > 1 {
		panic("only one arg allowed")
	}

	if v, ok := dict[key]; ok {
		if _v, _ok := v.([]string); _ok {
			return _v
		}
	}

	var t []string
	var def = getDefault(reflect.TypeOf(t), args)
	return def.Interface().([]string)
}

// MustGetDict must get the value of key, and the type of value must be Dict
// if key not exists or type of value is not string Dict, return default value
//
// only 0 or 1 extra arg is allowed, otherwise will panic
// note: MustGetDict(key) == MustGetDict(key, Dict(nil))
// get the first item in args as default value
func (dict Dict) MustGetDict(key interface{}, args ...Dict) Dict {
	if len(args) > 1 {
		panic("only one arg allowed")
	}

	if v, ok := dict[key]; ok {
		dict, err := DictReflect(v)
		if err == nil {
			return dict
		}
	}

	var t Dict
	var def = getDefault(reflect.TypeOf(t), args)
	ret, _ := DictReflect(def.Interface())
	return ret
}

// Pop get the value of key then delete the key
func (dict Dict) Pop(key interface{}) interface{} {
	if v, ok := dict[key]; ok {
		delete(dict, key)
		return v
	}

	return nil

}

// func (dict Dict) MustPop(key interface{}, args ...interface{}) interface{} {
// 	var def interface{} = getDefault(reflect.Interface, args)

// 	if v, ok := dict[key]; ok {
// 		delete(dict, key)
// 		return v
// 	} else {
// 		return def
// 	}
// }
