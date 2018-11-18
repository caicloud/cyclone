package workflowrun

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	cases := []struct {
		input string
		expected time.Duration
		err bool
	} {
		{
			input: "30min",
			expected: time.Minute * 30,
		},
		{
			input: "20m",
			expected: time.Minute * 20,
		},
		{
			input: "30s",
			expected: time.Second * 30,
		},
		{
			input: "1hour",
			expected: time.Hour,
		},
		{
			input: "1h30s",
			expected: time.Hour + time.Second * 30,
		},
		{
			input: "1h30m",
			expected: time.Hour + time.Minute * 30,
		},
		{
			input: "1h30m30s",
			expected: time.Hour + time.Minute * 30 + time.Second * 30,
		},
		{
			input: "1h:30m:30s",
			err: true,
		},
		{
			input: "1H30m",
			expected: time.Hour + time.Minute * 30,
		},
		{
			input: "1H30MIN",
			expected: time.Hour + time.Minute * 30,
		},
		{
			input: "1H30MIN----",
			err: true,
		},
		{
			input: "1millisecond",
			err: true,
		},
		{
			input: "",
			err: true,
		},
	}

	for _, c := range cases {
		out, err := ParseTime(c.input)
		if c.err {
			if err == nil {
				t.Errorf("%s expected to be invalid, but get %v", c.input, out)
			}
		} else {
			if out != c.expected {
				t.Errorf("%s expected to be %v, but got %v", c.input, c.expected, out)
			}
		}
	}
}
