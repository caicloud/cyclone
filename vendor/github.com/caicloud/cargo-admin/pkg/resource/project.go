package resource

import (
	"context"
	"fmt"
	"time"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/transaction"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/nirvana/errors"
	"gopkg.in/mgo.v2"
)

const (
	ReadWriteAccess = "rw"
	ReadAccess      = "r"
)

// projects are sorted by [isPublic, creationTime], public projects and newly created projects come first
func ListProjects(ctx context.Context, tenant string, registry string, includePublic bool, q string, start, limit int) (int, []*types.Project, error) {
	var total int
	var pInfos []*models.ProjectInfo
	var err error
	if q != "" {
		total, pInfos, err = models.Project.FindOnePageWithPrefix(tenant, registry, includePublic, q, start, limit)
		if err != nil {
			return 0, nil, ErrorUnknownInternal.Error(err)
		}
	} else {
		total, pInfos, err = models.Project.FindOnePage(tenant, registry, includePublic, start, limit)
		if err != nil {
			return 0, nil, ErrorUnknownInternal.Error(err)
		}
	}

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return 0, nil, err
	}

	hProjects, err := cli.AllProjects(q, "")
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

func convertHarborProjectsToMap(hprojects []*harbor.HarborProject) map[int64]*harbor.HarborProject {
	ret := make(map[int64]*harbor.HarborProject)
	for _, hproject := range hprojects {
		ret[hproject.ProjectID] = hproject
	}
	return ret
}

func GetProject(ctx context.Context, tenant, registry, pname string) (*types.Project, error) {
	pinfo, err := models.Project.FindByNameWithoutTenant(registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.NotFound.Build(types.ProjectNotExist, "${project} not exist").Error(pname)
		}
		return nil, ErrorUnknownInternal.Error(err)
	}

	if pinfo.Tenant != tenant && !pinfo.IsPublic {
		return nil, errors.NotFound.Build(types.ProjectNotExist, "${project} not exist").Error(pname)
	}

	log.Infof("%s", pname)

	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}
	project, err := cli.GetProject(pinfo.ProjectId)
	if err != nil {
		log.Errorf("get project from harbor error: %v, projectId: %s, projectName: %s", err, pinfo.ProjectId, pname)
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
			Description:         pinfo.Description,
			LastImageUpdateTime: project.UpdateTime,
		},
		Status: &types.ProjectStatus{
			Synced:           true,
			RepositoryCount:  project.RepoCount,
			ReplicationCount: cntSource + cntTarget,
		},
	}, nil
}

func CreateProject(ctx context.Context, tenant, registry, pname, desc string, isPublic bool) (*types.Project, error) {
	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	t := transaction.New()

	// Create project in registry
	var pid int64
	action := &transaction.Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			// If project already exists in harbor, use it directly
			p, err := cli.GetProjectByName(pname)
			if err != nil {
				return nil, err
			}
			if p != nil {
				return []interface{}{p.ProjectID}, nil
			}

			pid, err = cli.CreateProject(pname, isPublic)
			if err != nil {
				log.Errorf("create project: %s in harbor error: %v, tenant: %s", pname, err, tenant)
				return nil, err
			}
			return []interface{}{pid}, nil
		},
		Rollbacker: func(args ...interface{}) error {
			return cli.DeleteProject(args[0].(int64))
		},
	}
	t.Add(action)

	// Create project in Cargo-Admin
	action = &transaction.Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			err := models.Project.Save(&models.ProjectInfo{
				Name:           pname,
				Registry:       registry,
				Tenant:         tenant,
				ProjectId:      pid,
				Description:    desc,
				IsPublic:       false,
				IsProtected:    false,
				CreationTime:   now,
				LastUpdateTime: now,
			})
			return nil, err
		},
		Rollbacker: func(args ...interface{}) error {
			return models.Project.Delete(tenant, registry, pname)
		},
	}
	t.Add(action)

	err = t.Run()
	if err != nil {
		return nil, err
	}

	return &types.Project{
		Metadata: &types.ProjectMetadata{
			Name:           pname,
			CreationTime:   now,
			LastUpdateTime: now,
		},
		Spec: &types.ProjectSpec{
			IsPublic:            false,
			IsProtected:         false,
			Registry:            registry,
			Description:         desc,
			LastImageUpdateTime: now,
		},
		Status: &types.ProjectStatus{
			Synced:           true,
			RepositoryCount:  0,
			ReplicationCount: 0,
		}}, nil
}

