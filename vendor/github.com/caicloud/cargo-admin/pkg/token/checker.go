package token

import (
	"fmt"
	"strings"

	"github.com/caicloud/cargo-admin/pkg/utils/parser"

	"github.com/caicloud/nirvana/log"

	"github.com/docker/distribution/registry/auth/token"
)

var checkerMap = map[string]accessChecker{
	"repository": &repoChecker{},
	"registry":   &registryChecker{},
}

type accessChecker interface {
	check(perm *permManager, action *token.ResourceActions) error
}

type repoChecker struct {
}

type registryChecker struct {
}

func (c *registryChecker) check(perm *permManager, access *token.ResourceActions) error {
	// Only registry catalog access allowed
	if access.Name != "catalog" {
		return fmt.Errorf("unable to handle, type: %s, name: %s", access.Type, access.Name)
	}

	// Only admin user can access
	if !perm.IsRegistryAdmin(perm.RegistryInfo) && !perm.IsSystemAdmin(perm.GetUsername()) {
		//Set the actions to empty if the user is not admin
		access.Actions = []string{}
	}
	return nil
}

func (c *repoChecker) check(perm *permManager, access *token.ResourceActions) error {
	image, err := parser.ImageName(access.Name)
	if err != nil {
		return err
	}

	p, err := perm.ProjectPerm(image.Project)
	if err != nil {
		log.Errorf("get project permission for %s error: %v", image.Project, err)
		return err
	}

	access.Actions = permToActions(p)
	return nil
}

func CheckPerm(perm *permManager, accesses []*token.ResourceActions) error {
	var err error
	for _, a := range accesses {
		c, ok := checkerMap[a.Type]
		if !ok {
			a.Actions = []string{}
			log.Warningf("no access checker found for access type: %s, skip it, the access of resource '%s' will be set empty.", a.Type, a.Name)
			continue
		}
		err = c.check(perm, a)
		if err != nil {
			log.Errorf("permission check error: %v", err)
			return err
		}
	}
	return nil
}

func permToActions(p string) []string {
	res := []string{}
	if strings.Contains(p, "W") {
		res = append(res, "push")
	}
	if strings.Contains(p, "M") {
		res = append(res, "*")
	}
	if strings.Contains(p, "R") {
		res = append(res, "pull")
	}
	if strings.Contains(p, "*") {
		res = append(res, "*")
	}
	return res
}
