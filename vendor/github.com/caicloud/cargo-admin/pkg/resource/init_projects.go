package resource

import (
	"time"

	"github.com/caicloud/cargo-admin/pkg/env"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"

	"github.com/davecgh/go-spew/spew"
)

var DefaultPublicProjectsMap map[string]*env.DefaultPublicProject

type ConflictStrategy string

func InitProject(projects []*env.DefaultPublicProject) error {
	DefaultPublicProjectsMap = make(map[string]*env.DefaultPublicProject)
	if len(projects) == 0 {
		log.Info("no default project to add")
		return nil
	}
	for _, p := range projects {
		log.Infof("add default project: %s", p.Name)
		if _, ok := DefaultPublicProjectsMap[p.Name]; !ok {
			DefaultPublicProjectsMap[p.Name] = p
		}
		err := EnsureDefaultPublicProject(p)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func EnsureDefaultPublicProject(p *env.DefaultPublicProject) error {
	log.Infof("ensure default public project '%s' exists in registry '%s'", p.Name, p.Harbor)
	cli, err := harbor.ClientMgr.GetClient(p.Harbor)
	if err != nil {
		return err
	}
	hProjects, err := cli.AllProjects(p.Name, "true")
	if err != nil {
		return err
	}

	hExist := false
	for _, hp := range hProjects {
		if p.Name == hp.Name {
			hExist = true
			err := upsertProjectInfo(p, hp)
			if err != nil {
				log.Errorf("upsertProjectInfo error: %v, default public project: %s, harbor project: %s",
					err, spew.Sdump(p), spew.Sdump(p))
				return err
			}
			log.Infof("add default public project: %v into cargo-admin", p.Name)
			break
		}
	}
	if !hExist {
		err := createDefaulPublicProject(cli, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func createDefaulPublicProject(cli *harbor.Client, p *env.DefaultPublicProject) error {
	log.Infof("create public project '%s' in harbor", p.Name)
	pid, err := cli.CreateProject(p.Name, true)
	if err != nil {
		log.Errorf("create project: %s error: %v", p.Name, err)
		return err
	}
	log.Infof("project '%s' created successfully", p.Name)

	hProject, err := cli.GetProject(pid)
	if err != nil {
		log.Errorf("get project: %s from harbor: %d error: %v", pid, p.Harbor, err)
		return err
	}
	err = upsertProjectInfo(p, hProject)
	if err != nil {
		log.Errorf("upsertProjectInfo error: %v, default public project: %s, harbor project: %s",
			err, spew.Sdump(p), spew.Sdump(hProject))
		return err
	}

	return nil
}

func upsertProjectInfo(p *env.DefaultPublicProject, project *harbor.HarborProject) error {
	pInfo := getDefaultPublicProjectInfo(p, project)
	b, err := models.Project.IsExist(env.SystemTenant, p.Harbor, p.Name)
	if err != nil {
		return err
	}
	if !b {
		log.Infof("default public project: %s has not existed in cargo-admin", p.Name)
		err = models.Project.Save(pInfo)
		if err != nil {
			log.Errorf("mongo error: %v", err)
			return err
		}
		log.Infof("save default public project: %s successfully", p.Name)
		return nil
	}

	switch p.IfExists {
	case env.ForceStrategy:
		err := models.Project.DeleteWithoutTenant(p.Harbor, p.Name)
		if err != nil {
			log.Errorf("mongo error: %v", err)
			return err
		}
		err = models.Project.Save(pInfo)
		if err != nil {
			log.Errorf("mongo error: %v", err)
			return err
		}
	case env.IgnoreStrategy:
		old, err := models.Project.FindByNameWithoutTenant(p.Harbor, p.Name)
		if err != nil {
			return err
		}
		log.Infof("project %s has existed in cargo-admin, ignore...", p.Name)
		log.Infof("project %s details: %s", p.Name, spew.Sdump(old))
	}
	return nil
}

func getDefaultPublicProjectInfo(p *env.DefaultPublicProject, project *harbor.HarborProject) *models.ProjectInfo {
	return &models.ProjectInfo{
		Name:           p.Name,
		Registry:       p.Harbor,
		Tenant:         env.SystemTenant,
		ProjectId:      project.ProjectID,
		Description:    "",
		IsPublic:       true,
		IsProtected:    true,
		CreationTime:   project.CreationTime,
		LastUpdateTime: time.Now(),
	}
}
