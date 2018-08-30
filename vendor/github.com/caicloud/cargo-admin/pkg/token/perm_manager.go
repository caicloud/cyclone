package token

import (
	"github.com/caicloud/cargo-admin/pkg/cauth"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/resource"

	"github.com/caicloud/nirvana/log"

	"gopkg.in/mgo.v2"
)

var RoleTypePerm = map[string]string{
	cauth.RoleTypeOwner: "*",
	cauth.RoleTypeUser:  "RW",
	cauth.RoleTypeGuest: "R",
}

type permManager struct {
	*basicAuth
	*models.RegistryInfo
	*cauth.CauthClient
}

func (m *permManager) ProjectPerm(project string) (string, error) {
	pInfo, err := m.getProject(project)
	if err != nil || pInfo == nil {
		return "", err
	}

	if m.IsRegistryAdmin(m.RegistryInfo) {
		return m.asAdmin(pInfo)
	}

	if m.GetUsername() == "" {
		return m.asAnonymous(pInfo)
	}

	if m.IsCycloneUser() {
		return m.asCycloneUser(pInfo)
	}

	userInfo, err := m.CauthClient.GetUser(m.GetUsername())
	if err != nil {
		log.Errorf("get user information for '%s' error: %v", m.GetUsername(), err)
		return m.asAnonymous(pInfo)
	}

	if userInfo.Invalid {
		log.Infof("user '%s' is inactive", m.GetUsername())
		return m.asAnonymous(pInfo)
	}

	if isSystemAdmin(userInfo) {
		return m.asAdmin(pInfo)
	}

	return m.asUser(pInfo)
}

// For admin user, all permission granted
func (m *permManager) asAdmin(pInfo *models.ProjectInfo) (string, error) {
	return "RWM", nil
}

// For anonymous access, only READ on public project is allowed
func (m *permManager) asAnonymous(pInfo *models.ProjectInfo) (string, error) {
	if pInfo.IsPublic {
		return "R", nil
	}
	log.Infof("project: '%s' is not public, set empty permission for anonymous user", pInfo.Name)
	return "", nil
}

// Cyclone will use tenant id to register an account for docker access. And the generated account name will
// have "__cyclone__" as prefix. For token request from Cyclone, we check existence of the tenant and whether
// the project belongs to the tenant
func (m *permManager) asCycloneUser(pInfo *models.ProjectInfo) (string, error) {
	if !resource.IsMultiTenantEnabled() {
		return "RW", nil
	}

	if pInfo.Tenant != m.GetCycloneUserName() {
		if pInfo.IsPublic {
			return "R", nil
		}

		log.Infof("access to project %s as tenant %s denied, project belongs to tenant %s", pInfo.Name, m.GetCycloneUserName(), pInfo.Tenant)
		return "", nil
	}

	return "RW", nil
}

func (m *permManager) asUser(pInfo *models.ProjectInfo) (string, error) {
	if !resource.IsMultiTenantEnabled() {
		return "RW", nil
	}

	// Normal user can only pull from public project, no push allowed
	if pInfo.IsPublic {
		return "R", nil
	}

	tenants, err := m.ListTenants(m.GetUsername())
	if err != nil {
		log.Errorf("get tenants error: %s", err)
		return "", err
	}

	if len(tenants.Items) == 0 {
		log.Infof("user %s does belongs to no tenants in cauth, set empty permission", m.GetUsername())
		return "", nil
	}

	for _, tenant := range tenants.Items {
		if tenant.ID == pInfo.Tenant {
			if m.IsTenantAdmin(tenant, m.GetUsername()) {
				return "*", nil
			}
			return m.checkAuthZ(pInfo.Tenant, pInfo.Name, m.GetUsername())
		}
	}

	log.Infof("project: %s is not public project, all tenants including user: %s don't own the project", pInfo.Name, m.GetUsername())
	return "", nil
}

