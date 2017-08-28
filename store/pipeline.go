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
	"fmt"
	"time"

	"github.com/caicloud/cyclone/pkg/api"
	"gopkg.in/mgo.v2/bson"
)

// CreatePipeline creates the pipeline, returns the pipeline created.
func (d *DataStore) CreatePipeline(pipeline *api.Pipeline) (*api.Pipeline, error) {
	pipeline.ID = bson.NewObjectId().Hex()
	pipeline.CreatedTime = time.Now()
	pipeline.UpdatedTime = time.Now()
	err := d.pipelineCollection.Insert(pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

// FindPipelineByName finds the pipeline by name in one project. If find no pipeline or more than one pipeline, return
// error.
func (d *DataStore) FindPipelineByName(projectID string, name string) (*api.Pipeline, error) {
	query := bson.M{"projectId": projectID, "name": name}
	count, err := d.pipelineCollection.Find(query).Count()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("there is no pipeline with name %s", name)
	} else if count > 1 {
		return nil, fmt.Errorf("there are %d pipelines with the same name %s", count, name)
	}

	pipeline := &api.Pipeline{}
	err = d.pipelineCollection.Find(query).One(pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

// FindPipelineByID finds the pipeline by id.
func (d *DataStore) FindPipelineByID(pipelineID string) (*api.Pipeline, error) {
	pipeline := &api.Pipeline{}
	err := d.pipelineCollection.FindId(pipelineID).One(pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

// UpdatePipeline updates the pipeline.
func (d *DataStore) UpdatePipeline(pipeline *api.Pipeline) error {
	updatedPipeline := *pipeline
	updatedPipeline.UpdatedTime = time.Now()

	return d.pipelineCollection.UpdateId(pipeline.ID, updatedPipeline)
}

// DeletePipelineByID deletes the pipeline by id.
func (d *DataStore) DeletePipelineByID(pipelineID string) error {
	return d.pipelineCollection.RemoveId(pipelineID)
}
