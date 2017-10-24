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

	"github.com/zoumo/logdog"
	"gopkg.in/mgo.v2"

	"github.com/caicloud/cyclone/api/conversion"
	"github.com/caicloud/cyclone/pkg/api"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
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
	GetPipelineRecordLogs(pipelineRecordID string) (string, error)
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
	version, err := m.dataStore.FindVersionByID(pipelineRecordID)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, httperror.ErrorContentNotFound.Format(fmt.Sprintf("pipeline record %s", pipelineRecordID))
		}
		return nil, err
	}

	pipelineRecord, err := conversion.ConvertVersionToPipelineRecord(version)
	if err != nil {
		return nil, err
	}

	return pipelineRecord, nil
}

// ListPipelineRecords finds the pipeline records by pipeline id.
func (m *pipelineRecordManager) ListPipelineRecords(projectName string, pipelineName string, queryParams api.QueryParams) ([]api.PipelineRecord, int, error) {
	project, err := m.dataStore.FindProjectByName(projectName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, httperror.ErrorContentNotFound.Format(projectName)
		}
		return nil, 0, err
	}

	pipeline, err := m.dataStore.FindPipelineByName(project.ID, pipelineName)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, 0, httperror.ErrorContentNotFound.Format(pipelineName)
		}
		return nil, 0, err
	}

	versions, total, err := m.dataStore.FindVersionsWithPaginationByServiceID(pipeline.ServiceID, queryParams.Filter, queryParams.Start, queryParams.Limit)
	if err != nil {
		return nil, 0, err
	}

	pipelineRecords := []api.PipelineRecord{}
	for _, version := range versions {
		record, err := conversion.ConvertVersionToPipelineRecord(&version)
		if err != nil {
			return nil, 0, err
		}
		pipelineRecords = append(pipelineRecords, *record)
	}

	return pipelineRecords, total, nil
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
	return m.dataStore.DeleteVersionByID(pipelineRecordID)
}

// ClearPipelineRecordsOfPipeline deletes all the pipeline records of one pipeline by pipeline id.
func (m *pipelineRecordManager) ClearPipelineRecordsOfPipeline(pipelineID string) error {
	ds := m.dataStore

	pipeline, err := ds.FindPipelineByID(pipelineID)
	if err != nil {
		return err
	}

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

	return nil
}

// GetPipelineRecordLogs gets the pipeline record logs by id.
func (m *pipelineRecordManager) GetPipelineRecordLogs(pipelineRecordID string) (string, error) {
	pipelineRecord, err := m.GetPipelineRecord(pipelineRecordID)
	if err != nil {
		return "", err
	}

	logdog.Debugf("Pipeline record is %s", pipelineRecord)

	status := pipelineRecord.Status
	if status != api.Success && status != api.Failed {
		return "", fmt.Errorf("Can not get the logs as pipeline record %s is %s, please try after it finishes",
			pipelineRecordID, status)
	}

	log, err := m.dataStore.FindVersionLogByVersionID(pipelineRecordID)
	if err != nil {
		if err == mgo.ErrNotFound {
			return "", httperror.ErrorContentNotFound.Format(fmt.Sprintf("log of pipeline record %s", pipelineRecordID))
		}
		return "", fmt.Errorf("Fail to get the log for pipeline record %s as %s", pipelineRecordID, err.Error())
	}

	return log.Logs, nil
}
