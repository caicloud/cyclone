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
)

// PipelineManager represents the interface to manage pipeline.
type PipelineManager interface {
	CreatePipeline(pipeline *api.Pipeline) (*api.Pipeline, error)
	GetPipeline(projectName string, pipelineName string) (*api.Pipeline, error)
	ListPipelines(projectName string) ([]api.Pipeline, error)
	UpdatePipeline(projectName string, pipelineName string, newPipeline *api.Pipeline) (*api.Pipeline, error)
	DeletePipeline(projectName string, pipelineName string) error
	ClearPipelinesOfProject(projectName string) error
}

// pipelineManager represents the manager for pipeline.
type pipelineManager struct {
	dataStore *store.DataStore
}

// NewPipelineManager creates a pipeline manager.
func NewPipelineManager(dataStore *store.DataStore) (PipelineManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new pipeline manager as data store is nil.")
	}

	return &pipelineManager{dataStore}, nil
}

// CreatePipeline creates a pipeline.
func (m *pipelineManager) CreatePipeline(pipeline *api.Pipeline) (*api.Pipeline, error) {
	return m.dataStore.CreatePipeline(pipeline)
}

// GetPipeline gets the pipeline by name in one project.
func (m *pipelineManager) GetPipeline(projectName string, pipelineName string) (*api.Pipeline, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		return nil, err
	}

	return m.dataStore.FindPipelineByName(project.ID, pipelineName)
}

// ListPipelines lists all pipelines in one project.
func (m *pipelineManager) ListPipelines(projectName string) ([]api.Pipeline, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		return nil, err
	}

	return m.dataStore.FindPipelinesByProjectID(project.ID)
}

// UpdatePipeline updates the pipeline by name in one project.
func (m *pipelineManager) UpdatePipeline(projectName string, pipelineName string, newPipeline *api.Pipeline) (*api.Pipeline, error) {
	pipeline, err := m.GetPipeline(projectName, pipelineName)
	if err != nil {
		return nil, err
	}

	// Update the properties of the pipeline.
	// TODO (robin) Whether need a method for this merge?
	if len(newPipeline.Name) > 0 {
		pipeline.Name = newPipeline.Name
	}

	if len(newPipeline.Description) > 0 {
		pipeline.Description = newPipeline.Description
	}

	if len(newPipeline.Owner) > 0 {
		pipeline.Owner = newPipeline.Owner
	}

	if newPipeline.Build != nil {
		pipeline.Build = newPipeline.Build
	}

	if newPipeline.AutoTrigger != nil {
		pipeline.AutoTrigger = newPipeline.AutoTrigger
	}

	if err = m.dataStore.UpdatePipeline(pipeline); err != nil {
		return nil, err
	}

	return pipeline, nil
}

// DeletePipeline deletes the pipeline by name in one project.
func (m *pipelineManager) DeletePipeline(projectName string, pipelineName string) error {
	pipeline, err := m.GetPipeline(projectName, pipelineName)
	if err != nil {
		return err
	}

	return m.dataStore.DeletePipelineByID(pipeline.ID)
}

// ClearPipelinesOfProject deletes all pipelines in one project.
func (m *pipelineManager) ClearPipelinesOfProject(projectName string) error {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		return err
	}

	return m.dataStore.DeletePipelinesByProjectID(project.ID)
}
