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
	"github.com/caicloud/cyclone/cloud"
	"gopkg.in/mgo.v2/bson"
)

// InsertCloud insert a new cloud document to db
func (d *DataStore) InsertCloud(doc *cloud.Options) error {
	col := d.s.DB(defaultDBName).C(cloudCollection)
	err := col.Insert(doc)
	return err
}

// UpsertCloud insert a new cloud document to db
func (d *DataStore) UpsertCloud(doc *cloud.Options) error {
	col := d.s.DB(defaultDBName).C(cloudCollection)
	_, err := col.Upsert(bson.M{"name": doc.Name}, doc)
	return err
}

// FindAllClouds returns all clouds
func (d *DataStore) FindAllClouds() ([]cloud.Options, error) {
	clouds := []cloud.Options{}

	col := d.s.DB(defaultDBName).C(cloudCollection)
	err := col.Find(bson.M{}).All(&clouds)
	return clouds, err
}

// FindCloudByName returns a cloud in db by name
func (d *DataStore) FindCloudByName(name string) (*cloud.Options, error) {
	c := &cloud.Options{}
	col := d.s.DB(defaultDBName).C(cloudCollection)
	err := col.Find(bson.M{"name": name}).One(c)
	return c, err
}

// DeleteCloudByName delete a cloud in db by name
func (d *DataStore) DeleteCloudByName(name string) error {
	col := d.s.DB(defaultDBName).C(cloudCollection)
	err := col.Remove(bson.M{"name": name})
	return err
}
