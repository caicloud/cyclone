package docker

import (
	"reflect"
	"testing"
)

func TestLoadedImages(t *testing.T) {
	cases := []struct {
		output   string
		expected []string
	}{
		{
			output: `Loaded image: busy-box:v1.0.0\n
Loaded image: busy-box:latest\n`,
			expected: []string{"busy-box:v1.0.0", "busy-box:latest"},
		},
		{
			output:   `Loaded image: cargo.caicloudprivatetest.com/devops_test/busy_box:8.9-alpine`,
			expected: []string{"cargo.caicloudprivatetest.com/devops_test/busy_box:8.9-alpine"},
		},
		{
			output:   `Loaded image: cargo.caicloudprivatetest.com/caicloud/node:8.9-alpine`,
			expected: []string{"cargo.caicloudprivatetest.com/caicloud/node:8.9-alpine"},
		},
		{
			output:   `open /Users/foo/Desktop/busybox.tar: no such file or directory`,
			expected: make([]string, 0),
		},
	}

	for _, c := range cases {
		result := LoadedImages(c.output)
		if !reflect.DeepEqual(result, c.expected) {
			t.Errorf("Expected %v, but got %v", c.expected, result)
		}
	}
}
