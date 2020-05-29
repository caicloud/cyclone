package values

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateValue(t *testing.T) {
	RandomString = &randomString{stringGenerator: stringGenerator}
	NowTimeString = &nowTimeString{
		nowTimeGetter: Now,
	}

	cases := []struct {
		Origin   string
		Expected string
	}{
		{
			"",
			"",
		},
		{
			"aaa",
			"aaa",
		},
		{
			"$(random:1)",
			"A",
		},
		{
			"$(random:2)",
			"BB",
		},
		{
			"$(random:x)",
			"$(random:x)",
		},
		{
			"$(timenow:RFC3339)",
			"2019-05-24T11:10:13+08:00",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expected, GenerateValue(c.Origin))
	}
}
