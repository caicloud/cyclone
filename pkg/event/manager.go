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

	"github.com/caicloud/cyclone/cmd/worker/options"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/cloud"
	"github.com/caicloud/cyclone/pkg/store"
)

// maxRetry represents the max number of retry for event when cloud is busy.
// TODO(robin) Make this configurable.
const maxRetry = 60

var (
	defaultWorkerOptions *options.WorkerOptions
)

// Init init event manager
// Step1: new event manager
// Step2: create a goroutine to watch events
func Init(opts *options.WorkerOptions) {
	ds := store.NewStore()
	em := NewEventManager(ds)

	defaultWorkerOptions = opts

	go em.WatchEvent()
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
	err := createWorkerForEvent(event)
	if err != nil {
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
	workerCfg := event.Project.Worker
	buildInfo := event.Pipeline.Build.BuildInfo
	performParams := event.PipelineRecord.PerformParams
	workerInfo := &api.WorkerInfo{
		Name: "cyclone-worker-" + event.ID,
	}

	if workerCfg != nil {
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
					workerInfo.MountPath = "/root/.m2"
				case api.NPMBuildTool:
					workerInfo.MountPath = "/root/.npm"
				default:
					// Just log error and let the pipeline to run in non-cache mode.
					return fmt.Errorf("Build tool %s is not supported, only supports: %s, %s", tool, api.MavenBuildTool, api.NPMBuildTool)
				}

				workerInfo.CacheVolume = cache.Name
			}
		}
	}

	cp, err := getCloudProvider(workerCfg)
	if err != nil {
		log.Error(err)
		return err
	}

	opts := defaultWorkerOptions
	opts.EventID = event.ID
	workerInfo, err = cp.Provision(workerInfo, opts)
	if err != nil {
		return err
	}

	event.WorkerInfo = workerInfo

	// err = worker.Do()
	// if err != nil {
	// 	return err
	// }

	// update worker info to event
	// event.Worker = worker.GetWorkerInfo()

	// log.Infof("Worker info: %v", event.Worker)

	// trigger after get an valid worker
	// triggerHooks(event, PostStartPhase)

	// No need to call UpdateEvent() here, as the callers will update the event status after call createWorkerForEvent().
	// UpdateEvent(event)
	go CheckWorkerTimeout(event)

	return nil
}

func postHookEvent(event *api.Event) {
	log.Info("posthook of event")
	log.Infof("worker： %s", event.Project.Worker)
	terminateEventWorker(event)
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
		terminateEventWorker(event)
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
	// worker, err := CloudController.LoadWorker(event.Worker)
	// if err != nil {
	// 	log.Errorf("load worker %v with error: %v", event.Worker, err)
	// 	return
	// }

	pipelineRecord := event.PipelineRecord
	now := time.Now()
	workerInfo := event.WorkerInfo
	if now.After(workerInfo.DueTime) {
		pipelineRecord.Status = api.Failed
		err := UpdateEvent(event)
		if err != nil {
			log.Errorf("update event %s error: %v", event.ID, err)
		}
		return
	}

	time.Sleep(workerInfo.DueTime.Sub(now))

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

func terminateEventWorker(event *api.Event) {
	cp, err := getCloudProvider(event.Project.Worker)
	if err != nil {
		logdog.Warnf("load worker %v with error: %v", event.WorkerInfo, err)
	} else {
		if event.WorkerInfo != nil {
			err = cp.TerminateWorker(event.WorkerInfo.Name)
			if err != nil {
				logdog.Warnf("Terminate worker err: %v", err)
			}
		}
	}
}

func getCloudProvider(w *api.Worker) (cloud.Provider, error) {

	log.Infof("worker: %+v", w)

	ds := store.NewStore()
	defer ds.Close()

	// Use the incluster cloud by default.
	var clusterName string
	// if w != nil && w.CloudOptions != nil {
	// 	clusterName = w.CloudOptions.CloudName
	if w != nil && w.Location != nil {
		clusterName = w.Location.ClusterName
	} else {
		clusterName = "inCluster"
		log.Warningf("No cluster specified for this project, will use the defult cluster %s", clusterName)
	}

	c, err := ds.FindCloudByName(clusterName)
	if err != nil {
		err = fmt.Errorf("fail to find cloud %s as %v", clusterName, err)
		log.Error(err)
		return nil, err
	}

	return cloud.NewCloudProvider(c)
}
