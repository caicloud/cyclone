package regex

import (
	"testing"
)

func TestGetGitlabMRID(t *testing.T) {
	testCases := map[string]struct {
		ref    string
		strict bool
		id     int
		flag   bool
	}{
		"general gitlab mr": {
			"refs/merge-requests/12/head",
			true,
			12,
			true,
		},
		"invalid gitlab mr": {
			"refs/merge-requests/a/head",
			true,
			0,
			false,
		},
		"not end with head strict": {
			"refs/merge-requests/1/head:master",
			true,
			0,
			false,
		},
		"not end with head not strict": {
			"refs/merge-requests/1/head:master",
			false,
			1,
			true,
		},
	}

	for d, tc := range testCases {
		id, flag := GetGitlabMRID(tc.ref, tc.strict)
		if flag != tc.flag {
			t.Errorf("%s failed as error Expect flag %v to be %v", d, flag, tc.flag)
		}
		if id != tc.id {
			t.Errorf("%s failed as error : Expect result %d equals to %d", d, id, tc.id)
		}
	}
}

func TestGetGetGithubPRID(t *testing.T) {
	testCases := map[string]struct {
		ref  string
		id   int
		flag bool
	}{
		"general github pr": {
			"refs/pull/12/merge",
			12,
			true,
		},
		"invalid github pr": {
			"refs/pull/a/merge",
			0,
			false,
		},
	}

	for d, tc := range testCases {
		id, flag := GetGithubPRID(tc.ref)
		if flag != tc.flag {
			t.Errorf("%s failed as error Expect flag %v to be %v", d, flag, tc.flag)
		}
		if id != tc.id {
			t.Errorf("%s failed as error : Expect result %d equals to %d", d, id, tc.id)
		}
	}
}
