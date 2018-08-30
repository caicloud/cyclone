package docker

import (
	"regexp"
	"strings"
)

var RepoPattern = regexp.MustCompile("^(?P<REPO>[\\w_\\-]+)(?::(?P<TAG>[\\w_.\\-]+))?$")

func ParseName(name string) (registry, project, repo, tag string, ok bool) {
	ok = false
	segs := strings.Split(name, "/")
	l := len(segs)
	if l <= 0 || l >= 4 {
		return
	}

	m := RepoPattern.FindStringSubmatch(segs[l-1])
	if len(m) <= 2 {
		return
	}
	repo = m[1]
	tag = m[2]
	if tag == "" {
		tag = "latest"
	}

	if l >= 2 {
		if l == 2 && (strings.ContainsAny(segs[l-2], ".:") || segs[l-2] == "localhost") {
			registry = segs[l-2]
			ok = true
			return
		}
		project = segs[l-2]
	}

	if l >= 3 {
		registry = segs[l-3]
	}

	ok = true
	return
}
