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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/notify"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
	"github.com/caicloud/cyclone/websocket"
	"github.com/zoumo/logdog"
	"golang.org/x/oauth2"
	"gopkg.in/mgo.v2/bson"
)

const (
	// PostStartPhase hooks phase
	PostStartPhase = "postStart"
	// PreStopPhase hooks phase
	PreStopPhase = "preStop"
)

// postHookJob is the job finished post hook.
func postHookJob(job *store.Job) {
	createPipelineRecordPostHook(job)
	w, err := CloudController.LoadWorker(job.PipelineRecord.Worker)
	if err != nil {
		logdog.Warnf("load worker err: %v", err)
	} else {
		err = w.Terminate()
		if err != nil {
			logdog.Warnf("Terminate worker err: %v", err)
		}
	}
}

// createPipelineRecord is the create version handler.
func createPipelineRecord(job *store.Job) error {
	logdog.Infof("create version handler")

	opts := workerOptions.DeepCopy()
	opts.Quota = Resource2Quota(job.PipelineRecord.BuildResource, opts.Quota)
	opts.Namespace = getNamespaceFromJob(job)
	worker, err := CloudController.Provision(string(job.PipelineRecord.ID), opts)
	if err != nil {
		return err
	}

	// set worker info to pipelineRecord
	job.PipelineRecord.Worker = worker.GetWorkerInfo()

	err = worker.Do()
	if err != nil {
		return err
	}

	ds := store.NewStore()
	defer ds.Close()

	ds.RemoveMassage(job.PipelineRecord.ID)

	// update worker info to pipelineRecord
	job.PipelineRecord.Worker = worker.GetWorkerInfo()

	// trigger after get an valid worker
	triggerHooks(job, PostStartPhase)

	go CheckWorkerTimeout(job)

	webhook := job.Pipeline.Repository.Webhook
	if remote, err := remoteManager.FindRemote(webhook); webhook == api.GITHUB && err == nil {
		if err = remote.PostCommitStatus(job.Pipeline, job.PipelineRecord); err != nil {
			logdog.Errorf("Unable to post commit status to github: %v", err)
		}
	}
	return nil
}

// createPipelineRecordPostHook is the create version post hook.
func createPipelineRecordPostHook(job *store.Job) {
	logdog.Infof("create version post hook")
	if job.Status == api.EventStatusSuccess {
		job.PipelineRecord.Status = api.VersionHealthy
	} else if job.Status == api.EventStatusCancel {
		job.PipelineRecord.Status = api.VersionCancel
	} else {
		job.PipelineRecord.Status = api.VersionFailed
	}
	job.PipelineRecord.EndTime = time.Now()

	operation := string(job.PipelineRecord.Operation)
	// Record that whether this job is a deploy for project. According this flag, we will make some special operations.
	DeployInProject := false
	if (operation == string(api.DeployOperation)) && (job.PipelineRecord.ProjectVersionID != "") {
		DeployInProject = true
	}

	ds := store.NewStore()
	defer ds.Close()

	if err := ds.UpdatePipelineRecord(job.PipelineRecord); err != nil {
		logdog.Errorf("Unable to update version status post hook for %+v: %v", job.Pipeline, err)
	}

	if remote, err := remoteManager.FindRemote(job.Pipeline.Repository.Webhook); err == nil {
		if err := remote.PostCommitStatus(job.Pipeline, job.PipelineRecord); err != nil {
			logdog.Errorf("Unable to post commit status to %s: %v", job.Pipeline.Repository.Webhook, err)
		}
	}

	if DeployInProject == false {
		if job.PipelineRecord.Status == api.VersionHealthy {
			if err := ds.AddNewVersion(job.Pipeline.ID, job.PipelineRecord.ID); err != nil {
				logdog.Errorf("Unable to add new PipelineRecord in post hook for %+v: %v", job.PipelineRecord, err)
			}
		} else {
			if err := ds.AddNewFailVersion(job.Pipeline.ID, job.PipelineRecord.ID); err != nil {
				logdog.Errorf("Unable to add new PipelineRecord in post hook for %+v: %v", job.PipelineRecord, err)
			}
		}

	}
	if err := ds.UpdateServiceLastInfo(job.PipelineRecord.ServiceID, job.PipelineRecord.CreateTime, job.PipelineRecord.Name); err != nil {
		logdog.Errorf("Unable to update new version info in service %+v: %v", job.PipelineRecord, err)
	}

	// Use for checking project's version deploy status.
	if DeployInProject == true {
		job.PipelineRecord.FinalStatus = "finished"
	}
	// trigger after create end
	triggerHooks(job, PreStopPhase)
}

// triggerHooks triggers webhooks for specific phase
func triggerHooks(job *store.Job, phase string) {

	hooks := job.Pipeline.Hooks

	if hooks == nil {
		return
	}

	for _, hook := range hooks {
		if hook.Phase == phase {
			logdog.Info("trigger version hook", log.Fields{"phase": phase})
			data := map[string]string{
				"status":      "failed",
				"serviceId":   job.Pipeline.ID,
				"serviceName": job.Pipeline.Name,
				"versionId":   job.PipelineRecord.VersionID,
				"versionName": job.PipelineRecord.Name,
				"image":       "",
			}
			registryLocation := workerOptions.WorkerEnvs.RegistryLocation
			if phase == PostStartPhase {
				data["status"] = "building"
			} else if job.PipelineRecord.Status == api.EventStatusSuccess {
				data["status"] = "success"
				// registry/username/service_name:version_name
				data["image"] = fmt.Sprintf("%s/%s/%s:%s",
					registryLocation,
					strings.ToLower(job.Pipeline.Owner),
					strings.ToLower(job.Pipeline.Name),
					job.PipelineRecord.Name)
			}
			jsonStr, _ := json.Marshal(data)

			client := getClientWithOauth2(hook.Token)
			_, err := client.Post(hook.Callback, "application/json", bytes.NewBuffer(jsonStr))
			if err != nil {
				logdog.Error("error occur when callback", logdog.Fields{"url": hook.Callback, "err": err})
				break
			}

		}
	}
}

func getClientWithOauth2(token *oauth2.Token) *http.Client {
	var client *http.Client
	if token != nil {
		client = oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(token))
	} else {
		client = &http.Client{}
	}
	return client
}

// Resource2Quota TODO: FIXME
func Resource2Quota(resource api.BuildResource, def cloud.Quota) cloud.Quota {
	if resource.CPU == 0.0 && resource.Memory == 0.0 {
		return def
	}

	quota := cloud.Quota{
		cloud.ResourceLimitsCPU:    cloud.MustParseCPU(resource.CPU / 1024),
		cloud.ResourceLimitsMemory: cloud.MustParseMemory(resource.Memory),
	}

	return quota
}

// getNamespaceFromJob gets the namespace from job data for worker. Will return empty string if can not get it.
func getNamespaceFromJob(job *store.Job) string {
	if job == nil {
		log.Error("Can not get namespace from data as job is nil")
		return ""
	}

	if v, e := job.pipelineRecord.Data["namespace"]; e {
		if sv, ok := v.(string); ok {
			return sv
		}
	}

	return ""
}
