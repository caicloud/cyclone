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

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/store"
	"github.com/zoumo/logdog"
)

// ProjectManager represents the interface to manage project.
type ProjectManager interface {
	CreateProject(project *api.Project) (*api.Project, error)
	GetProject(projectName string) (*api.Project, error)
	UpdateProject(projectName string, newProject *api.Project) (*api.Project, error)
	DeleteProject(projectName string) error
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
	return m.dataStore.CreateProject(project)
}

// GetProject gets the project by name.
func (m *projectManager) GetProject(projectName string) (*api.Project, error) {
	return m.dataStore.FindProjectByName(projectName)
}

// UpdateProject updates the project by name.
func (m *projectManager) UpdateProject(projectName string, newProject *api.Project) (*api.Project, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		return nil, err
	}

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
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		return err
	}

	// Delete the pipelines in this project.
	if err = m.pipelineManager.ClearPipelinesOfProject(project.ID); err != nil {
		logdog.Errorf("Fail to delete all pipelines in the project %s as %s", projectName, err.Error())
		return err
	}

	if err = m.dataStore.DeleteProjectByID(project.ID); err != nil {
		logdog.Errorf("Fail to delete the project %s as %s", projectName, err.Error())
		return err
	}

	return nil
}
