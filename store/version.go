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

// FindVersionsByCondition finds a version entity by service ID and version name.
func (d *DataStore) FindVersionsByCondition(serviceID, versionname string) ([]api.Version, error) {
	versions := []api.Version{}
	filter := bson.M{"service_id": serviceID, "name": versionname}
	col := d.s.DB(defaultDBName).C(versionCollectionName)
	err := col.Find(filter).Iter().All(&versions)
	return versions, err
}

// NewVersionDocument creates a new document (record) in mongodb. It returns version
// id of the newly created version.
func (d *DataStore) NewVersionDocument(version *api.Version) (string, error) {
	version.VersionID = bson.NewObjectId().Hex()
	col := d.s.DB(defaultDBName).C(versionCollectionName)
	_, err := col.Upsert(bson.M{"_id": version.VersionID}, version)
	return version.VersionID, err
}

// UpdateVersionDocument updates a version entirely.
func (d *DataStore) UpdateVersionDocument(versionID string, version api.Version) error {
	filter := bson.M{"_id": versionID}
	change := mgo.Change{
		Update: bson.M{"$set": version},
	}
	col := d.s.DB(defaultDBName).C(versionCollectionName)
	_, err := col.Find(filter).Apply(change, &version)
	return err
}

// FindVersionByID finds a version entity by ID.
func (d *DataStore) FindVersionByID(versionID string) (*api.Version, error) {
	version := &api.Version{}
	col := d.s.DB(defaultDBName).C(versionCollectionName)
	err := col.Find(bson.M{"_id": versionID}).One(version)
	return version, err
}

// FindVersionsByServiceID finds a version entity by service ID.
func (d *DataStore) FindVersionsByServiceID(serviceID string) ([]api.Version, error) {
	versions := []api.Version{}
	filter := bson.M{"service_id": serviceID}
	col := d.s.DB(defaultDBName).C(versionCollectionName)
	err := col.Find(filter).Sort("-create_time").Iter().All(&versions)
	return versions, err
}

// FindVersionsWithPaginationByServiceID finds a page of versions by service ID.
func (d *DataStore) FindVersionsWithPaginationByServiceID(serviceID string, filter map[string]interface{}, start, limit int) ([]api.Version, int, error) {
	versions := []api.Version{}
	if filter == nil {
		filter = bson.M{"service_id": serviceID}
	} else {
		filter["service_id"] = serviceID
	}

	col := d.s.DB(defaultDBName).C(versionCollectionName)
	query :=  col.Find(filter)
	total, err := query.Count()
	if err != nil {
		return versions, 0, err
	}

	// If there is no limit, return all.
	if limit != 0 {
		query.Limit(limit)
	}

	if err = query.Skip(start).Sort("-create_time").All(&versions); err != nil {
		return versions, 0, err
	}

	return versions, total, err
}

// FindRecentVersionsByServiceID finds a set of versions with conditions by service ID.
func (d *DataStore) FindRecentVersionsByServiceID(serviceID string, filter map[string]interface{}, limit int) ([]api.Version, int, error) {
	versions := []api.Version{}
	if filter == nil {
		filter = bson.M{"service_id": serviceID}
	} else {
		filter["service_id"] = serviceID
	}

	col := d.s.DB(defaultDBName).C(versionCollectionName)
	query :=  col.Find(filter)
	total, err := query.Count()
	if err != nil {
		return versions, 0, err
	}

	// If there is no limit, return all.
	if limit != 0 {
		query.Limit(limit)
	}

	if err = query.Sort("-create_time").All(&versions); err != nil {
		return versions, 0, err
	}

	return versions, total, err
}

// FindLatestVersionByServiceID finds the latest version entity by service ID.
func (d *DataStore) FindLatestVersionByServiceID(serviceID string) (*api.Version, error) {
	version := &api.Version{}
	filter := bson.M{"service_id": serviceID}
	col := d.s.DB(defaultDBName).C(versionCollectionName)
	err := col.Find(filter).Sort("-create_time").Limit(1).One(version)
	return version, err
}

// DeleteVersionByID removes version by versionID.
func (d *DataStore) DeleteVersionByID(versionID string) error {
	col := d.s.DB(defaultDBName).C(versionCollectionName)
	err := col.Remove(bson.M{"_id": versionID})
	return err
}
