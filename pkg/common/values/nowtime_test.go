package values

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Now() time.Time {
	t, _ := time.Parse(time.RFC3339, "2019-05-24T11:10:13+08:00")
	return t
}

func TestNowTimeValue(t *testing.T) {
	generator := &nowTimeString{
		nowTimeGetter: Now,
	}

	cases := []struct {
		Params   interface{}
		Expected string
	}{
		{
			NowTimeStringParam{},
			"1558667413",
		},
		{
			NowTimeStringParam{Format: time.RFC3339},
			"2019-05-24T11:10:13+08:00",
		},
		{
			"2019-05-24T11:10:13+08:00",
			"2019-05-24T11:10:13+08:00",
		},
		{
			"$(timenow:RFC3339)",
			"2019-05-24T11:10:13+08:00",
		},
		{
			"$(timenow:)",
			"1558667413",
		},
		{
			"aaaaa",
			"aaaaa",
		},
		{
			"$(timenow:20060102150405)",
			"20190524111013",
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expected, generator.Value(c.Params))
	}
}
