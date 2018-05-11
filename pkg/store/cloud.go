/*
Copyright 2016 caicloud authors. All rights reserved.

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
	"github.com/caicloud/cyclone/pkg/api"
	"gopkg.in/mgo.v2/bson"
)

// InsertCloud insert a new cloud document to db
func (d *DataStore) InsertCloud(c *api.Cloud) error {
	c.ID = bson.NewObjectId().Hex()

	return d.cloudCollection.Insert(c)
}

// UpsertCloud insert a new cloud document to db
func (d *DataStore) UpsertCloud(c *api.Cloud) error {
	_, err := d.cloudCollection.Upsert(bson.M{"name": c.Name}, c)

	return err
}

// FindAllClouds returns all clouds
func (d *DataStore) FindAllClouds() ([]api.Cloud, error) {
	cs := []api.Cloud{}

	err := d.cloudCollection.Find(bson.M{}).All(&cs)
	return cs, err
}

// FindCloudByName returns a cloud in db by name
func (d *DataStore) FindCloudByName(name string) (*api.Cloud, error) {
	c := &api.Cloud{}

	err := d.cloudCollection.Find(bson.M{"name": name}).One(c)
	return c, err
}

// DeleteCloudByName delete a cloud in db by name
func (d *DataStore) DeleteCloudByName(name string) error {
	return d.cloudCollection.Remove(bson.M{"name": name})
}
