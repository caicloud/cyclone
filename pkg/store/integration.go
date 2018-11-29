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
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/cyclone/pkg/api"
	"time"
)

// InsertIntegration insert a new integration document to db
func (d *DataStore) InsertIntegration(i *api.Integration) error {
	i.ID = bson.NewObjectId().Hex()

	i.CreationTime = time.Now()
	i.LastUpdateTime = time.Now()
	return d.integrationCollection.Insert(i)
}

// FindAllIntegrations returns all integrations
func (d *DataStore) FindAllIntegrations() ([]api.Integration, error) {
	is := []api.Integration{}

	err := d.integrationCollection.Find(bson.M{}).All(&is)
	return is, err
}

// GetIntegration returns a integration in db by name
func (d *DataStore) GetIntegration(name string) (*api.Integration, error) {
	i := &api.Integration{}

	err := d.integrationCollection.Find(bson.M{"name": name}).One(i)
	return i, err
}

// DeleteIntegrationByName delete a integration in db by name
func (d *DataStore) DeleteIntegrationByName(name string) error {
	return d.integrationCollection.Remove(bson.M{"name": name})
}

// InsertIntegration updates the integration,
// please make sure the integration id is provided before call this method.
func (d *DataStore) UpdateIntegration(i *api.Integration) error {
	i.LastUpdateTime = time.Now()
	return d.integrationCollection.UpdateId(i.ID, i)
}
