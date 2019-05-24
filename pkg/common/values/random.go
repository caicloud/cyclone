package values

import (
	"strconv"
	"strings"
)

type randomString struct {
	stringGenerator func(int) string
}

type RandomValueParam struct {
	Length int `json:"length"`
}

// Value generates random string value based on the input params.
func (r *randomString) Value(param interface{}) string {
	switch v := param.(type) {
	case int:
		return r.value(&RandomValueParam{Length: v})
	case int64:
		return r.value(&RandomValueParam{Length: int(v)})
	case string:
		return r.Parse(v)
	case RandomValueParam:
		return r.value(&v)
	case *RandomValueParam:
		return r.value(v)
	default:
		return ""
	}
}

func (r *randomString) value(param *RandomValueParam) string {
	if param.Length <= 0 {
		return ""
	}

	return r.stringGenerator(param.Length)
}

// Parse parses random values from a string. If the input string is a valid random value ref
// value: $RANDOM:<length>, for example: $RANDOM:5, then generate a random value accordingly,
// otherwise return the input string itself.
func (r *randomString) Parse(v string) string {
	trimed := strings.TrimSpace(v)
	if !strings.HasPrefix(trimed, "$RANDOM:") {
		return v
	}

	p := strings.TrimPrefix(trimed, "$RANDOM:")
	length, err := strconv.ParseInt(p, 10, 32)
	if err != nil {
		return v
	}

	return r.Value(length)
}
