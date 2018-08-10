package github

import "testing"

func TestSplitStatusesURL(t *testing.T) {
	testCases := map[string]struct {
		url              string
		owner, repo, ref string
	}{
		"general github": {
			"https://api.github.com/repos/aaa/bbb/statuses/414330ad2bab35c8919ec2f3a2b20ac7cc103c28",
			"aaa", "bbb", "414330ad2bab35c8919ec2f3a2b20ac7cc103c28",
		},
	}

	for d, tc := range testCases {
		owner, repo, ref, err := splitStatusesURL(tc.url)
		if err != nil {
			t.Error("%s failed as error Expect error to be nil")
		}
		if owner != tc.owner || repo != tc.repo || ref != tc.ref {
			t.Errorf("%s failed as error : Expect result %s/%s/%s equals to %s/%s/%s",
				d, owner, repo, ref, tc.owner, tc.repo, tc.ref)
		}
	}
}
