package handler

import (
	"testing"
)

func TestSplitStatusesURL(t *testing.T) {
	testCases := map[string]struct {
		url string
		ref string
	}{
		"general github": {
			"https://api.github.com/repos/aaa/bbb/statuses/414330ad2bab35c8919ec2f3a2b20ac7cc103c28",
			"414330ad2bab35c8919ec2f3a2b20ac7cc103c28",
		},
	}
	for d, tc := range testCases {
		ref, err := extractCommitSha(tc.url)
		if err != nil {
			t.Error("%s failed as error Expect error to be nil")
		}

		if ref != tc.ref {
			t.Errorf("%s failed as error : Expect result %s equals to %s", d, ref, tc.ref)
		}
	}
}
