package values

import (
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
)

// Generator is used to generate specific type of string value, for example, fixed length random string, timestamp.
type Generator interface {
	// Value generates a string value according to params. How params is interpreted is determined by the generator implementation.
	Value(interface{}) string
	// Parse parses input string that represents a specific type value. For example, a given length random
	// string. If the input is not a valid given type ref value, the origin input value is returned.
	Parse(v string) string
}

var (
	// RandomString is used to generate a random string
	RandomString Generator
	// NowTimeString is used to generate current timestamp
	NowTimeString Generator
)

func init() {
	RandomString = &randomString{stringGenerator: rand.String}
	NowTimeString = &nowTimeString{nowTimeGetter: time.Now}
}

// GenerateValue passes through the input value to all the supported value generators to get the final value. For example,
// $(random:5) --> xafce
// $(timenow:RFC3339) --> 2019-05-24T11:10:13+08:00
func GenerateValue(value string) string {
	value = RandomString.Parse(value)
	value = NowTimeString.Parse(value)

	return value
}
