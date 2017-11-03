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
	"strings"

	log "github.com/zoumo/logdog"
	"gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/scm"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/store"
)

// ProjectManager represents the interface to manage project.
type ProjectManager interface {
	CreateProject(project *api.Project) (*api.Project, error)
	GetProject(projectName string) (*api.Project, error)
	ListProjects(queryParams api.QueryParams) ([]api.Project, int, error)
	UpdateProject(projectName string, newProject *api.Project) (*api.Project, error)
	DeleteProject(projectName string) error
	ListRepos(projectName string) ([]api.Repository, error)
	ListBranches(projectName string, repo string) ([]string, error)
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
	projectName := project.Name
	if _, err := m.GetProject(projectName); err == nil {
		return nil, httperror.ErrorAlreadyExist.Format(projectName)
	}

	if err := generateSCMToken(project.SCM); err != nil {
		return nil, err
	}

	return m.dataStore.CreateProject(project)
}

// GetProject gets the project by name.
func (m *projectManager) GetProject(projectName string) (*api.Project, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format(projectName)
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

	if err := generateSCMToken(newProject.SCM); err != nil {
		return nil, err
	}
	project.SCM = newProject.SCM

	// Update the properties of the project.
	// TODO (robin) Whether need a method for this merge?
	if len(newProject.Name) > 0 {
		project.Name = newProject.Name
	}

	if len(newProject.Description) > 0 {
		project.Description = newProject.Description
	}

	if len(newProject.Owner) > 0 {
		project.Owner = newProject.Owner
	}

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
	sp, err := scm.GetSCMProvider(scmConfig.Type)
	if err != nil {
		return nil, err
	}

	return sp.ListRepos(scmConfig)
}

// ListBranches lists the branches of the SCM repos authorized for the project.
func (m *projectManager) ListBranches(projectName string, repo string) ([]string, error) {
	project, err := m.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	scmConfig := project.SCM
	sp, err := scm.GetSCMProvider(scmConfig.Type)
	if err != nil {
		return nil, err
	}

	return sp.ListBranches(scmConfig, repo)
}

func generateSCMToken(config *api.SCMConfig) error {
	if config == nil {
		return httperror.ErrorContentNotFound.Format("SCM config")
	}

	// Trim suffix '/' of Gitlab server to ensure that the token can work, otherwise there will be 401 error.
	config.Server = strings.TrimSuffix(config.Server, "/")

	// Get the SCM token when SCM is Github or Gitlab, and the username is provided.
	if config.Type != api.SVN && len(config.Username) != 0 {
		provider, err := scm.GetSCMProvider(config.Type)
		if err != nil {
			return err
		}

		token, err := provider.GetToken(config)
		if err != nil {
			log.Errorf("fail to get SCM token for user %s as %s", config.Username, err.Error())
			return err
		}

		config.Token = token
		// Cleanup the password for security.
		config.Password = ""
	}

	return nil
}
