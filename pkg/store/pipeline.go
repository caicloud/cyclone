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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// CreatePipeline creates the pipeline, returns the pipeline created.
func (d *DataStore) CreatePipeline(pipeline *api.Pipeline) (*api.Pipeline, error) {
	if pipeline.ID != "" {
		pipeline.ID = bson.NewObjectId().Hex()
	}

	pipeline.CreationTime = time.Now()
	pipeline.LastUpdateTime = time.Now()
	err := d.pipelineCollection.Insert(pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

// FindPipelineByName finds the pipeline by name in one project. If find no pipeline or more than one pipeline, return
// error.
func (d *DataStore) FindPipelineByName(projectID string, name string) (*api.Pipeline, error) {
	query := bson.M{"projectID": projectID, "name": name}
	count, err := d.pipelineCollection.Find(query).Count()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, mgo.ErrNotFound
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

// FindPipelineByServiceID finds the pipeline by service id.
func (d *DataStore) FindPipelineByServiceID(serviceID string) (*api.Pipeline, error) {
	query := bson.M{"serviceID": serviceID}
	pipeline := &api.Pipeline{}
	err := d.pipelineCollection.Find(query).One(pipeline)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

// FindPipelinesByProjectID finds the pipelines by project id. Will returns all pipelines in this project.
func (d *DataStore) FindPipelinesByProjectID(projectID string, queryParams api.QueryParams) ([]api.Pipeline, int, error) {
	pipelines := []api.Pipeline{}
	query := bson.M{"projectID": projectID}
	collection := d.pipelineCollection.Find(query)

	count, err := collection.Count()
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return pipelines, count, nil
	}

	if queryParams.Start > 0 {
		collection.Skip(queryParams.Start)
	}
	if queryParams.Limit > 0 {
		collection.Limit(queryParams.Limit)
	}

	if err = collection.All(&pipelines); err != nil {
		return nil, 0, err
	}

	return pipelines, count, nil
}

// UpdatePipeline updates the pipeline, please make sure the pipeline id is provided before call this method.
func (d *DataStore) UpdatePipeline(pipeline *api.Pipeline) error {
	pipeline.LastUpdateTime = time.Now()

	return d.pipelineCollection.UpdateId(pipeline.ID, pipeline)
}

// DeletePipelineByID deletes the pipeline by id.
func (d *DataStore) DeletePipelineByID(pipelineID string) error {
	return d.pipelineCollection.RemoveId(pipelineID)
}

// DeletePipelinesByProjectID deletes all the pipelines in one project by project id.
func (d *DataStore) DeletePipelinesByProjectID(projectID string) error {
	return d.pipelineCollection.Remove(bson.M{"projectID": projectID})
}
