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

	log "github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"k8s.io/client-go/pkg/util/wait"
)

const (
	defaultDBName                string = "cyclone"
	serviceCollectionName        string = "ServiceCollection"
	versionCollectionName        string = "VersionCollection"
	versionLogCollectionName     string = "VersionLogCollection"
	recourceCollectionName       string = "RecourceCollection"
	projectVersionCollectionName string = "ProjectVersionCollection"
	daemonCollectionName         string = "DaemonCollection"
	remoteCollectionName         string = "RemoteCollection"
	workerNodeCollection         string = "WorkerNodeCollection"
	deployCollectionName         string = "DeployCollection"
	ResourceCollectionName       string = "ResourceCollection"
	CloudsCollection             string = "clouds"

	projectCollectionName        string = "projects"
	pipelineCollectionName       string = "pipelines"
	pipelineRecordCollectionName string = "pipelineRecords"

	socketTimeout  = time.Second * 5
	syncTimeout    = time.Second * 5
	tickerDuration = time.Second * 5
)

var (
	session *mgo.Session
	mclosed chan struct{}
)

// DataStore is the type for mongo db store.
type DataStore struct {
	s *mgo.Session

	// Collections
	projectCollection        *mgo.Collection
	pipelineCollection       *mgo.Collection
	pipelineRecordCollection *mgo.Collection
}

// Init store mongo client session
func Init(host string, gracePeriod time.Duration, closing chan struct{}) (*mgo.Session, chan struct{}, error) {
	mclosed = make(chan struct{})
	var err error

	// dail mongo session
	// wait mongodb set up
	wait.Poll(time.Second, gracePeriod, func() (bool, error) {
		session, err = mgo.Dial(host)
		return err == nil, nil
	})

	if err != nil {
		log.Errorf("Unable connect to mongodb addr %s", host)
		return nil, nil, err
	}

	log.Infof("connect to mongodb addr: %s", host)
	session.SetMode(mgo.Strong, true)

	go backgroundMongo(closing)

	return session, mclosed, nil
}

// NewStore copy a mongo client session
func NewStore() *DataStore {
	s := session.Copy()

	// s is for old api, it will be closed after each use.
	// The new collections are for new api, they will be reused.
	return &DataStore{
		s:                        s,
		projectCollection:        session.DB(defaultDBName).C(projectCollectionName),
		pipelineCollection:       session.DB(defaultDBName).C(pipelineCollectionName),
		pipelineRecordCollection: session.DB(defaultDBName).C(pipelineRecordCollectionName),
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

// Background goroutine for mongo. It can hold mongo connection & close session when progress exit.
func backgroundMongo(closing chan struct{}) {
	ticker := time.NewTicker(tickerDuration)
	for {
		select {
		case <-ticker.C:
			if err := session.Ping(); err != nil {
				log.Errorf("Ping Mongodb with error %s", err.Error())
				session.Refresh()
				session.SetSocketTimeout(socketTimeout)
				session.SetSyncTimeout(syncTimeout)
			}
		case <-closing:
			session.Close()
			log.Info("Mongodb session has been closed")
			close(mclosed)
			return
		}
	}
}
