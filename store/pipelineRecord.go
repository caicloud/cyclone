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

package store

import (
	"time"

	"github.com/caicloud/cyclone/pkg/api"
	"gopkg.in/mgo.v2/bson"
)

// CreatePipelineRecord creates the pipeline record, returns the pipeline record created.
func (d *DataStore) CreatePipelineRecord(pipelineRecord *api.PipelineRecord) (*api.PipelineRecord, error) {
	pipelineRecord.ID = bson.NewObjectId().Hex()
	pipelineRecord.StartTime = time.Now()

	if err := d.pipelineRecordCollection.Insert(pipelineRecord); err != nil {
		return nil, err
	}

	return pipelineRecord, nil
}

// FindPipelineRecordsByPipelineID finds the pipeline record by pipelineID.
func (d *DataStore) FindPipelineRecordsByPipelineID(pipelineID string) ([]api.PipelineRecord, error) {
	query := bson.M{"pipelineID": pipelineID}

	pipelineRecords := []api.PipelineRecord{}
	if err := d.pipelineRecordCollection.FindId(query).All(pipelineRecords); err != nil {
		return nil, err
	}

	return pipelineRecords, nil
}

// FindPipelineRecordByID finds the pipeline record by id.
func (d *DataStore) FindPipelineRecordByID(pipelineRecordID string) (*api.PipelineRecord, error) {
	pipelineRecord := &api.PipelineRecord{}
	if err := d.pipelineRecordCollection.FindId(pipelineRecordID).One(pipelineRecord); err != nil {
		return nil, err
	}

	return pipelineRecord, nil
}

// UpdatePipelineRecord updates the pipeline record.
func (d *DataStore) UpdatePipelineRecord(pipelineRecord *api.PipelineRecord) error {
	updatedPipelineRecord := *pipelineRecord
	updatedPipelineRecord.EndTime = time.Now()

	return d.pipelineRecordCollection.UpdateId(pipelineRecord.ID, updatedPipelineRecord)
}

// DeletePipelineRecord deletes the pipeline record by id.
func (d *DataStore) DeletePipelineRecord(pipelineRecordID string) error {
	return d.pipelineRecordCollection.RemoveId(pipelineRecordID)
}
