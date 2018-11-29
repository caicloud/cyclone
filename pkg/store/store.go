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

	"github.com/caicloud/cyclone/pkg/wait"
)

const (
	defaultDBName                string = "cyclone"
	cloudCollectionName          string = "clouds"
	projectCollectionName        string = "projects"
	pipelineCollectionName       string = "pipelines"
	pipelineRecordCollectionName string = "pipelineRecords"
	eventCollectionName          string = "events"
	integrationCollectionName    string = "integrations"

	socketTimeout  = time.Second * 5
	syncTimeout    = time.Second * 5
	tickerDuration = time.Second * 5
)

var (
	session *mgo.Session
	saltKey string
	mclosed chan struct{}
)

// DataStore is the type for mongo db store.
type DataStore struct {
	s       *mgo.Session
	saltKey string

	// Collections
	cloudCollection          *mgo.Collection
	projectCollection        *mgo.Collection
	pipelineCollection       *mgo.Collection
	pipelineRecordCollection *mgo.Collection
	eventCollection          *mgo.Collection
	integrationCollection    *mgo.Collection
}

// Init store mongo client session
func Init(host string, gracePeriod time.Duration, closing chan struct{}, key string) (chan struct{}, error) {
	saltKey = key
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
		return nil, err
	}

	log.Infof("connect to mongodb addr: %s", host)
	// Set the session mode as Eventual to ensure that the socket is created for each request.
	// Can switch to other mode only after the old APIs are cleaned up.
	session.SetMode(mgo.Eventual, true)

	go backgroundMongo(closing)

	err = ensureIndexes()
	if err != nil {
		log.Errorf("Fail to create indexes as %v", err)
		return nil, err
	}

	return mclosed, nil
}

// ensureIndexes ensures the indexes for each collection.
func ensureIndexes() error {
	projectCollection := session.DB(defaultDBName).C(projectCollectionName)
	projectIndex := mgo.Index{Key: []string{"name"}, Unique: true}
	err := projectCollection.EnsureIndex(projectIndex)
	if err != nil {
		log.Errorf("fail to create index for project as %v", err)
		return err
	}

	pipelineCollection := session.DB(defaultDBName).C(pipelineCollectionName)
	pipelineIndex := mgo.Index{Key: []string{"name", "projectID"}, Unique: true}
	if err = pipelineCollection.EnsureIndex(pipelineIndex); err != nil {
		log.Errorf("fail to create index for pipeline as %v", err)
		return err
	}

	cloudCollection := session.DB(defaultDBName).C(cloudCollectionName)
	cloudIndex := mgo.Index{Key: []string{"name"}, Unique: true}
	if err = cloudCollection.EnsureIndex(cloudIndex); err != nil {
		log.Errorf("fail to create index for cloud as %v", err)
		return err
	}

	integrationCollection := session.DB(defaultDBName).C(integrationCollectionName)
	integrationIndex := mgo.Index{Key: []string{"name"}, Unique: true}
	if err = integrationCollection.EnsureIndex(integrationIndex); err != nil {
		log.Errorf("fail to create index for integration as %v", err)
		return err
	}

	return nil
}

// NewStore copy a mongo client session
func NewStore() *DataStore {
	s := session.Copy()
	return &DataStore{
		s:                        s,
		saltKey:                  saltKey,
		cloudCollection:          session.DB(defaultDBName).C(cloudCollectionName),
		projectCollection:        session.DB(defaultDBName).C(projectCollectionName),
		pipelineCollection:       session.DB(defaultDBName).C(pipelineCollectionName),
		pipelineRecordCollection: session.DB(defaultDBName).C(pipelineRecordCollectionName),
		eventCollection:          session.DB(defaultDBName).C(eventCollectionName),
		integrationCollection:    session.DB(defaultDBName).C(integrationCollectionName),
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
