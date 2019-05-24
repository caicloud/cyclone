package values

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func stringGenerator(length int) string {
	if length%2 != 0 {
		return strings.Repeat("A", length)
	}

	return strings.Repeat("B", length)
}

func TestRandomValue(t *testing.T) {
	generator := &randomString{stringGenerator: stringGenerator}

	cases := []struct {
		Params   interface{}
		Expected string
	}{
		{
			5,
			"AAAAA",
		},
		{
			RandomValueParam{Length: 4},
			"BBBB",
		},
		{
			&RandomValueParam{Length: 3},
			"AAA",
		},
		{
			1.5,
			"",
		},
		{
			"value",
			"value",
		},
		{
			"$RANDOM:7",
			"AAAAAAA",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expected, generator.Value(c.Params))
	}
}
