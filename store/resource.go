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
	"github.com/caicloud/cyclone/api"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// NewResourceDocument creates a new document (record) in mongodb.
func (d *DataStore) NewResourceDocument(resource *api.Resource) error {
	col := d.s.DB(defaultDBName).C(ResourceCollectionName)
	_, err := col.Upsert(bson.M{"_id": resource.UserID}, resource)
	return err
}

// UpdateResourceDocument update a document (record) in mongodb.
func (d *DataStore) UpdateResourceDocument(resource *api.Resource) error {
	col := d.s.DB(defaultDBName).C(ResourceCollectionName)
	_, err := col.Upsert(bson.M{"_id": resource.UserID}, resource)
	return err
}

// FindResourceByID finds a resource entity by userID.
func (d *DataStore) FindResourceByID(userID string) (*api.Resource, error) {
	col := d.s.DB(defaultDBName).C(ResourceCollectionName)
	resource := &api.Resource{}
	err := col.Find(bson.M{"_id": userID}).One(resource)
	return resource, err
}

// UpdateResourceStatus updates resource's memory and cpu.
func (d *DataStore) UpdateResourceStatus(userID string, memory float64, cpu float64) error {
	col := d.s.DB(defaultDBName).C(ResourceCollectionName)
	filter := bson.M{"_id": userID}
	change := mgo.Change{
		Update: bson.M{"$set": bson.M{"left_resource": api.BuildResource{memory, cpu}}},
	}

	resource := api.Resource{}
	_, err := col.Find(filter).Apply(change, &resource)
	return err
}
