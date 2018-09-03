package mock

import (
	"sort"

	"github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"
)

type ProjectModel struct {
	projects []*models.ProjectInfo
}

func NewProjectModel() *ProjectModel {
	return &ProjectModel{
		projects: make([]*models.ProjectInfo, 0),
	}
}

func (p *ProjectModel) Len() int {
	return len(p.projects)
}

func (p *ProjectModel) Less(i, j int) bool {
	return p.projects[i].Name < p.projects[j].Name
}

func (p *ProjectModel) Swap(i, j int) {
	p.projects[i], p.projects[j] = p.projects[j], p.projects[i]
}

func (p *ProjectModel) EnsureIndexes() {
}

func (p *ProjectModel) Save(pinfo *models.ProjectInfo) error {
	p.projects = append(p.projects, pinfo)
	return nil
}

func (p *ProjectModel) GetTenantProjects(registry, tenant string) ([]*models.ProjectInfo, error) {
	return nil, nil
}

func (p *ProjectModel) GetGroupedProjects(registry string) ([]*models.ProjectGroup, error) {
	groups := make(map[string]*models.ProjectGroup)
	for _, project := range p.projects {
		g, ok := groups[project.Tenant]
		if !ok {
			groups[project.Tenant] = &models.ProjectGroup{
				Tenant:  project.Tenant,
				PIDs:    []int64{project.ProjectId},
				Publics: []bool{project.IsPublic},
			}
		} else {
			g.PIDs = append(g.PIDs, project.ProjectId)
			g.Publics = append(g.Publics, project.IsPublic)
		}
	}

	results := make([]*models.ProjectGroup, 0)
	for _, group := range groups {
		results = append(results, group)
	}
	return results, nil
}

func (p *ProjectModel) FindOnePage(tenant string, registry string, includePublic bool, start, limit int) (int, []*models.ProjectInfo, error) {
	return 0, nil, nil
}

func (p *ProjectModel) FindOnePageWithPrefix(tenant string, registry string, includePublic bool, prefix string, start, limit int) (int, []*models.ProjectInfo, error) {
	return 0, nil, nil
}

func (p *ProjectModel) FindOnePageOnlyPublic(registry string, start, limit int) (int, []*models.ProjectInfo, error) {
	return 0, nil, nil
}

func (p *ProjectModel) FindByName(tenant, registry, name string) (*models.ProjectInfo, error) {
	for _, project := range p.projects {
		if project.Tenant == tenant && project.Registry == registry && project.Name == name {
			return project, nil
		}
	}
	return &models.ProjectInfo{}, errors.ErrorUnknownInternal.Error("not found")
}

func (p *ProjectModel) FindByNameWithoutTenant(registry, name string) (*models.ProjectInfo, error) {
	for _, project := range p.projects {
		if project.Registry == registry && project.Name == name {
			return project, nil
		}
	}
	return nil, errors.ErrorUnknownInternal.Error("not found")
}

func (p *ProjectModel) FindAllByRegistry(registry string) ([]*models.ProjectInfo, error) {
	return nil, nil
}

func (p *ProjectModel) Delete(tenant, registry, name string) error {
	for idx, project := range p.projects {
		if project.Tenant == tenant && project.Registry == registry && project.Name == name {
			p.projects = append(p.projects[:idx], p.projects[idx+1:]...)
			break
		}
	}
	return nil
}

func (p *ProjectModel) DeleteWithoutTenant(registry, name string) error {
	for idx, project := range p.projects {
		if project.Registry == registry && project.Name == name {
			p.projects = append(p.projects[:idx], p.projects[idx+1:]...)
			break
		}
	}
	return nil
}

func (p *ProjectModel) DeleteAllByRegistry(registry string) error {
	for idx, project := range p.projects {
		if project.Registry == registry {
			p.projects = append(p.projects[:idx], p.projects[idx+1:]...)
			break
		}
	}
	return nil
}

func (p *ProjectModel) Update(tenant, registry, name, desc string) error {
	for _, project := range p.projects {
		if project.Tenant == tenant && project.Registry == registry && project.Name == name {
			project.Description = desc
			break
		}
	}
	return nil
}

func (p *ProjectModel) FindAllSortByName(tenant string, registry string, includePublic bool) ([]*models.ProjectInfo, error) {
	sort.Sort(p)
	return p.projects, nil
}

func (p *ProjectModel) IsExist(tenant, registry, name string) (bool, error) {
	for _, project := range p.projects {
		if project.Tenant == tenant && project.Registry == registry && project.Name == name {
			return true, nil
		}
	}
	return false, nil
}
