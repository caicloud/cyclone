/*
Copyright 2017 caicloud authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manager

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/zoumo/logdog"
	"gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
	"github.com/caicloud/cyclone/pkg/store"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	slug "github.com/caicloud/cyclone/pkg/util/slugify"
)

// ProjectManager represents the interface to manage project.
type ProjectManager interface {
	CreateProject(project *api.Project) (*api.Project, error)
	GetProject(projectName string) (*api.Project, error)
	GetProjectByID(id string) (*api.Project, error)
	ListProjects(queryParams api.QueryParams) ([]api.Project, int, error)
	UpdateProject(projectName string, newProject *api.Project) (*api.Project, error)
	DeleteProject(projectName string) error
	ListRepos(projectName string) ([]api.Repository, error)
	ListBranches(projectName string, repo string) ([]string, error)
	ListTags(projectName string, repo string) ([]string, error)
	ListDockerfiles(projectName string, repo string) ([]string, error)
	GetTemplateType(projectName string, repo string) (string, error)
	GetStatistics(projectName string, start, end time.Time) (*api.PipelineStatusStats, error)
}

// projectManager represents the manager for project.
type projectManager struct {
	dataStore       *store.DataStore
	pipelineManager PipelineManager
}

// NewProjectManager creates a project manager.
func NewProjectManager(dataStore *store.DataStore, pipelineManager PipelineManager) (ProjectManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new project manager as data store is nil.")
	}

	if pipelineManager == nil {
		return nil, fmt.Errorf("Fail to new project manager as pipeline manager is nil.")
	}

	return &projectManager{dataStore, pipelineManager}, nil
}

// CreateProject creates a project.
func (m *projectManager) CreateProject(project *api.Project) (*api.Project, error) {
	if project.Name == "" && project.Alias == "" {
		return nil, httperror.ErrorValidationFailed.Error("project name and alias", "can not neither be empty")
	}

	nameEmpty := false
	if project.Name == "" && project.Alias != "" {
		project.Name = slug.Slugify(project.Alias, false, -1)
		nameEmpty = true
	}

	if p, err := m.GetProject(project.Name); err == nil {
		log.Errorf("name %s conflict, project alias:%s, exist project alias:%s",
			project.Name, project.Alias, p.Alias)
		if nameEmpty {
			project.Name = slug.Slugify(project.Name, true, -1)
		} else {
			return nil, httperror.ErrorAlreadyExist.Error(project.Name)
		}

	}

	if err := scm.GenerateSCMToken(project.SCM); err != nil {
		return nil, err
	}

	return m.dataStore.CreateProject(project)
}

// GetProject gets the project by name.
func (m *projectManager) GetProject(projectName string) (*api.Project, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Error(projectName)
		}

		return nil, err
	}

	return project, nil
}

// GetProjectByID gets the project by id.
func (m *projectManager) GetProjectByID(id string) (*api.Project, error) {
	project, err := m.dataStore.FindProjectByID(id)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Error(id)
		}

		return nil, err
	}

	return project, nil
}

// ListProjects lists all projects of one owner.
func (m *projectManager) ListProjects(queryParams api.QueryParams) ([]api.Project, int, error) {
	return m.dataStore.GetProjects(queryParams)
}

// UpdateProject updates the project by name.
func (m *projectManager) UpdateProject(projectName string, newProject *api.Project) (*api.Project, error) {
	project, err := m.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	if err := scm.GenerateSCMToken(newProject.SCM); err != nil {
		return nil, err
	}
	project.SCM = newProject.SCM

	// Update the properties of the project.
	// TODO (robin) Whether need a method for this merge?
	project.Alias = newProject.Alias
	project.Description = newProject.Description

	if len(newProject.Owner) > 0 {
		project.Owner = newProject.Owner
	}

	project.Worker = newProject.Worker
	project.Registry = newProject.Registry

	if err = m.dataStore.UpdateProject(project); err != nil {
		return nil, err
	}

	return project, nil
}

// DeleteProject deletes the project by name.
func (m *projectManager) DeleteProject(projectName string) error {
	project, err := m.GetProject(projectName)
	if err != nil {
		return err
	}

	//delete project logs folder.
	m.DeleteProjectLogs(project.ID)

	// Delete the pipelines in this project.
	if err = m.pipelineManager.ClearPipelinesOfProject(projectName); err != nil {
		log.Errorf("Fail to delete all pipelines in the project %s as %s", projectName, err.Error())
		return err
	}

	if err = m.dataStore.DeleteProjectByID(project.ID); err != nil {
		log.Errorf("Fail to delete the project %s as %s", projectName, err.Error())
		return err
	}

	return nil
}

// ListRepos lists the SCM repos authorized for the project.
func (m *projectManager) ListRepos(projectName string) ([]api.Repository, error) {
	project, err := m.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	scmConfig := project.SCM
	sp, err := scm.GetSCMProvider(scmConfig)
	if err != nil {
		return nil, err
	}

	return sp.ListRepos()
}

// ListBranches lists the branches of the SCM repos authorized for the project.
func (m *projectManager) ListBranches(projectName string, repo string) ([]string, error) {
	project, err := m.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	scmConfig := project.SCM
	sp, err := scm.GetSCMProvider(scmConfig)
	if err != nil {
		return nil, err
	}

	return sp.ListBranches(repo)
}

// ListBranches lists the tags of the SCM repos authorized for the project.
func (m *projectManager) ListTags(projectName string, repo string) ([]string, error) {
	project, err := m.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	scmConfig := project.SCM
	sp, err := scm.GetSCMProvider(scmConfig)
	if err != nil {
		return nil, err
	}

	return sp.ListTags(repo)
}

// ListDockerfiles lists the dockerfiles of the SCM repos authorized for the project.
func (m *projectManager) ListDockerfiles(projectName string, repo string) ([]string, error) {
	project, err := m.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	scmConfig := project.SCM
	sp, err := scm.GetSCMProvider(scmConfig)
	if err != nil {
		return nil, err
	}

	return sp.ListDockerfiles(repo)
}

// GetTemplateType get the template type of the SCM repos authorized for the project.
func (m *projectManager) GetTemplateType(projectName string, repo string) (string, error) {
	project, err := m.GetProject(projectName)
	if err != nil {
		return "", err
	}

	scmConfig := project.SCM
	sp, err := scm.GetSCMProvider(scmConfig)
	if err != nil {
		return "", err
	}

	return sp.GetTemplateType(repo)
}

// GetStatistics gets the statistic by project name.
func (m *projectManager) GetStatistics(projectName string, start, end time.Time) (*api.PipelineStatusStats, error) {
	project, err := m.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	pipelines, err := m.GetPipelines(project.ID)
	if err != nil {
		return nil, err
	}

	totalRecords := []api.PipelineRecord{}
	for _, pipeline := range pipelines {
		// find all records ( start<={records}.startTime<end && {records}.pipelineID=pipeline.ID )
		records, _, err := m.dataStore.FindPipelineRecordsByStartTime(pipeline.ID, start, end)
		if err != nil {
			return nil, err
		}
		totalRecords = append(totalRecords, records...)
	}

	return transRecordsToStats(totalRecords, start, end)
}

func (m *projectManager) GetPipelines(projectID string) ([]api.Pipeline, error) {
	pipelines, _, err := m.dataStore.FindPipelinesByProjectID(projectID, api.QueryParams{})
	return pipelines, err
}

func (m *projectManager) DeleteProjectLogs(projectID string) error {
	// get project folder
	projectFolder := m.getProjectFolder(projectID)

	// remove project folder
	if err := os.RemoveAll(projectFolder); err != nil {
		log.Errorf("remove project folder %s error:%v", projectFolder, err)
		return err
	}

	return nil
}

// getProjectFolder gets the folder path for the project.
func (m *projectManager) getProjectFolder(projectID string) string {
	return strings.Join([]string{cycloneHome, projectID}, string(os.PathSeparator))
}
