package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"

	"github.com/vmware/harbor/src/common/utils"
	"gopkg.in/mgo.v2"
)

func ListRepositories(ctx context.Context, tenant, registry, project, query, sort string, start, limit int) (int, []*types.Repository, error) {
	rInfo, err := models.Registry.FindByName(registry)
	if err != nil {
		return 0, nil, err
	}
	pInfo, err := models.Project.FindByNameWithoutTenant(registry, project)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		if err == mgo.ErrNotFound {
			return 0, nil, ErrorContentNotFound.Error(fmt.Sprintf("project: %v", project))
		}
		return 0, nil, ErrorUnknownInternal.Error(err)
	}
	if pInfo.Tenant != tenant && !pInfo.IsPublic {
		return 0, nil, ErrorContentNotFound.Error(fmt.Sprintf("project: %v", project))
	}

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return 0, nil, err
	}

	page, pageSize := convertPageParams(start, limit)

	// Repo names include project name, so if the given query is substring of project name, all repos will
	// be found, so in this case, we add project name as prefix.
	if strings.Index(pInfo.Name, query) >= 0 {
		query = pInfo.Name + "/" + query
	}
	total, repositories, err := cli.ListRepos(pInfo.ProjectId, query, sort, page, pageSize)
	if err != nil {
		log.Errorf("list repositories from harbor error: %v, projectId: %s, projectName: %s", err, project, pInfo.ProjectId)
		return 0, nil, err
	}

	repos := make([]*types.Repository, 0, len(repositories))
	for _, repository := range repositories {
		_, repoName := utils.ParseRepository(repository.Name)
		if err != nil {
			log.Errorf("list tags from harbor error: %v", err)
			return 0, nil, err
		}
		repos = append(repos, &types.Repository{
			Metadata: &types.RepositoryMetadata{
				Name:           repoName,
				CreationTime:   repository.CreationTime,
				LastUpdateTime: repository.UpdateTime,
			},
			Spec: &types.RepositorySpec{
				Project:     project,
				FullName:    fmt.Sprintf("%s/%s/%s", rInfo.Domain, project, repoName),
				Description: "",
			},
			Status: &types.RepositoryStatus{
				TagCount:  repository.TagsCount,
				PullCount: repository.PullCount,
			},
		})
	}
	return total, repos, err
}

func GetRepository(ctx context.Context, tenant string, registry string, pname string, repoName string) (*types.Repository, error) {
	rInfo, err := models.Registry.FindByName(registry)
	if err != nil {
		return nil, err
	}
	pinfo, err := models.Project.FindByNameWithoutTenant(registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrorContentNotFound.Error(fmt.Sprintf("project: %v", pname))
		}
		return nil, ErrorUnknownInternal.Error(err)
	}
	if pinfo.Tenant != tenant && !pinfo.IsPublic {
		return nil, ErrorContentNotFound.Error(fmt.Sprintf("project: %v", pname))
	}

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}
	repository, err := cli.GetRepository(pinfo.ProjectId, pinfo.Name, repoName)
	if err != nil {
		if err.Error() == ErrorContentNotFound.Error(fmt.Sprintf("repository: %s", repoName)).Error() {
			return &types.Repository{
				Metadata: &types.RepositoryMetadata{
					Name: repoName,
				},
				Spec: &types.RepositorySpec{
					Project:  pname,
					FullName: fmt.Sprintf("%s/%s/%s", rInfo.Domain, pname, repoName),
				},
			}, nil
		}
		log.Errorf("get repository: %s from cargo error: %v, projectId: %s, projectName: %s", repoName, err, pinfo.ProjectId, pname)
		return nil, err
	}
	tags, err := cli.ListTags(pname, repoName)
	if err != nil {
		log.Errorf("list tags from harbor error: %v", err)
		return nil, err
	}
	if len(tags) == 0 {
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("the repository: %s/%s has no tag", pname, repoName))
	}

	var des string
	b, err := models.Repository.IsExist(registry, pinfo.ProjectId, repoName)
	if err != nil {
		return nil, err
	}

	if b {
		repoInfo, err := models.Repository.FindByName(registry, pinfo.ProjectId, repoName)
		if err != nil {
			return nil, err
		}
		des = repoInfo.Description
	}

	return &types.Repository{
		Metadata: &types.RepositoryMetadata{
			Name:           repoName,
			CreationTime:   repository.CreationTime,
			LastUpdateTime: repository.UpdateTime,
		},
		Spec: &types.RepositorySpec{
			Project:     pname,
			FullName:    fmt.Sprintf("%s/%s/%s", rInfo.Domain, pname, repoName),
			Description: des,
		},
		Status: &types.RepositoryStatus{
			TagCount:  repository.TagsCount,
			PullCount: repository.PullCount,
		},
	}, nil
}

