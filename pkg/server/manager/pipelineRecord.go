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

// PipelineRecordManager represents the interface to manage pipeline record.
type PipelineRecordManager interface {
	CreatePipelineRecord(pipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error)
	GetPipelineRecord(pipelineRecordID string) (*api.PipelineRecord, error)
	ListPipelineRecords(projectName string, pipelineName string, queryParams api.QueryParams) ([]api.PipelineRecord, int, error)
	UpdatePipelineRecord(pipelineRecordID string, pipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error)
	DeletePipelineRecord(pipelineRecordID string) error
	ClearPipelineRecordsOfPipeline(pipelineID string) error
}

// pipelineRecordManager represents the manager for pipeline record.
type pipelineRecordManager struct {
	dataStore *store.DataStore
}

// NewPipelineRecordManager creates a pipeline record manager.
func NewPipelineRecordManager(dataStore *store.DataStore) (PipelineRecordManager, error) {
	if dataStore == nil {
		return nil, fmt.Errorf("Fail to new pipeline record manager as data store is nil")
	}

	return &pipelineRecordManager{dataStore}, nil
}

// CreatePipelineRecord creates a pipeline record.
func (m *pipelineRecordManager) CreatePipelineRecord(pipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error) {
	return m.dataStore.CreatePipelineRecord(pipelineRecord)
}

// GetPipelineRecord gets the pipeline record by id.
func (m *pipelineRecordManager) GetPipelineRecord(pipelineRecordID string) (*api.PipelineRecord, error) {
	return m.dataStore.FindPipelineRecordByID(pipelineRecordID)
}

// ListPipelineRecords finds the pipeline records by pipelineID.
func (m *pipelineRecordManager) ListPipelineRecords(projectName string, pipelineName string, queryParams api.QueryParams) ([]api.PipelineRecord, int, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		return nil, 0, err
	}

	pipeline, err := m.dataStore.FindPipelineByName(project.ID, pipelineName)
	if err != nil {
		return nil, 0, err
	}

	return m.dataStore.FindPipelineRecordsByPipelineID(pipeline.ID, queryParams)
}

// UpdatePipelineRecord updates pipeline record by id.
func (m *pipelineRecordManager) UpdatePipelineRecord(pipelineRecordID string, newPipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error) {
	pipelineRecord, err := m.dataStore.FindPipelineRecordByID(pipelineRecordID)
	if err != nil {
		return nil, err
	}

	// Update the properties of the pipeline record.
	if newPipelineRecord.Status != "" {
		pipelineRecord.Status = newPipelineRecord.Status
	}
	if newPipelineRecord.StageStatus != nil {
		pipelineRecord.StageStatus = newPipelineRecord.StageStatus
	}

	if err = m.dataStore.UpdatePipelineRecord(pipelineRecord); err != nil {
		return nil, err
	}

	return pipelineRecord, nil
}

// DeletePipelineRecord deletes the pipeline record by id.
func (m *pipelineRecordManager) DeletePipelineRecord(pipelineRecordID string) error {
	return m.dataStore.DeletePipelineRecordByID(pipelineRecordID)
}

// ClearPipelineRecordsOfPipeline deletes all the pipeline records of one pipeline by pipeline id.
func (m *pipelineRecordManager) ClearPipelineRecordsOfPipeline(pipelineID string) error {
	return m.dataStore.DeletePipelineRecordsByPipelineID(pipelineID)
}
