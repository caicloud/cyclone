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

// NewVersionLogDocument creates a new document (record) in mongodb. It returns version
// log id of the newly created version.
func (d *DataStore) NewVersionLogDocument(versionLog *api.VersionLog) (string, error) {
	col := d.s.DB(defaultDBName).C(versionLogCollectionName)
	versionLog.LogID = uuid.NewV4().String()
	_, err := col.Upsert(bson.M{"_id": versionLog.LogID}, versionLog)
	return versionLog.LogID, err
}

// FindVersionLogByID finds a version log entity by ID.
func (d *DataStore) FindVersionLogByID(LogID string) (*api.VersionLog, error) {
	col := d.s.DB(defaultDBName).C(versionLogCollectionName)
	log := &api.VersionLog{}
	err := col.Find(bson.M{"_id": LogID}).One(log)
	return log, err
}

// FindVersionLogByVersionID finds a version log entity by version ID.
func (d *DataStore) FindVersionLogByVersionID(versionID string) (*api.VersionLog, error) {
	col := d.s.DB(defaultDBName).C(versionLogCollectionName)
	log := &api.VersionLog{}
	filter := bson.M{"version_id": versionID}
	err := col.Find(filter).One(log)
	return log, err
}

// UpdateVersionLogDocument updates a document (record) in mongodb.
func (d *DataStore) UpdateVersionLogDocument(versionLog *api.VersionLog) error {
	col := d.s.DB(defaultDBName).C(versionLogCollectionName)
	_, err := col.Upsert(bson.M{"_id": versionLog.LogID}, versionLog)
	return err
}