func UpdateRepository(ctx context.Context, tenant string, registry string, pname string, repoName string, des string) (*types.Repository, error) {
	rInfo, err := models.Registry.FindByName(registry)
	if err != nil {
		return nil, err
	}
	pinfo, err := models.Project.FindByNameWithoutTenant(registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrorContentNotFound.Error(fmt.Sprintf("project: %v", pname))
		}
		return nil, ErrorUnknownInternal.Error(err)
	}

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}
	repository, err := cli.GetRepository(pinfo.ProjectId, pinfo.Name, repoName)
	if err != nil {
		log.Errorf("get repository: %s from cargo error: %v, projectId: %s, projectName: %s", repoName, err, pname, pinfo.ProjectId)
		return nil, err
	}

	err = models.Repository.Upsert(registry, pinfo.ProjectId, repoName, des)
	if err != nil {
		log.Errorf("update repository: %s from cargo error: %v, projectId: %s, projectName: %s", repoName, err, pname, pinfo.ProjectId)
		return nil, err
	}

	return &types.Repository{
		Metadata: &types.RepositoryMetadata{
			Name:           repoName,
			CreationTime:   repository.CreationTime,
			LastUpdateTime: repository.UpdateTime,
		},
		Spec: &types.RepositorySpec{
			Project:     pname,
			FullName:    fmt.Sprintf("%s/%s/%s", rInfo.Domain, pname, repoName),
			Description: des,
		},
		Status: &types.RepositoryStatus{
			TagCount:  repository.TagsCount,
			PullCount: repository.PullCount,
		},
	}, nil
}

func DeleteRepository(ctx context.Context, tenant string, registry string, pname string, repoName string) error {
	_, err := models.Registry.FindByName(registry)
	if err != nil {
		return err
	}
	pinfo, err := models.Project.FindByNameWithoutTenant(registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return ErrorContentNotFound.Error(fmt.Sprintf("project: %v", pname))
		}
		return ErrorUnknownInternal.Error(err)
	}
	if pinfo.IsPublic && pinfo.Tenant != tenant {
		return ErrorDeleteFailed.Error(pname, fmt.Sprintf("project %v is public project, the repos in public project can only be deleted by system-admin tenant", pname))
	}

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return err
	}
	err = cli.DeleteRepository(pname, repoName)
	if err != nil {
		log.Errorf("delete repository:%s from cargo error: %v, projectId: %s, projectName: %s", repoName, err, pname, pinfo.ProjectId)
		return err
	}

	return nil
}

func DeleteAllRepos(registry string, pid int64, pname string) error {
	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return err
	}
	repos, err := cli.ListAllRepositories(pid)
	if err != nil {
		return err
	}

	total := len(repos)
	deleted := 0
	for _, repo := range repos {
		project, repository := utils.ParseRepository(repo.Name)
		err := cli.DeleteRepository(project, repository)
		if err != nil {
			log.Errorf("delete repository: %s/%s error: %v", err, project, repository)
		}
		deleted++
	}

	if total > 0 {
		log.Infof("%d out of %d repo from project [%s] deleted", deleted, total, pname)
	}
	if deleted < total {
		msg := fmt.Sprintf("delete all repos for project %s error, only %d out of %d deleted", pname, deleted, total)
		log.Errorf(msg)
		return ErrorUnknownInternal.Error(msg)
	}

	return nil
}
