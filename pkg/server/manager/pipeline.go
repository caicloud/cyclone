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
	"sync"

	"github.com/zoumo/logdog"
	"gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/event"
	"github.com/caicloud/cyclone/pkg/store"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/remote"
)

// PipelineManager represents the interface to manage pipeline.
type PipelineManager interface {
	CreatePipeline(projectName string, pipeline *api.Pipeline) (*api.Pipeline, error)
	GetPipeline(projectName string, pipelineName string) (*api.Pipeline, error)
	ListPipelines(projectName string, queryParams api.QueryParams, recentCount, recentSuccessCount, recentFailedCount int) ([]api.Pipeline, int, error)
	UpdatePipeline(projectName string, pipelineName string, newPipeline *api.Pipeline) (*api.Pipeline, error)
	DeletePipeline(projectName string, pipelineName string) error
	ClearPipelinesOfProject(projectName string) error
}

// pipelineManager represents the manager for pipeline.
type pipelineManager struct {
	dataStore             *store.DataStore
	remoteManager         *remote.Manager
	pipelineRecordManager PipelineRecordManager

	// TODO (robin) Move event manager to pipeline record manager.
	eventManager event.EventManager
}

// NewPipelineManager creates a pipeline manager.
func NewPipelineManager(dataStore *store.DataStore, pipelineRecordManager PipelineRecordManager) (PipelineManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new pipeline manager as data store is nil")
	}

	remoteManager := remote.NewManager()

	if pipelineRecordManager == nil {
		return nil, fmt.Errorf("Fail to new pipeline manager as pipeline record is nil")
	}

	eventManager := event.NewEventManager(dataStore)

	return &pipelineManager{dataStore, remoteManager, pipelineRecordManager, eventManager}, nil
}

// CreatePipeline creates a pipeline.
func (m *pipelineManager) CreatePipeline(projectName string, pipeline *api.Pipeline) (*api.Pipeline, error) {
	// Check the existence of the project and pipeline.
	if _, err := m.GetPipeline(projectName, pipeline.Name); err == nil {
		return nil, httperror.ErrorAlreadyExist.Format(pipeline.Name)
	}

	createdPipeline, err := m.dataStore.CreatePipeline(pipeline)
	if err != nil {
		return nil, err
	}

	return createdPipeline, nil
}

// GetPipeline gets the pipeline by name in one project.
func (m *pipelineManager) GetPipeline(projectName string, pipelineName string) (*api.Pipeline, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format(projectName)
		}

		return nil, err
	}

	pipeline, err := m.dataStore.FindPipelineByName(project.ID, pipelineName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format(pipelineName)
		}

		return nil, err
	}

	return pipeline, nil
}

// ListPipelines lists all pipelines in one project.
func (m *pipelineManager) ListPipelines(projectName string, queryParams api.QueryParams,
	recentCount, recentSuccessCount, recentFailedCount int) ([]api.Pipeline, int, error) {
	ds := m.dataStore

	project, err := ds.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, httperror.ErrorContentNotFound.Format(projectName)
		}
		return nil, 0, err
	}

	pipelines, total, err := ds.FindPipelinesByProjectID(project.ID, queryParams)
	if err != nil {
		return nil, 0, err
	}

	if recentCount <= 0 && recentSuccessCount <= 0 && recentFailedCount <= 0 {
		return pipelines, total, nil
	}

	wg := sync.WaitGroup{}
	for i, _ := range pipelines {
		wg.Add(1)

		go func(pipeline *api.Pipeline) {
			defer wg.Done()

			if recentCount > 0 {
				recentRecords, _, err := ds.FindRecentRecordsByPipelineID(pipeline.ID, nil, recentCount)
				if err != nil {
					logdog.Error(err)
				} else {
					pipeline.RecentRecords = recentRecords
				}
			}

			if recentSuccessCount > 0 {
				filter := map[string]interface{}{
					"status": api.Success,
				}
				recentSuccessRecords, _, err := ds.FindRecentRecordsByPipelineID(pipeline.ID, filter, recentSuccessCount)
				if err != nil {
					logdog.Error(err)
				} else {
					pipeline.RecentSuccessRecords = recentSuccessRecords
				}
			}

			if recentFailedCount > 0 {
				filter := map[string]interface{}{
					"status": api.Failed,
				}
				recentFailedRecords, _, err := ds.FindRecentRecordsByPipelineID(pipeline.ID, filter, recentFailedCount)
				if err != nil {
					logdog.Error(err)
				} else {
					pipeline.RecentFailedRecords = recentFailedRecords
				}
			}
		}(&pipelines[i])
	}
	wg.Wait()

	return pipelines, total, nil
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

	if len(newPipeline.Alias) > 0 {
		pipeline.Alias = newPipeline.Alias
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

	return m.deletePipeline(pipeline)
}

// ClearPipelinesOfProject deletes all pipelines in one project.
func (m *pipelineManager) ClearPipelinesOfProject(projectName string) error {
	pipelines, _, err := m.ListPipelines(projectName, api.QueryParams{}, 0, 0, 0)
	if err != nil {
		return nil
	}

	for _, pipeline := range pipelines {
		if err := m.deletePipeline(&pipeline); err != nil {
			return err
		}
	}

	return nil
}

// deletePipeline deletes the pipeline.
func (m *pipelineManager) deletePipeline(pipeline *api.Pipeline) error {
	ds := m.dataStore
	var err error

	// Delete the pipeline records of this pipeline.
	if err = m.pipelineRecordManager.ClearPipelineRecordsOfPipeline(pipeline.ID); err != nil {
		return fmt.Errorf("Fail to delete all pipeline records for pipeline %s as %s", pipeline.Name, err.Error())
	}

	if err = ds.DeletePipelineByID(pipeline.ID); err != nil {
		return fmt.Errorf("Fail to delete the pipeline %s as %s", pipeline.Name, err.Error())
	}

	return nil
}
