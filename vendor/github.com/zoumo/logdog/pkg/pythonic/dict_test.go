package pythonic

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDict(t *testing.T) {
	// 1.
	dict := NewDict()
	dict["test"] = "test"
	dict[123] = 324325

	// 2.
	dict2 := Dict{"test": "test", 1: 1314}
	dict2[123] = 123

	assert.IsType(t, dict2.String(), "string")
}

func TestKey(t *testing.T) {
	dict := Dict{"test": "test", 1: 1314, 1.1: 111}
	assert.EqualValues(t, []interface{}{"test", 1, 1.1}, dict.Keys())
	assert.True(t, dict.HasKey("test"))
	assert.False(t, dict.HasKey("test2"))
	assert.False(t, dict.HasKey("1"))
	assert.True(t, dict.HasKey(1))
	assert.True(t, dict.HasKey(1.1))
	assert.True(t, dict.HasKey(1.10))

}

func TestSet(t *testing.T) {
	dict := NewDict()

	dict.SetDefault("test", "test")
	assert.Equal(t, "test", dict["test"])

	dict.SetDefault("test", "test2")
	assert.Equal(t, "test", dict["test"])

	dict.Set("test", "test2")
	assert.Equal(t, "test2", dict["test"])
}

func TestUpdate(t *testing.T) {
	dict := Dict{"test": "test"}
	dict2 := Dict{"test2": "test2"}
	dict.Update(dict2)
	assert.Equal(t, dict, Dict{"test": "test", "test2": "test2"})

}

func TestPop(t *testing.T) {
	dict := Dict{"test": "test"}
	str := dict.Pop("test")
	assert.Equal(t, "test", str)
	assert.Equal(t, Dict{}, dict)

	n := dict.Pop("test")
	assert.Nil(t, n)
}

func TestGet(t *testing.T) {
	dict := Dict{
		1:      1314,
		2:      1.2,
		3:      1314,
		"test": "test",
		"bool": true,
		"list": []string{"string", "list"},
		"map":  map[interface{}]string{"map": "map test", 1: "1"},
	}
	// test Get
	assert.Nil(t, dict.Get("null"))
	assert.Equal(t, dict.Get("test"), "test")

	// test MustGet simple type
	assert.Nil(t, dict.MustGet("null"))
	assert.Equal(t, dict.MustGet("test"), "test")
	assert.Equal(t, dict.MustGet("null", "null"), "null")
	assert.Panics(t, func() { dict.MustGet("test", 1, 2) })
	assert.True(t, dict.MustGetBool("bool"))
	assert.False(t, dict.MustGetBool("bool2"))

	// MustGetInt
	assert.Equal(t, dict.MustGetInt("test"), 0)
	assert.Equal(t, dict.MustGetInt("test", 1), 1)
	assert.Equal(t, dict.MustGetInt(1), 1314)
	assert.Equal(t, dict.MustGetInt(2), 1)
	assert.EqualValues(t, dict.MustGetInt64(2), 1)
	assert.EqualValues(t, dict.MustGetInt64("test"), 0)
	assert.EqualValues(t, dict.MustGetInt64(3), 1314)

	// MustGetFloat
	assert.Equal(t, dict.MustGetFloat64("test"), 0.0)
	assert.Equal(t, dict.MustGetFloat64(2), 1.2)

	// MustGetString
	assert.Equal(t, dict.MustGetString("test", "1"), "test")
	assert.Equal(t, dict.MustGetString("null"), "")

	// test MustGet Array
	assert.Nil(t, dict.MustGetArray("null"))
	assert.Equal(t, dict.MustGetArray("list"), []interface{}{"string", "list"})
	assert.Equal(t, dict.MustGetArray("null", []interface{}{"test", "array"}), []interface{}{"test", "array"})

	assert.Nil(t, dict.MustGetStringArray("null"))
	assert.Equal(t, dict.MustGetStringArray("list", []string{}), []string{"string", "list"})

	// test MustGet Dict
	assert.Equal(t, dict.MustGetDict("test"), Dict{})
	assert.Equal(t, dict.MustGetDict("map"), Dict{"map": "map test", 1: "1"})

}

func TestDelete(t *testing.T) {
	dict := NewDict()
	dict["test2"] = "3243243"
	delete(dict, "test2")
	assert.Nil(t, dict["test2"])
	dict["test2"] = "3243243"
	dict.Delete("test2")
	assert.Nil(t, dict["test2"])
}

