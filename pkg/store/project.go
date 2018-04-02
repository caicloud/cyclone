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
	"sync"
	"time"

	log "github.com/golang/glog"
	"github.com/mozillazg/go-slugify"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/cyclone/pkg/api"
)

// CreateProject creates the project, returns the project created.
func (d *DataStore) CreateProject(project *api.Project) (*api.Project, error) {
	if project.Name == "" && project.Alias != "" {
		project.Name = slugify.Slugify(project.Alias)
	}

	project.ID = bson.NewObjectId().Hex()
	project.CreationTime = time.Now()
	project.LastUpdateTime = time.Now()

	// Encrypt the passwords.
	if err := encryptPasswordsForProjects(project, d.saltKey); err != nil {
		return nil, err
	}

	if err := d.projectCollection.Insert(project); err != nil {
		return nil, err
	}

	return project, nil
}

// FindProjectByName finds the project by name. If find no project or more than one project, return error.
func (d *DataStore) FindProjectByName(name string) (*api.Project, error) {
	query := bson.M{"name": name}
	count, err := d.projectCollection.Find(query).Count()
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, mgo.ErrNotFound
	} else if count > 1 {
		return nil, fmt.Errorf("there are %d projects with the same name %s", count, name)
	}

	project := &api.Project{}
	if err = d.projectCollection.Find(query).One(project); err != nil {
		return nil, err
	}

	// Decrypt the passwords.
	if err := decryptPasswordsForProjects(project, d.saltKey); err != nil {
		return nil, err
	}

	return project, nil
}

// FindProjectByID finds the project by id.
func (d *DataStore) FindProjectByID(projectID string) (*api.Project, error) {
	project := &api.Project{}
	if err := d.projectCollection.FindId(projectID).One(project); err != nil {
		return nil, err
	}

	// Decrypt the passwords.
	if err := decryptPasswordsForProjects(project, d.saltKey); err != nil {
		return nil, err
	}

	return project, nil
}

// UpdateProject updates the project, please make sure the project id is provided before call this method.
func (d *DataStore) UpdateProject(project *api.Project) error {
	if project.Name == "" && project.Alias != "" {
		project.Name = slugify.Slugify(project.Alias)
	}

	// Encrypt the passwords.
	if err := encryptPasswordsForProjects(project, d.saltKey); err != nil {
		return err
	}

	project.LastUpdateTime = time.Now()

	return d.projectCollection.UpdateId(project.ID, project)
}

// DeleteProjectByID deletes the project by id.
func (d *DataStore) DeleteProjectByID(projectID string) error {
	return d.projectCollection.RemoveId(projectID)
}

// GetProjects gets all projects. Will returns all projects.
func (d *DataStore) GetProjects(queryParams api.QueryParams) ([]api.Project, int, error) {
	projects := []api.Project{}
	collection := d.projectCollection.Find(nil)

	count, err := collection.Count()
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return projects, count, nil
	}

	if queryParams.Start > 0 {
		collection.Skip(queryParams.Start)
	}
	if queryParams.Limit > 0 {
		collection.Limit(queryParams.Limit)
	}

	if err = collection.All(&projects); err != nil {
		return nil, 0, err
	}

	wg := sync.WaitGroup{}
	for i := range projects {
		wg.Add(1)
		go func(p *api.Project) {
			defer wg.Done()
			if err := decryptPasswordsForProjects(p, d.saltKey); err != nil {
				log.Errorf("fail to decrypt passwords for project %s as %v", p.Name, err)
			}
		}(&projects[i])
	}

	wg.Wait()

	return projects, count, nil
}
