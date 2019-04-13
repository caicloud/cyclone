package config

import (
	"testing"
)

func TestValidateNotification(t *testing.T) {
	testCases := map[string]struct {
		endpoints []NotificationEndpoint
		expected  bool
	}{
		"different endpoint names": {
			endpoints: []NotificationEndpoint{
				{
					Name: "n1",
					URL:  "http://n1.cyclone.dev",
				},
				{
					Name: "n2",
					URL:  "http://n2.cyclone.dev",
				},
			},
			expected: true,
		},
		"same endpoint names": {
			endpoints: []NotificationEndpoint{
				{
					Name: "n1",
					URL:  "http://n1.cyclone.dev",
				},
				{
					Name: "n1",
					URL:  "http://n2.cyclone.dev",
				},
			},
			expected: false,
		},
	}

	for d, tc := range testCases {
		result := validateNotification(tc.endpoints)
		if result != tc.expected {
			t.Errorf("Test case %s failed: expected %t, but got %t", d, tc.expected, result)
		}
	}
}
