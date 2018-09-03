package token

import (
	"reflect"
	"testing"

	"github.com/docker/distribution/registry/auth/token"
)

func TestGetActions(t *testing.T) {
	cases := []struct {
		scopes  []string
		actions []*token.ResourceActions
	}{
		{
			scopes: []string{"repository:library/golang:*", "repository:library/busybox:pull,push"},
			actions: []*token.ResourceActions{
				{
					Type:    "repository",
					Name:    "library/golang",
					Actions: []string{"*"},
				},
				{
					Type:    "repository",
					Name:    "library/busybox",
					Actions: []string{"pull", "push"},
				},
			},
		},
		{
			scopes: []string{"repository:library/golang:* repository:library/busybox:pull,push"},
			actions: []*token.ResourceActions{
				{
					Type:    "repository",
					Name:    "library/golang",
					Actions: []string{"*"},
				},
				{
					Type:    "repository",
					Name:    "library/busybox",
					Actions: []string{"pull", "push"},
				},
			},
		},
		{
			scopes: []string{"repository:library/golang", "repository:library/busybox"},
			actions: []*token.ResourceActions{
				{
					Type:    "repository",
					Name:    "library/golang",
					Actions: []string{},
				},
				{
					Type:    "repository",
					Name:    "library/busybox",
					Actions: []string{},
				},
			},
		},
	}

	for _, c := range cases {
		actions := GetActions(c.scopes)
		if !reflect.DeepEqual(actions, c.actions) {
			t.Errorf("Get actions error, expected %s, but got %s", c.actions, actions)
		}
	}
}
