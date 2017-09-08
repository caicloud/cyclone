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

// FindPipelineRecordsByPipelineID finds the pipeline records by pipelineID.
func (d *DataStore) FindPipelineRecordsByPipelineID(pipelineID string, queryParams api.QueryParams) ([]api.PipelineRecord, int, error) {
	pipelineRecords := []api.PipelineRecord{}
	query := bson.M{"pipelineId": pipelineID}
	collection := d.pipelineCollection.Find(query)

	count, err := collection.Count()
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return pipelineRecords, count, nil
	}

	if queryParams.Start > 0 {
		collection.Skip(queryParams.Start)
	}
	if queryParams.Limit > 0 {
		collection.Limit(queryParams.Limit)
	}

	if err = collection.All(&pipelineRecords); err != nil {
		return nil, 0, err
	}

	return pipelineRecords, count, nil
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
	pipelineRecord.EndTime = time.Now()

	return d.pipelineRecordCollection.UpdateId(pipelineRecord.ID, pipelineRecord)
}

// DeletePipelineRecordByID deletes the pipeline record by id.
func (d *DataStore) DeletePipelineRecordByID(pipelineRecordID string) error {
	return d.pipelineRecordCollection.RemoveId(pipelineRecordID)
}

// DeletePipelineRecordsByPipelineID deletes all the pipeline records of one pipeline by pipeline id.
func (d *DataStore) DeletePipelineRecordsByPipelineID(pipelineID string) error {
	return d.pipelineRecordCollection.Remove(bson.M{"pipelineId": pipelineID})
}
