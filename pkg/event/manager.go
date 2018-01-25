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

package event

import (
	"time"

	log "github.com/golang/glog"
	"github.com/zoumo/logdog"
	mgo "gopkg.in/mgo.v2"
	"k8s.io/client-go/rest"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/remote"
	"github.com/caicloud/cyclone/store"
)

var (
	CloudController *cloud.Controller

	workerOptions *cloud.WorkerOptions

	// remote api manager
	remoteManager *remote.Manager
)

// maxRetry represents the max number of retry for event when cloud is busy.
// TODO(robin) Make this configurable.
const maxRetry = 60

// Init init event manager
// Step1: init cloud controller
// Step2: init event operation map
// Step3: new event manager
// Step4: create a goroutine to watch events
// Step5: new a remote api manager
func Init(wopts *cloud.WorkerOptions, cloudAutoDiscovery bool) {

	initCloudController(wopts, cloudAutoDiscovery)

	initOperationMap()

	ds := store.NewStore()

	em := NewEventManager(ds)
	go em.WatchEvent()

	remoteManager = remote.NewManager()
}

// FIXME, so ugly
// load clouds from database
func initCloudController(wopts *cloud.WorkerOptions, cloudAutoDiscovery bool) {
	CloudController = cloud.NewController()
	// load clouds from store
	ds := store.NewStore()
	defer ds.Close()
	clouds, err := ds.FindAllClouds()
	if err != nil {
		logdog.Error("Can not find clouds from ds", logdog.Fields{"err": err})
		return
	}

	CloudController.AddClouds(clouds...)

	if len(CloudController.Clouds) == 0 && cloudAutoDiscovery {
		addInClusterK8SCloud()
	}

	workerOptions = wopts
}

func addInClusterK8SCloud() {
	ds := store.NewStore()
	defer ds.Close()
	_, err := rest.InClusterConfig()
	if err == nil {
		// in k8s cluster
		opt := cloud.Options{
			Kind:         cloud.KindK8SCloud,
			Name:         "_inCluster",
			K8SInCluster: true,
		}
		err := CloudController.AddClouds(opt)
		if err != nil {
			logdog.Warn("Can not add inCluster k8s cloud to database")
		}
		err = ds.InsertCloud(&opt)
		if err != nil {
			logdog.Warn("Can not add inCluster k8s cloud to database")
		}
	}
}

// EventManager represents the manager of events.
type EventManager interface {
	HandleEvent(event *api.Event) error
	WatchEvent()
}

type eventManager struct {
	ds *store.DataStore
}

// NewEventManager creates the event manager.
func NewEventManager(ds *store.DataStore) EventManager {
	return &eventManager{
		ds: ds,
	}
}

// HandleEvent handles the event.
func (em *eventManager) HandleEvent(event *api.Event) error {
	if err := mapOperation[event.Operation].Handler(event); err != nil {
		if cloud.IsAllCloudsBusyErr(err) && event.Retry < maxRetry {
			log.Info("All system worker are busy, wait for 10 seconds")
			event.Retry++
			em.ds.ResetEvent(event)
			return nil
		}

		event.Status = api.EventStatusFail
		event.ErrorMessage = err.Error()
		log.Errorf("handle event %s as err: %s", event.EventID, err.Error())
		postHookEvent(event)

		if err = em.ds.DeleteEvent(string(event.EventID)); err != nil {
			log.Errorf("fail to delete event %s as err: %s", event.EventID, err.Error())
		}
	} else {
		event.Status = api.EventStatusRunning
		event.QueueStatus = api.Handling
		if event.Operation == CreateVersionOps {
			event.Version.Status = api.VersionRunning
			if err := em.ds.UpdateVersionDocument(event.Version.VersionID, event.Version); err != nil {
				log.Errorf("Unable to update version status post hook for %+v: %v", event.Version, err)
			}
		}
		if err = em.ds.UpdateEvent(event); err != nil {
			return err
		}
	}

	return nil
}

// WatchEvent watches the event queue, and handles the events.
func (em *eventManager) WatchEvent() {
	for {
		time.Sleep(1 * time.Second)

		event, err := em.ds.NextEvent()
		if err != nil {
			if err != mgo.ErrNotFound {
				log.Errorf("fail to get the event as %s", err.Error())
			}
			continue
		}

		if err = em.HandleEvent(event); err != nil {
			log.Errorf("fail to handle the event as %s", err.Error())
		}
	}
}

// UpdateEvent updates the event. If it is finished, delete it and trigger the post hook.
func UpdateEvent(event *api.Event) error {
	ds := store.NewStore()
	defer ds.Close()

	if IsEventFinished(event) {
		log.Infof("Event %s has finished", event.EventID)
		if err := ds.DeleteEvent(string(event.EventID)); err != nil {
			log.Errorf("fail to delete the event %s", event.EventID)
			return err
		}
		postHookEvent(event)
	} else {
		if err := ds.UpdateEvent(event); err != nil {
			log.Errorf("fail to update the event %s", event.EventID)
			return err
		}
	}

	return nil
}

// DeleteEvent deletes the event. If it is running, delete its worker at the same time.
func DeleteEvent(id string) error {
	event, err := GetEvent(id)
	if err != nil {
		log.Errorf("fail to get the event %s as %s", event.EventID, err.Error())
		return err
	}
	ds := store.NewStore()
	defer ds.Close()

	// Delete the event in event queue.
	if err := ds.DeleteEvent(string(event.EventID)); err != nil {
		log.Errorf("fail to delete the event %s", event.EventID)
		return err
	}

	if event.QueueStatus == api.Handling {
		log.Infof("terminate the worker for handling event %s", event.EventID)
		terminateEventWorker(event.Worker)
	}

	return nil
}

// GetEvent get the event by ID.
func GetEvent(id string) (*api.Event, error) {
	ds := store.NewStore()
	defer ds.Close()

	return ds.GetEventByID(id)
}

// IsEventFinished return true if event is finished.
func IsEventFinished(event *api.Event) bool {
	if event.Status == api.EventStatusSuccess ||
		event.Status == api.EventStatusFail ||
		event.Status == api.EventStatusCancel {
		return true
	}
	return false
}

// CheckWorkerTimeout ...
func CheckWorkerTimeout(event *api.Event) {
	if IsEventFinished(event) {
		return
	}
	worker, err := CloudController.LoadWorker(event.Worker)
	if err != nil {
		log.Error("load worker error")
		return
	}

	ok, left := worker.IsTimeout()
	if ok {
		event.Status = api.EventStatusFail
		UpdateEvent(event)
		return
	}

	time.Sleep(left)

	ds := store.NewStore()
	defer ds.Close()
	event, err = ds.GetEventByID(string(event.EventID))
	if err != nil {
		return
	}

	if !IsEventFinished(event) {
		log.Infof("event time out: %v", event)
		event.Status = api.EventStatusCancel
		UpdateEvent(event)
	}
}

func terminateEventWorker(workerInfo cloud.WorkerInfo) {
	w, err := CloudController.LoadWorker(workerInfo)
	if err != nil {
		logdog.Warnf("load worker err: %v", err)
	} else {
		err = w.Terminate()
		if err != nil {
			logdog.Warnf("Terminate worker err: %v", err)
		}
	}
}