func TestReflect(t *testing.T) {
	// 在go语言的底层,interface作为两个成员实现：一个type和一个value。
	// value被称为接口的动态值， 它是一个任意的具体值，而type则为value的类型。
	// 对于var i int = 3， 一个接口值示意性地表示为(int, 3)。
	// 一个空接口var v interface{} 可以表示为(nil, nil), 其中的type和value都未设置
	// 这时候 v == nil为true, 即接口值为nil
	// 而一旦v被赋予了一个有效的type, 不管这个type是什么, 接口都不为nil
	// 如:
	// var v interface{}
	// v == nil // true
	// var a []interface{}
	// a == nil // true
	// v = a // v ~= ([]interface{}, nil)  v != nil
	// v == nil // false
	// 只有对于interface的时候nil判断需要进行type和value的双重判断
	// 对于pointer, channel, func, map, slice这些零值为nil的类型来说, 只对value进行nil判断
	// var i *int // (*int, nil)
	// var c chan int // (chan int, nil)
	// var m map[int][int] // (map[int]int, nil)
	// var s []int // ([]int, nil)
	// var f func() // (func(), nil)
	// 这里提供了reflect前后的各个类型的zero值对比
	zero := func(v interface{}) *reflect.Value {
		var temp = reflect.Zero(reflect.TypeOf(v))
		return &temp
	}
	var v interface{} // (nil, nil)
	var ok bool
	var f float64
	assert.True(t, v == nil, "interface{} zero value: <nil>")
	assert.Equal(t, v, nil, "interface{} zero value: <nil>")
	assert.Equal(t, v, interface{}(nil), "interface{} zero value: <nil>")

	assert.Equal(t, zero("1").String(), "", "reflect string zero value: \"\"")
	assert.EqualValues(t, zero(1).Int(), 0, "reflect int zero value: 0")
	assert.Equal(t, zero(1).Kind(), reflect.Int, "reflect int zero value type: int")
	assert.EqualValues(t, zero(1.0).Float(), 0, "reflect float zero value: 0")
	assert.Equal(t, zero(1.0).Kind(), reflect.Float64, "reflect int zero value type: float64")
	assert.EqualValues(t, zero(f).Float(), 0, "reflect float64 zero value: 0")
	assert.Equal(t, zero(f).Kind(), reflect.Float64, "reflect int zero value type: float64")
	// 下面这个测试会失败, 不能对nil进行类型判断, 而所有的interface{}都有一个对应的实际类型
	// assert.True(t, zero(v).Kind(), "reflect interface{} zero value: nil")

	// test slice
	var a []interface{}
	assert.True(t, a == nil, "[]interface{} zero value is []interface{}(nil)")
	a, ok = zero(a).Interface().([]interface{})
	assert.True(t, ok)
	assert.True(t, a == nil, "reflect []interface{} zero value: []interface{}(nil)")
	assert.Equal(t, zero(a).Kind(), reflect.Slice, "reflect []interface{} zero value type: slice")

	// test map
	var m map[string]string
	assert.True(t, m == nil, "map zero value is map[string]string(nil)")
	m, ok = zero(m).Interface().(map[string]string)
	assert.True(t, ok)
	assert.True(t, m == nil, "reflect map zero value: map[string]string(nil)")
	assert.Equal(t, zero(m).Kind(), reflect.Map, "reflect map[string]string zero value type: map")

	// test Dict
	var dict Dict
	assert.True(t, dict == nil, "Dict zero value Dict(nil)")
	assert.Empty(t, dict, "Dict zero value is empty")
	dict, ok = zero(dict).Interface().(Dict)
	assert.True(t, ok)
	assert.True(t, dict == nil, "reflect Dict zero value: Dict(nil)")
	assert.Equal(t, zero(dict).Kind(), reflect.Map, "reflect Dict zero value type: map")

	// test nil
	assert.True(t, []interface{}(nil) == nil)
	assert.True(t, Dict(nil) == nil)
	assert.True(t, map[string]string(nil) == nil)
	assert.True(t, interface{}(nil) == nil)

	var x1 interface{}
	var x2 interface{}
	x2 = []interface{}(nil)
	assert.False(t, x1 == x2)

}
