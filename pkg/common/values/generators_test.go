package values

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRefValue(t *testing.T) {
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
			"$RANDOM:1",
			"A",
		},
		{
			"$RANDOM:2",
			"BB",
		},
		{
			"$RANDOM:x",
			"$RANDOM:x",
		},
		{
			"$TIMENOW:RFC3339",
			"2019-05-24T11:10:13+08:00",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expected, ParseRefValue(c.Origin))
	}
}
