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
	"container/list"
	"encoding/json"
	"sync"
	"time"

	"k8s.io/client-go/rest"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/remote"
	"github.com/caicloud/cyclone/store"
	"github.com/zoumo/logdog"
	log "github.com/zoumo/logdog"
	"golang.org/x/net/context"
)

var (
	CloudController *cloud.Controller

	workerOptions *cloud.WorkerOptions

	// remote api manager
	remoteManager *remote.Manager
)

// maxRetry represents the max number of retry for event when cloud is busy.
// TODO(robin) Make this configurable.
const maxRetry = 10

// Init init event manager
// Step1: init event operation map
// Step2: new a etcd client
// Step3: load unfinished events from etcd
// Step4: create a unfinished events watcher
// Step5: new a remote api manager
func Init(wopts *cloud.WorkerOptions, cloudAutoDiscovery bool) {

	initCloudController(wopts, cloudAutoDiscovery)

	initOperationMap()

	go watchMongoMQ()

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

func watchMongoMQ() {
	ds := store.NewStore()
	defer ds.Close()

	for {
		time.Sleep(time.Second)

		massage, err := ds.GetMassage()
		if err != nil {
			logdog.Errorf("watch mongoMQ err: %s", err.Error())
			continue
		}

		if massage == nil {
			logdog.Debugf("massage queue empty")
			continue
		}

		go handleJob(massage)
	}
}

func handleJob(massage *store.Massage) {
	for {
		event := massage.Event
		err := handleEvent(&event)
		if err != nil {
			// if err == resource.ErrUnableSupport {
			// 	log.Info("Waiting for resource to be relaesed...")
			// 	time.Sleep(time.Second * 10)
			// 	continue
			// }
			// worker busy
			// if err == ErrWorkerBusy {
			// 	log.Info("All system worker are busy, wait for 10 seconds")
			// 	time.Sleep(time.Second * 10)
			// 	continue
			// }
			ds := store.NewStore()
			defer ds.Close()

			if cloud.IsAllCloudsBusyErr(err) {
				// Pop out this event and execute the next event
				// Increase the retry of this event and push it back into the queue.
				event.Retry++
				if event.Retry > maxRetry {
					event.Retry = 0
					ds.ResetMassage(massage)
					return
				}

				log.Info("All system worker are busy, wait for 10 seconds")
				time.Sleep(time.Second * 10)
				continue
			}

			// remove the event from queue which had run
			ds.RemoveMassage(massage.ID)

			event.Status = api.EventStatusFail
			event.ErrorMessage = err.Error()
			log.Error("handle event err", log.Fields{"error": err, "event": event})
			postHookEvent(&event)
			continue
		}

		ds := store.NewStore()
		defer ds.Close()

		// remove the event from queue which had run
		ds.RemoveMassage(massage.ID)

		event.Status = api.EventStatusRunning

		if event.Operation == "create-version" {
			event.Version.Status = api.VersionRunning
			if err := ds.UpdateVersionDocument(event.Version.VersionID, event.Version); err != nil {
				log.Errorf("Unable to update version status post hook for %+v: %v", event.Version, err)
			}
		}
	}
}

// IsEventFinished return event is finished
func IsEventFinished(event *api.Event) bool {
	if event.Status == api.EventStatusSuccess ||
		event.Status == api.EventStatusFail ||
		event.Status == api.EventStatusCancel {
		return true
	}
	return false
}

// LoadEventFromPipelineRecord loads event info from pipelineRecord.  || TODO
func LoadEventFromPipelineRecord(id string) (*api.Event, error) {
	return event, nil
}

// UpdateEventToPipelineRecord updates the pipelineRecord based on event    || TODO 
func UpdateEventToPipelineRecord(event *api.Event) error {
	return nil
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
		UpdateEventToPipelineRecord(event)
		return
	}

	time.Sleep(left)

	event, err = LoadEventFromPipelineRecord(event.EventID)
	if err != nil {
		return
	}

	if !IsEventFinished(event) {
		log.Infof("event time out: %v", event)
		event.Status = api.EventStatusCancel
		UpdateEventToPipelineRecord(event)
	}
}
