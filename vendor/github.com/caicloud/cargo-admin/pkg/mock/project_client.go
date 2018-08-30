package mock

import (
	"strconv"

	"github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
)

type ProjectClient struct {
	nextId   int64
	projects []*harbor.HarborProject
}

func NewProjectClient() *ProjectClient {
	return &ProjectClient{
		nextId:   1,
		projects: make([]*harbor.HarborProject, 0),
	}
}

func (client *ProjectClient) Add(name string, isPublic bool, repoCnt int) {
	client.newProject(name, isPublic, repoCnt)
}

func (client *ProjectClient) CreateProject(name string, isPublic bool) (int64, error) {
	return client.newProject(name, isPublic, 0)
}

func (client *ProjectClient) newProject(name string, isPublic bool, repoCnt int) (int64, error) {
	client.projects = append(client.projects, &harbor.HarborProject{
		ProjectID: client.nextId,
		Name:      name,
		RepoCount: repoCnt,
		Metadata: map[string]string{
			"public": strconv.FormatBool(isPublic),
		},
	})
	client.nextId++
	return client.nextId - 1, nil
}

func (client *ProjectClient) GetProject(pid int64) (*harbor.HarborProject, error) {
	for _, p := range client.projects {
		if p.ProjectID == pid {
			return p, nil
		}
	}
	return nil, errors.ErrorUnknownInternal.Error("not found")
}

func (client *ProjectClient) GetProjectByName(name string) (*harbor.HarborProject, error) {
	for _, p := range client.projects {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, errors.ErrorUnknownInternal.Error("not found")
}

func (client *ProjectClient) ListProjects(page, pageSize int, name, public string) (int, []*harbor.HarborProject, error) {
	return 0, nil, nil
}

func (client *ProjectClient) AllProjects(name, public string) ([]*harbor.HarborProject, error) {
	return nil, nil
}

func (client *ProjectClient) DeleteProject(pid int64) error {
	for i, p := range client.projects {
		if p.ProjectID == pid {
			client.projects = append(client.projects[:i], client.projects[i+1:]...)
			return nil
		}
	}
	return errors.ErrorUnknownInternal.Error("not found")
}

func (client *ProjectClient) GetProjectDeleteable(pid int64) (*harbor.HarborProjectDeletableResp, error) {
	return nil, nil
}

func (client *ProjectClient) GetRepoCount(pid int64) (int, error) {
	for _, p := range client.projects {
		if p.ProjectID == pid {
			return p.RepoCount, nil
		}
	}
	return 0, errors.ErrorUnknownInternal.Error("not found")
}
