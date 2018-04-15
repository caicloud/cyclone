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
	"fmt"
	"time"

	log "github.com/golang/glog"
	"github.com/zoumo/logdog"
	mgo "gopkg.in/mgo.v2"
	"k8s.io/client-go/rest"

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/store"
)

var (
	CloudController *cloud.Controller

	workerOptions *cloud.WorkerOptions
)

// maxRetry represents the max number of retry for event when cloud is busy.
// TODO(robin) Make this configurable.
const maxRetry = 60

// Init init event manager
// Step1: init cloud controller
// Step2: new event manager
// Step3: create a goroutine to watch events
func Init(wopts *cloud.WorkerOptions, cloudAutoDiscovery bool) {
	initCloudController(wopts, cloudAutoDiscovery)
	ds := store.NewStore()

	em := NewEventManager(ds)
	go em.WatchEvent()
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
	eventID := event.ID
	pipelineRecord := event.PipelineRecord
	if err := createWorkerForEvent(event); err != nil {
		if cloud.IsAllCloudsBusyErr(err) && event.Retry < maxRetry {
			log.Info("All system worker are busy, wait for 10 seconds")
			event.Retry++
			em.ds.ResetEvent(event)
			return nil
		}

		pipelineRecord.Status = api.Failed
		pipelineRecord.ErrorMessage = err.Error()
		log.Errorf("handle event %s as err: %s", eventID, err.Error())
		postHookEvent(event)

		if err = em.ds.DeleteEvent(eventID); err != nil {
			log.Errorf("fail to delete event %s as err: %s", eventID, err.Error())
		}
	} else {
		event.QueueStatus = api.Handling
		pipelineRecord.Status = api.Running
		pipelineRecord.StartTime = time.Now()
		if err = em.ds.UpdateEvent(event); err != nil {
			log.Errorf("fail to update event %s as err: %s", eventID, err.Error())
			return err
		}
	}

	if err := em.ds.UpdatePipelineRecord(pipelineRecord); err != nil {
		log.Errorf("fail to update the pipeline record %s: %v", event.ID, err)
		return err
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

		go em.HandleEvent(event)
	}
}

func createWorkerForEvent(event *api.Event) error {
	log.Infof("Create worker for event %s", event.ID)
	opts := workerOptions.DeepCopy()
	workerCfg := event.Project.Worker
	buildInfo := event.Pipeline.Build.BuildInfo
	performParams := event.PipelineRecord.PerformParams
	if workerCfg != nil {
		if len(workerCfg.Namespace) != 0 {
			opts.Namespace = workerCfg.Namespace
		}

		// Only use cache when:
		// * the project has enable caches
		// * the pipeline has enable cache and set the build tool
		// * the perform params has enable cache
		// * the project has enable the cache for the build tool of the pipeline
		if buildInfo != nil && buildInfo.CacheDependency && buildInfo.BuildTool != nil && performParams.CacheDependency {
			tool := buildInfo.BuildTool.Name
			if cache, ok := workerCfg.DependencyCaches[tool]; ok {
				switch tool {
				case api.MavenBuildTool:
					opts.MountPath = "/root/.m2"
				case api.NPMBuildTool:
					opts.MountPath = "/root/.npm"
				default:
					// Just log error and let the pipeline to run in non-cache mode.
					return fmt.Errorf("Build tool %s is not supported, only supports: %s, %s", tool, api.MavenBuildTool, api.NPMBuildTool)
				}

				opts.CacheVolume = cache.Name
			}
		}
	}

	worker, err := CloudController.Provision(event.ID, opts)
	if err != nil {
		return err
	}

	// set worker info to event
	event.Worker = worker.GetWorkerInfo()

	err = worker.Do()
	if err != nil {
		return err
	}

	// update worker info to event
	event.Worker = worker.GetWorkerInfo()

	// trigger after get an valid worker
	// triggerHooks(event, PostStartPhase)

	// No need to call UpdateEvent() here, as the callers will update the event status after call createWorkerForEvent().
	// UpdateEvent(event)
	go CheckWorkerTimeout(event)

	return nil
}

func postHookEvent(event *api.Event) {
	log.Info("posthook of event")
	terminateEventWorker(event.Worker)
}

// UpdateEvent updates the event. If it is finished, delete it and trigger the post hook.
//func (em *eventManager) UpdateEvent(event *api.Event) error {
func UpdateEvent(event *api.Event) error {
	eventID := event.ID
	ds := store.NewStore()
	defer ds.Close()

	if IsEventFinished(event) {
		log.Infof("Event %s has finished", eventID)
		if err := ds.DeleteEvent(string(eventID)); err != nil {
			log.Errorf("fail to delete the event %s", eventID)
			return err
		}
		event.PipelineRecord.EndTime = time.Now()
		postHookEvent(event)
	} else {
		if err := ds.UpdateEvent(event); err != nil {
			log.Errorf("fail to update the event %s", eventID)
			return err
		}
	}

	err := ds.UpdatePipelineRecord(event.PipelineRecord)
	if err != nil {
		log.Errorf("Fail to update the pipeline record %s", event.PipelineRecord.ID)
		return err
	}

	return nil
}

// IsEventFinished return true if event is finished.
func IsEventFinished(event *api.Event) bool {
	status := event.PipelineRecord.Status
	if status == api.Success ||
		status == api.Failed ||
		status == api.Aborted {
		return true
	}
	return false
}

// DeleteEvent deletes the event. If it is running, delete its worker at the same time.
func DeleteEvent(id string) error {

	event, err := GetEvent(id)
	if err != nil {
		log.Errorf("fail to get the event %s as %s", id, err.Error())
		return err
	}
	ds := store.NewStore()
	defer ds.Close()

	// Delete the event in event queue.
	if err := ds.DeleteEvent(id); err != nil {
		log.Errorf("fail to delete the event %s", id)
		return err
	}

	if event.QueueStatus == api.Handling {
		log.Infof("terminate the worker for handling event %s", id)
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

	pipelineRecord := event.PipelineRecord
	ok, left := worker.IsTimeout()
	if ok {

		pipelineRecord.Status = api.Failed
		err = UpdateEvent(event)
		if err != nil {
			log.Errorf("update event %s error: %v", event.ID, err)
		}
		return
	}

	time.Sleep(left)

	ds := store.NewStore()
	defer ds.Close()
	e, err := ds.GetEventByID(event.ID)
	if err != nil {
		return
	}

	if !IsEventFinished(e) {
		e.PipelineRecord.Status = api.Aborted
		err = UpdateEvent(e)
		if err != nil {
			log.Errorf("update event %s error: %v", e.ID, err)
		}
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