func UpdateProject(ctx context.Context, tenant, registry, pname, desc string) (*types.Project, error) {
	log.Infof("description: %v", desc)
	pinfo, err := models.Project.FindByName(tenant, registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrorContentNotFound.Error(fmt.Sprintf("project: %s", pname))
		}
		return nil, err
	}

	t := transaction.New()

	// update project in Cargo-Admin
	action := &transaction.Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			err := models.Project.Update(tenant, registry, pname, desc)
			if err != nil {
				log.Errorf("update project in mongo error: %v", err)
			}
			return nil, err
		},
		Rollbacker: func(args ...interface{}) error {
			return models.Project.Update(tenant, registry, pname, pinfo.Description)
		},
	}
	t.Add(action)

	// get project from registry
	var project *harbor.HarborProject
	action = &transaction.Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			cli, err := harbor.ClientMgr.GetClient(registry)
			if err != nil {
				log.Errorf("get registry client error: %v", err)
				return nil, ErrorUnknownInternal.Error("not found")
			}
			project, err = cli.GetProject(pinfo.ProjectId)
			return nil, err
		},
		Rollbacker: func(args ...interface{}) error {
			return nil
		},
	}
	t.Add(action)

	err = t.Run()
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
			Description:         desc,
			LastImageUpdateTime: project.UpdateTime,
		},
		Status: &types.ProjectStatus{
			Synced:           true,
			RepositoryCount:  project.RepoCount,
			ReplicationCount: cntSource + cntTarget,
		},
	}, nil
}

// To delete a project, all repos and replications would also be deleted
func DeleteProject(ctx context.Context, tenant, registry, pname string) error {
	pInfo, err := models.Project.FindByName(tenant, registry, pname)
	if err != nil {
		if err == mgo.ErrNotFound {
			return errors.NotFound.Build(types.ProjectNotExist, "${project} not exist").Error(pname)
		}
		return ErrorUnknownInternal.Error(err)
	}
	if pInfo.IsProtected {
		return errors.Forbidden.Build(types.ProjectProtected, "can't delete protected project ${project}").Error(pname)
	}

	return deleteProject(pInfo, registry, tenant)
}

func deleteProject(pInfo *models.ProjectInfo, registry, tenant string) error {
	cli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}

	// If project doesn't exist in Harbor, delete from cargo-admin directly
	exist, err := cli.ProjectExist(pInfo.ProjectId)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}
	if !exist {
		log.Infof("project '%d' not exist in Harbor, remove directly", pInfo.ProjectId)
		err = models.Project.Delete(tenant, registry, pInfo.Name)
		if err != nil {
			log.Errorf("delete project from mongo error: %v", err)
			return ErrorUnknownInternal.Error(err)
		}
		return nil
	}

	// Delete all replications if any
	err = DeleteAllReplications(pInfo.Name)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}

	// Delete all repos under the project if any
	err = DeleteAllRepos(registry, pInfo.ProjectId, pInfo.Name)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}

	t := transaction.New()

	// delete project from cargo-admin
	action := &transaction.Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			err = models.Project.Delete(tenant, registry, pInfo.Name)
			if err != nil {
				log.Errorf("delete project from mongo error: %v", err)
				return nil, ErrorUnknownInternal.Error(err)
			}
			return nil, nil
		},
		Rollbacker: func(args ...interface{}) error {
			return models.Project.Save(pInfo)
		},
	}
	t.Add(action)

	// delete project from Harbor
	action = &transaction.Action{
		Runner: func(args ...interface{}) ([]interface{}, error) {
			err = cli.DeleteProject(pInfo.ProjectId)
			if err != nil {
				log.Errorf("Delete project %s in Harbor error: %v", pInfo.Name, err)
				return nil, ErrorUnknownInternal.Error(err)
			}
			return nil, nil
		},
		Rollbacker: func(args ...interface{}) error {
			return nil
		},
	}
	t.Add(action)

	return t.Run()
}
