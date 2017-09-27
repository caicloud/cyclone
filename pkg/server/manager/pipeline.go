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

	"github.com/caicloud/cyclone/api/conversion"
	"github.com/caicloud/cyclone/event"
	"github.com/caicloud/cyclone/pkg/api"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
	"github.com/caicloud/cyclone/remote"
	"github.com/caicloud/cyclone/store"
	"github.com/zoumo/logdog"
	"gopkg.in/mgo.v2"
)

// PipelineManager represents the interface to manage pipeline.
type PipelineManager interface {
	CreatePipeline(projectName string, pipeline *api.Pipeline) (*api.Pipeline, error)
	GetPipeline(projectName string, pipelineName string) (*api.Pipeline, error)
	ListPipelines(projectName string, queryParams api.QueryParams) ([]api.Pipeline, int, error)
	UpdatePipeline(projectName string, pipelineName string, newPipeline *api.Pipeline) (*api.Pipeline, error)
	DeletePipeline(projectName string, pipelineName string) error
	ClearPipelinesOfProject(projectName string) error
	PerformPipeline(projectName string, pipelineName string, performParams *api.PipelinePerformParams) error
}

// pipelineManager represents the manager for pipeline.
type pipelineManager struct {
	dataStore             *store.DataStore
	remoteManager         *remote.Manager
	pipelineRecordManager PipelineRecordManager
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

	return &pipelineManager{dataStore, remoteManager, pipelineRecordManager}, nil
}

// CreatePipeline creates a pipeline.
func (m *pipelineManager) CreatePipeline(projectName string, pipeline *api.Pipeline) (*api.Pipeline, error) {
	// Check the existence of the project and pipeline.
	if _, err := m.GetPipeline(projectName, pipeline.Name); err == nil {
		return nil, httperror.ErrorAlreadyExist.Format(pipeline.Name)
	}

	// TODO (robin) Remove the creation of service for pipeline after replace service with pipeline.
	service, err := conversion.ConvertPipelineToService(projectName, pipeline)
	if err != nil {
		return nil, fmt.Errorf("Fail to generate service for pipeline %s as %s", pipeline.Name, err.Error())
	}

	serviceID, err := m.dataStore.NewServiceDocument(service)
	if err != nil {
		return nil, fmt.Errorf("Fail to create service for pipeline %s as %s", pipeline.Name, err.Error())
	}

	err = event.SendCreateServiceEvent(service)
	if err != nil {
		return nil, fmt.Errorf("Fail to create service event for pipeline %s as %s", pipeline.Name, err.Error())
	}

	pipeline.ServiceID = serviceID

	createdPipeline, err := m.dataStore.CreatePipeline(pipeline)
	if err != nil {
		// Delete the service if fail to create pipeline.
		if err := m.dataStore.DeleteServiceByID(serviceID); err != nil {
			logdog.Error(err)
		}

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
func (m *pipelineManager) ListPipelines(projectName string, queryParams api.QueryParams) ([]api.Pipeline, int, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, httperror.ErrorContentNotFound.Format(projectName)
		}
		return nil, 0, err
	}

	return m.dataStore.FindPipelinesByProjectID(project.ID, queryParams)
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

	// TODO (robin) Remove the updating of service for pipeline after replace service with pipeline.
	service, err := m.dataStore.FindServiceByID(pipeline.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("Fail to find service for pipeline %s as %s", pipeline.Name, err.Error())
	}

	newService, err := conversion.ConvertPipelineToService(projectName, pipeline)
	if err != nil {
		return nil, fmt.Errorf("Fail to generate service for pipeline %s as %s", pipeline.Name, err.Error())
	}

	newService.ServiceID = service.ServiceID
	newService.Versions = service.Versions
	newService.VersionFails = service.VersionFails
	newService.LastVersionName = service.LastVersionName


	// Judge the change of repository url, if not change, just keep the status, if change, need to send event to
	// check the new status.
	if newService.Repository.URL == service.Repository.URL {
		newService.Repository.Status = service.Repository.Status
	} else {
		// Delete the old webhook.
		remote, err := m.remoteManager.FindRemote(service.Repository.Webhook)
		if err != nil {
			logdog.Error(err.Error())
		} else {
			if err := remote.DeleteHook(service); err != nil {
				logdog.Errorf("Fail to delete the webhook for pipeline %s as %s", pipeline.Name, err.Error())
			}
		}

		err = event.SendCreateServiceEvent(newService)
		if err != nil {
			return nil, fmt.Errorf("Fail to create service event for pipeline %s as %s", pipeline.Name, err.Error())
		}
	}
	_, err = m.dataStore.UpsertServiceDocument(newService)
	if err != nil {
		return nil, fmt.Errorf("Fail to update service for pipeline %s as %s", pipeline.Name, err.Error())
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
	pipelines, _, err := m.ListPipelines(projectName, api.QueryParams{})
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

// deletePipeline deletes the pipeline and its related services and versions. It can be removed after replace service
// with pipeline.
func (m *pipelineManager) deletePipeline(pipeline *api.Pipeline) error {
	ds := m.dataStore

	// Delete the versions related to this pipeline.
	versions, err := ds.FindVersionsByServiceID(pipeline.ServiceID)
	if err != nil {
		return err
	}

	for _, version := range versions {
		if err := ds.DeleteVersionByID(version.VersionID); err != nil {
			return fmt.Errorf("Fail to delete the versions for pipeline %s as %s", pipeline.Name, err.Error())
		}
	}

	if service, err := ds.FindServiceByID(pipeline.ServiceID); err == nil {
		// Delete the webhook.
		remote, err := m.remoteManager.FindRemote(service.Repository.Webhook)
		if err != nil {
			logdog.Error(err.Error())
		} else {
			if err := remote.DeleteHook(service); err != nil {
				logdog.Errorf("Fail to delete the webhook for pipeline %s as %s", pipeline.Name, err.Error())
			}
		}

		// Delete the service related to this pipeline
		err = ds.DeleteServiceByID(pipeline.ServiceID)
		if err != nil {
			return fmt.Errorf("Fail to delete the service for pipeline %s as %s", pipeline.Name, err.Error())
		}
	}

	// Delete the pipeline records of this pipeline.
	if err = m.pipelineRecordManager.ClearPipelineRecordsOfPipeline(pipeline.ID); err != nil {
		return fmt.Errorf("Fail to delete all pipeline records for pipeline %s as %s", pipeline.Name, err.Error())
	}

	if err = ds.DeletePipelineByID(pipeline.ID); err != nil {
		return fmt.Errorf("Fail to delete the pipeline %s as %s", pipeline.Name, err.Error())
	}

	return nil
}

// PerformPipeline performs the pipeline with params.
func (m *pipelineManager) PerformPipeline(projectName string, pipelineName string, performParams *api.PipelinePerformParams) error {
	// Find the pipeline in the project.
	pipeline, err := m.GetPipeline(projectName, pipelineName)
	if err != nil {
		return err
	}

	ds := m.dataStore
	service, err := ds.FindServiceByID(pipeline.ServiceID)
	if err != nil {
		return fmt.Errorf("Fail to find the related service of the pipeline %s", pipelineName)
	}

	version := conversion.ConvertPipelineParamsToVersion(performParams)
	version.ServiceID = service.ServiceID

	if _, err = ds.NewVersionDocument(version); err != nil {
		return err
	}

	return event.SendCreateVersionEvent(service, version)
}
