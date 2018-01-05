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

		go handleMassage(massage)
	}
}

func handleMassage(massage *store.Massage) {
	for {
		job := massage.Job
		ds := store.NewStore()
		defer ds.Close()

		err := createPipelineRecord(job)
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

			if cloud.IsAllCloudsBusyErr(err) {
				job.Retry++
				if job.Retry > maxRetry {
					job.Retry = 0
					ds.ResetMassage(massage)
					return
				}

				log.Info("All system worker are busy, wait for 10 seconds")
				time.Sleep(time.Second * 10)
				continue
			}

			// remove the job from queue which had run
			ds.RemoveMassage(massage.ID)

			job.PipelineRecord.Status = api.EventStatusFail
			job.PipelineRecord.ErrorMessage = err.Error()
			log.Error("handle job err", log.Fields{"error": err, "job": job})
			postHookJob(&job)
			continue
		}

		// remove the job from queue which had run
		ds.RemoveMassage(massage.ID)

		job.PipelineRecord.Status = api.EventStatusRunning

		if err := ds.UpdateJobToPipelineRecord(job.PipelineRecord.ID, job.PipelineRecord); err != nil {
			log.Errorf("Unable to update version status post hook for %+v: %v", job.PipelineRecord, err)
		}

	}
}

// IsJobFinished return job is finished
func IsJobFinished(job *store.Job) bool {
	if job.PipelineRecord.Status == api.EventStatusSuccess ||
		job.PipelineRecord.Status == api.EventStatusFail ||
		job.PipelineRecord.Status == api.EventStatusCancel {
		return true
	}
	return false
}

// LoadJobFromPipelineRecord loads job info from pipelineRecord.  || TODO
func LoadJobFromPipelineRecord(id string) (*store.Job, error) {
	return nil, nil
}

// UpdateJobToPipelineRecord updates the pipelineRecord based on job    || TODO
func UpdateJobToPipelineRecord(job *store.Job) error {
	return nil
}

// CheckWorkerTimeout ...
func CheckWorkerTimeout(job *store.Job) {
	if IsJobFinished(job) {
		return
	}
	worker, err := CloudController.LoadWorker(job.PipelineRecord.Worker)
	if err != nil {
		log.Error("load worker error")
		return
	}

	ok, left := worker.IsTimeout()
	if ok {
		job.PipelineRecord.Status = api.EventStatusFail
		UpdateJobToPipelineRecord(job)
		return
	}

	time.Sleep(left)

	job, err = LoadJobFromPipelineRecord(job.PipelineRecord.ID)
	if err != nil {
		return
	}

	if !IsJobFinished(job) {
		log.Infof("job time out: %v", job)
		job.PipelineRecord.Status = api.EventStatusCancel
		UpdateJobToPipelineRecord(job)
	}
}
