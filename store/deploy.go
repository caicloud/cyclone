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
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2/bson"
)

// NewDeployDocument creates a new document (record) in mongodb. It returns deploy
// id of the newly created deploy.
func (d *DataStore) NewDeployDocument(deploy *api.Deploy) (string, error) {
	deploy.DeployID = uuid.NewV4().String()
	col := d.s.DB(defaultDBName).C(deployCollectionName)
	_, err := col.Upsert(bson.M{"_id": deploy.DeployID}, deploy)
	return deploy.DeployID, err
}

// FindDeployByID finds a deploy entity by ID.
func (d *DataStore) FindDeployByID(deployID string) (*api.Deploy, error) {
	deploy := &api.Deploy{}
	col := d.s.DB(defaultDBName).C(deployCollectionName)
	err := col.Find(bson.M{"_id": deployID}).One(deploy)
	return deploy, err
}

// UpsertDeployDocument upsert a special deploy document
func (d *DataStore) UpsertDeployDocument(deploy *api.Deploy) (string, error) {
	col := d.s.DB(defaultDBName).C(deployCollectionName)
	_, err := col.Upsert(bson.M{"_id": deploy.DeployID}, deploy)
	return deploy.DeployID, err
}