// Get project information. It checks existence of project in Harbor and Cargo-Admin, if both
// exist, return the project information. Otherwise, nil is returned.
// One exception is registry admin, if the user is registry admin, only check whether project
// exist in Harbor.
func (m *permManager) getProject(project string) (*models.ProjectInfo, error) {
	cli, err := harbor.ClientMgr.GetClient(m.RegistryInfo.Name)
	if err != nil {
		return nil, err
	}

	exist, err := cli.ProjectExistByName(project)
	if err != nil {
		log.Errorf("check existence of project '%s' in Harbor error: %v", project, err)
		return nil, err
	}

	if !exist {
		log.Warningf("project '%s' not exist in harbor", project)
		return nil, nil
	}

	// For registry admin, no need to check project existence in Cargo-Admin, and as the
	// returned project information would not be used, we can return a default one.
	if m.IsRegistryAdmin(m.RegistryInfo) {
		return &models.ProjectInfo{}, nil
	}

	pInfo, err := models.Project.FindByNameWithoutTenant(m.RegistryInfo.Name, project)
	if err != nil {
		if err == mgo.ErrNotFound {
			log.Infof("project '%s' not exist in cargo-admin", project)
			return nil, nil
		}

		log.Errorf("check existence of project '%s' in Cargo-Admin error: %v", project, err)
		return nil, err
	}

	return pInfo, nil
}

func isSystemAdmin(userInfo *cauth.User) bool {
	return userInfo.Role == "owner" && !userInfo.Invalid
}

func (m *permManager) IsSystemAdmin(user string) bool {
	userInfo, err := m.CauthClient.GetUser(user)
	if err != nil {
		log.Errorf("get user information for '%s' error: %v", user, err)
		return false
	}
	return isSystemAdmin(userInfo)
}

func (m *permManager) IsTenantAdmin(tenant cauth.Tenant, user string) bool {
	for _, member := range tenant.Members {
		if member.Name == user {
			return member.Role == "owner"
		}
	}
	return false
}

func (m *permManager) checkAuthZ(tenant, project, username string) (string, error) {
	teams, err := m.CauthClient.ListTeams(username)
	if err != nil {
		return "", nil
	}

	subjects := append(make([]cauth.Subject, 0), cauth.Subject{cauth.TenantType, tenant})
	for _, t := range teams.Items {
		if t.Tenant == tenant {
			subjects = append(subjects, cauth.Subject{cauth.TeamType, t.ID})
		}
	}

	perm := ""
	for _, s := range subjects {
		roles, err := m.CauthClient.ListRoles(s.SubType, s.SubID)
		if err != nil {
			continue
		}

		for _, r := range roles.Items {
			p, ok := RoleTypePerm[r.Type]
			if !ok {
				log.Warningf("Unknown roleType: %s", r.Type)
				continue
			}

			if r.Group == cauth.RoleGroupAllProject {
				if r.Resource.Data[cauth.DataKeyRegistry].Value == m.RegistryInfo.Name &&
					r.Resource.Data[cauth.DataKeyTenant].Value == tenant {
					log.Infof("grant '%s' to project '%s' from role '%s-%s'", p, project, r.Group, r.Type)
					perm = perm + p
				}
			} else if r.Group == cauth.RoleGroupAllWorkspace && r.Type == cauth.RoleTypeOwner {
				if r.Resource.Data[cauth.DataKeyTenant].Value == tenant {
					log.Infof("grant '%s' to project '%s' from role '%s-%s'", p, project, r.Group, r.Type)
					perm = perm + p
				}
			} else if r.Group == cauth.RoleGroupOneProject {
				if r.Resource.Data[cauth.DataKeyRegistry].Value == m.RegistryInfo.Name &&
					r.Resource.Data[cauth.DataKeyProject].Value == project &&
					r.Resource.Data[cauth.DataKeyTenant].Value == tenant {
					log.Infof("grant '%s' to project '%s' from role '%s-%s'", p, project, r.Group, r.Type)
					perm = perm + p
				}
			}
		}
	}

	return perm, nil
}
