package resource

import (
	"context"
	"fmt"
	"time"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/log"

	"gopkg.in/mgo.v2"
)

var DefaultProtectedProjects = map[string]bool{
	"library":  true,
	"release":  true,
	"caicloud": true,
}

func CreatePublicProject(ctx context.Context, tenant, registry, pname, desc string) (*types.PublicProject, error) {
	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}

	// If project already exists in harbor, use it directly
	p, err := cli.GetProjectByName(pname)
	if err != nil {
		return nil, err
	}

	var pid int64
	if p == nil {
		pid, err = cli.CreateProject(pname, true)
		if err != nil {
			log.Errorf("create public project: %s in harbor error: %v, tenant: %s", pname, err, tenant)
			return nil, err
		}
	} else {
		pid = p.ProjectID
	}
	now := time.Now()

	err = models.Project.Save(&models.ProjectInfo{
		Name:           pname,
		Registry:       registry,
		Tenant:         tenant,
		ProjectId:      pid,
		Description:    desc,
		IsPublic:       true,
		IsProtected:    false,
		CreationTime:   now,
		LastUpdateTime: now,
	})
	if err != nil {
		log.Infof("create project %s failed due to mongo err: %v", pname, err)
		log.Infof("remove project %s from Harbor since creation in Cargo-Admin failed", pname)
		if cli.DeleteProject(pid) != nil {
			log.Errorf("delelete project %s from Harbor error: %v, this will result in orphen project in Harbor", pname, err)
		}
		return nil, ErrorUnknownInternal.Error(err)
	}
	log.Infof("create public project: %s success by tenant: %v, harbor projectId is: %d", pname, tenant, pid)

	return &types.PublicProject{
		Metadata: &types.ProjectMetadata{
			Name:           pname,
			CreationTime:   now,
			LastUpdateTime: now,
		},
		Spec: &types.ProjectSpec{
			IsPublic:            true,
			IsProtected:         false,
			Registry:            registry,
			Description:         desc,
			LastImageUpdateTime: now,
		},
		Status: &types.ProjectStatus{
			Synced:           true,
			RepositoryCount:  0,
			ReplicationCount: 0,
		},
	}, nil
}

func ListPublicProjects(ctx context.Context, tenant string, registry string, start, limit int) (int, []*types.Project, error) {
	total, pInfos, err := models.Project.FindOnePageOnlyPublic(registry, start, limit)
	if err != nil {
		return 0, nil, ErrorUnknownInternal.Error(err)
	}

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return 0, nil, err
	}

	hProjects, err := cli.AllProjects("", "true")
	if err != nil {
		log.Errorf("list all projects from registry: %s error: %s", registry, err)
		return 0, nil, err
	}
	hProjectMap := convertHarborProjectsToMap(hProjects)

	results := make([]*types.Project, 0, len(pInfos))
	for _, pInfo := range pInfos {
		hProject, ok := hProjectMap[pInfo.ProjectId]
		synced := true
		if !ok {
			log.Errorf("get project %d from Harbor failed: %v", pInfo.ProjectId, err)
			hProject = &harbor.HarborProject{}
			synced = false
		}

		cntSource, _ := models.Replication.CountAsSoure(tenant, registry, pInfo.Name)
		cntTarget, _ := models.Replication.CountAsTarget(tenant, registry, pInfo.Name)

		results = append(results, &types.Project{
			Metadata: &types.ProjectMetadata{
				Name:           pInfo.Name,
				CreationTime:   pInfo.CreationTime,
				LastUpdateTime: pInfo.LastUpdateTime,
			},
			Spec: &types.ProjectSpec{
				IsPublic:            pInfo.IsPublic,
				IsProtected:         pInfo.IsProtected,
				Registry:            pInfo.Registry,
				Description:         pInfo.Description,
				LastImageUpdateTime: hProject.UpdateTime,
			},
			Status: &types.ProjectStatus{
				Synced:           synced,
				RepositoryCount:  hProject.RepoCount,
				ReplicationCount: cntSource + cntTarget,
			},
		})
	}
	return total, results, nil
}

func UpdatePublicProject(ctx context.Context, tenant, registry, pname, des string) (*types.Project, error) {
	log.Infof("des: %v", des)
	pinfo, err := models.Project.FindByName(tenant, registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrorContentNotFound.Error(fmt.Sprintf("project: %s", pname))
		}
		return nil, ErrorUnknownInternal.Error(err)
	}

	err = models.Project.Update(tenant, registry, pname, des)
	if err != nil {
		log.Errorf("update project from mongo error: %v", err)
		return nil, ErrorUnknownInternal.Error(err)
	}

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}
	project, err := cli.GetProject(pinfo.ProjectId)
	if err != nil {
		log.Errorf("get project from harbor error: %v, projectId: %s, projectName: %s", err, pinfo.ProjectId, pname)
		return nil, err
	}

	err = models.Project.Update(tenant, registry, pname, des)
	if err != nil {
		return nil, err
	}

	cntSource, _ := models.Replication.CountAsSoure(tenant, registry, pinfo.Name)
	cntTarget, _ := models.Replication.CountAsTarget(tenant, registry, pinfo.Name)

	return &types.Project{
		Metadata: &types.ProjectMetadata{
			Name:           pname,
			CreationTime:   pinfo.CreationTime,
			LastUpdateTime: pinfo.LastUpdateTime,
		},
		Spec: &types.ProjectSpec{
			IsPublic:            pinfo.IsPublic,
			IsProtected:         pinfo.IsProtected,
			Registry:            pinfo.Registry,
			Description:         des,
			LastImageUpdateTime: project.UpdateTime,
		},
		Status: &types.ProjectStatus{
			Synced:           true,
			RepositoryCount:  project.RepoCount,
			ReplicationCount: cntSource + cntTarget,
		},
	}, nil
}

func DeletePublicProject(ctx context.Context, tenant, registry, pname string) error {
	pinfo, err := models.Project.FindByName(tenant, registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return ErrorContentNotFound.Error(fmt.Sprintf("project: %s", pname))
		}
		return ErrorUnknownInternal.Error(err)
	}
	if pinfo.IsProtected || DefaultProtectedProjects[pname] {
		return errors.Forbidden.Build(types.ProjectProtected, "can't delete protected project ${project}").Error(pname)
	}

	return deleteProject(pinfo, registry, tenant)
}
