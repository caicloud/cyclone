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
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// FindServiceByCondition finds a list of services via user ID and serice name.
func (d *DataStore) FindServiceByCondition(userID, servicename string) ([]api.Service, error) {
	services := []api.Service{}
	filter := bson.M{"user_id": userID, "name": servicename}
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	err := col.Find(filter).Iter().All(&services)
	return services, err
}

// NewServiceDocument creates a new document (record) in mongodb. It returns service
// id of the newly created service.
func (d *DataStore) NewServiceDocument(service *api.Service) (string, error) {
	service.ServiceID = uuid.NewV4().String()
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	_, err := col.Upsert(bson.M{"_id": service.ServiceID}, service)
	return service.ServiceID, err
}

// UpdateRepositoryStatus updates service repository status.
func (d *DataStore) UpdateRepositoryStatus(serviceID string, status api.RepositoryStatus) error {
	filter := bson.M{"_id": serviceID}
	change := mgo.Change{
		Update: bson.M{"$set": bson.M{"repository.status": status}},
	}

	service := api.Service{}
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	_, err := col.Find(filter).Apply(change, &service)
	return err
}

// FindServicesByUserID finds a list of services via user ID.
func (d *DataStore) FindServicesByUserID(userID string) ([]api.Service, error) {
	services := []api.Service{}
	filter := bson.M{"user_id": userID}
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	err := col.Find(filter).Iter().All(&services)
	return services, err
}

// FindServiceByID finds a service entity by ID.
func (d *DataStore) FindServiceByID(serviceID string) (*api.Service, error) {
	service := &api.Service{}
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	err := col.Find(bson.M{"_id": serviceID}).One(service)
	return service, err
}

// DeleteServiceByID removes service by service_id.
func (d *DataStore) DeleteServiceByID(serviceID string) error {
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	err := col.Remove(bson.M{"_id": serviceID})
	return err
}

// AddNewVersion adds a new success version (version ID) to a given service.
func (d *DataStore) AddNewVersion(serviceID string, versionID string) error {
	change := mgo.Change{
		Update: bson.M{"$push": bson.M{"versions": versionID}},
	}
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	_, err := col.Find(bson.M{"_id": serviceID}).Apply(change, nil)
	return err
}

// AddNewFailVersion adds a new fail version (version ID) to a given service.
func (d *DataStore) AddNewFailVersion(serviceID string, versionID string) error {
	change := mgo.Change{
		Update: bson.M{"$push": bson.M{"version_fails": versionID}},
	}
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	_, err := col.Find(bson.M{"_id": serviceID}).Apply(change, nil)
	return err
}

// UpdateServiceLastInfo updates service's lastCreateTIme and lastVersionName.
func (d *DataStore) UpdateServiceLastInfo(serviceID string, lasttime time.Time, lastname string) error {
	filter := bson.M{"_id": serviceID}
	change := mgo.Change{
		Update: bson.M{"$set": bson.M{"last_createtime": lasttime, "last_versionname": lastname}},
	}
	service := api.Service{}
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	_, err := col.Find(filter).Apply(change, &service)
	return err
}

// UpsertServiceDocument upsert a special serivce document
func (d *DataStore) UpsertServiceDocument(service *api.Service) (string, error) {
	col := d.s.DB(defaultDBName).C(serviceCollectionName)
	_, err := col.Upsert(bson.M{"_id": service.ServiceID}, service)
	return service.ServiceID, err
}
