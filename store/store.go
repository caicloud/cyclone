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
	"gopkg.in/mgo.v2"
)

const (
	defaultDBName                string = "cyclone"
	serviceCollectionName        string = "ServiceCollection"
	versionCollectionName        string = "VersionCollection"
	versionLogCollectionName     string = "VersionLogCollection"
	recourceCollectionName       string = "RecourceCollection"
	projectCollectionName        string = "ProjectCollection"
	projectVersionCollectionName string = "ProjectVersionCollection"
	daemonCollectionName         string = "DaemonCollection"
	remoteCollectionName         string = "RemoteCollection"
	workerNodeCollection         string = "WorkerNodeCollection"
	deployCollectionName         string = "DeployCollectionName"
	ResourceCollectionName       string = "ResourceCollection"
)

var (
	session *mgo.Session
)

// DataStore is the type for mongo db store.
type DataStore struct {
	s *mgo.Session
}

// Init store mongo client session
func Init(s *mgo.Session) {
	session = s
}

// NewStore copy a mongo client session
func NewStore() *DataStore {
	return &DataStore{
		s: session.Copy(),
	}
}

// Close close mongo client session
func (d *DataStore) Close() {
	d.s.Close()
}

// Ping ping mongo server
func (d *DataStore) Ping() error {
	return d.s.Ping()
}
