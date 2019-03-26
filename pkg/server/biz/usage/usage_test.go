package usage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	cases := []struct {
		value    string
		expected float64
		error    bool
	}{
		{
			value:    "64B",
			expected: 64,
			error:    false,
		},
		{
			value:    "8.0K",
			expected: 8.0 * 1024,
			error:    false,
		},
		{
			value:    "32M",
			expected: 32 * 1024 * 1024,
			error:    false,
		},
		{
			value:    "0B",
			expected: 0,
			error:    false,
		},
		{
			value:    "0",
			expected: 0,
			error:    false,
		},
		{
			value:    "1.2G",
			expected: 1.2 * 1024 * 1024 * 1024,
			error:    false,
		},
		{
			value:    "2.0T",
			expected: 2.0 * 1024 * 1024 * 1024 * 1024,
			error:    false,
		},
		{
			value:    "64",
			expected: 0,
			error:    true,
		},
		{
			value:    "32k",
			expected: 0,
			error:    true,
		},
		{
			value:    "1.6Mi",
			expected: 0,
			error:    true,
		},
		{
			value:    "1.6GB",
			expected: 0,
			error:    true,
		},
		{
			value:    "K",
			expected: 0,
			error:    true,
		}, {
			value:    "1.0S",
			expected: 0,
			error:    true,
		},
	}

	for _, c := range cases {
		result, err := Parse(c.value)
		if c.error {
			assert.Error(t, err, c.value)
		} else {
			assert.Equal(t, c.expected, result, c.value)
		}
	}
}
