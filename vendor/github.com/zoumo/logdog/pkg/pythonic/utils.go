package pythonic

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func spprint(value interface{}) string {
	if str, ok := value.(string); ok {
		return fmt.Sprintf("%q", str)
	}

	return fmt.Sprint(value)

}

// Float64 coerces into a float64
func Float64(v interface{}) (float64, error) {
	switch v.(type) {
	case float32, float64:
		return reflect.ValueOf(v).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(v).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int coerces into an int
func Int(v interface{}) (int, error) {
	switch v.(type) {
	case float32, float64:
		return int(reflect.ValueOf(v).Float()), nil
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(v).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Int64 coerces into an int64
func Int64(v interface{}) (int64, error) {
	switch v.(type) {
	case float32, float64:
		return int64(reflect.ValueOf(v).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(v).Uint()), nil
	}
	return 0, errors.New("invalid value type")
}

// Uint64 coerces into an uint64
func Uint64(v interface{}) (uint64, error) {
	switch v.(type) {
	case json.Number:
		return strconv.ParseUint(v.(json.Number).String(), 10, 64)
	case float32, float64:
		return uint64(reflect.ValueOf(v).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(v).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint(), nil
	}
	return 0, errors.New("invalid value type")
}
