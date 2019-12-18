package svn

import (
	"testing"
)

func TestGetSVNChangedFiles(t *testing.T) {
	m := `svnlook: warning: cannot set LC_CTYPE locale
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
