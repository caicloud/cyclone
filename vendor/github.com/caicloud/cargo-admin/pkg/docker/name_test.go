package docker

import "testing"

func TestParseName(t *testing.T) {
	cases := []struct {
		name string
		r    string
		p    string
		repo string
		tag  string
		ok   bool
	}{
		{"a.com/foo/bar:v1", "a.com", "foo", "bar", "v1", true},
		{"foo/bar:v1", "", "foo", "bar", "v1", true},
		{"bar:v1", "", "", "bar", "v1", true},
		{"bar", "", "", "bar", "latest", true},
		{"a.com/foo/bar", "a.com", "foo", "bar", "latest", true},
		{"acom/foo/bar:v1", "acom", "foo", "bar", "v1", true},
		{"a.com/foo/bar/baz:v1", "", "", "", "", false},
		{"a.com/foobar:v1", "a.com", "", "foobar", "v1", true},
		{"localhost/foobar", "localhost", "", "foobar", "latest", true},
		{"acom:5000/foobar:v1", "acom:5000", "", "foobar", "v1", true},
		{"a.com", "", "", "", "", false},
		{":v1.0", "", "", "", "", false},
		{"a.com:v2", "", "", "", "", false},
		{"bar.baz:v2", "", "", "", "", false},
		{"bar.baz:v2.0", "", "", "", "", false},
		{"bar-baz:v-2.0", "", "", "bar-baz", "v-2.0", true},
	}

	for _, c := range cases {
		r, p, repo, tag, ok := ParseName(c.name)
		if ok != c.ok || r != c.r || p != c.p || repo != c.repo || tag != c.tag {
			t.Errorf("Parse name %s error, expect {registry=%s, project=%s, repo=%s, tag=%s, ok=%t},"+
				" but got {registry=%s, project=%s, repo=%s, tag=%s, ok=%t}",
				c.name, c.r, c.p, c.repo, c.tag, c.ok, r, p, repo, tag, ok)
		}
	}
	ParseName("")
}
