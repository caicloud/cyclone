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
			t.Errorf("%s failed as error %v Expect error to be nil", d, err)
		}

		if ref != tc.ref {
			t.Errorf("%s failed as error : Expect result %s equals to %s", d, ref, tc.ref)
		}
	}
}

func TestGetSVNChangedFiles(t *testing.T) {
	var m string

	m = `svnlook: warning: cannot set LC_CTYPE locale
svnlook: warning: environment variable LC_CTYPE is UTF-8
svnlook: warning: please check that your locale name is correct
U   cyclone/test.go
U   cyclone/README.md
`
	s := struct{}{}
	expect := map[string]struct{}{"cyclone/test.go": s, "cyclone/README.md": s}
	fs := getSVNChangedFiles(m)
	for _, f := range fs {
		if _, ok := expect[f]; !ok {
			t.Errorf("%v not exist in expect map:%+v", f, expect)
		}
	}
}
