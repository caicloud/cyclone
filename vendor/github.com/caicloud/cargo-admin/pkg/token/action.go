package token

import (
	"strings"

	"github.com/docker/distribution/registry/auth/token"

	"github.com/caicloud/nirvana/log"
)

// Get resource actions from scope
func GetActions(scopes []string) []*token.ResourceActions {
	scopes = flat(scopes)
	log.Infof("scopes: %+v", scopes)
	var res []*token.ResourceActions
	for _, s := range scopes {
		if s == "" {
			continue
		}
		items := strings.Split(s, ":")
		length := len(items)

		t := ""
		n := ""
		a := []string{}

		if length == 1 {
			t = items[0]
		} else if length == 2 {
			t = items[0]
			n = items[1]
		} else {
			t = items[0]
			n = strings.Join(items[1:length-1], ":")
			if len(items[length-1]) > 0 {
				a = strings.Split(items[length-1], ",")
			}
		}

		res = append(res, &token.ResourceActions{
			Type:    t,
			Name:    n,
			Actions: a,
		})
	}
	return res
}

func flat(scopes []string) []string {
	result := make([]string, 0)
	for _, s := range scopes {
		result = append(result, strings.Split(s, " ")...)
	}
	return result
}
