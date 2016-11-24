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
	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/notify"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
)

const (
	// CreateServiceOps defines the operation to create a service, currently it
	// involves: clone repository (to check if repository exists).
	CreateServiceOps api.Operation = "create-service"

	// CreateVersionOps defines the operation to create a version, currently it
	// involves: clone repository, run CI if caicloud.yml exists and the operation
	// field in the version is not "Publish", thern tag it based on version name,
	// build docker image and push to caicloud registry, then run the postbuild
	// hook.
	CreateVersionOps api.Operation = "create-version"

	// CreateProjectVersionOps defines the operation to create a project version.
	CreateProjectVersionOps api.Operation = "create-projectversion"
)

//type EventID string
//type Operation string

// Handler is the type for event handler.
type Handler func(*api.Event) error

// PostHook is the type for post hook.
type PostHook func(*api.Event)

// mapOperation record event operation handler and post hook.
var mapOperation map[api.Operation]Operation

// Operation define event operation.
type Operation struct {
	// PostHook is called when event create.
	Handler Handler
	// PostHook is called when event has finished, and event status is set.
	PostHook PostHook
}

// initOperationMap registers event handlers to map.
func initOperationMap() {
	mapOperation = make(map[api.Operation]Operation)

	// create service ops
	mapOperation[CreateServiceOps] = Operation{
		Handler:  createServiceHandler,
		PostHook: createServicePostHook,
	}

	// create version ops
	mapOperation[CreateVersionOps] = Operation{
		Handler:  createVersionHandler,
		PostHook: createVersionPostHook,
	}
}

// handleEvent is the event create handler.
func handleEvent(event *api.Event) error {
	return mapOperation[event.Operation].Handler(event)
}

// postHookEvent is the event finished post hook.
func postHookEvent(event *api.Event) {
	mapOperation[event.Operation].PostHook(event)

	w, err := LoadWorker(event)
	if err != nil {
		log.Errorf("load worker err: %v", err)
		return
	}
	err = w.Fire()
	if err != nil {
		log.Errorf("fire worker err: %v", err)
		return
	}

	if err := resourceManager.ReleaseResource(event); err != nil {
		log.Errorf("Unable to release resource %v", err)
	}
	// Release resources of worker node.
	ds := store.NewStore()
	defer ds.Close()
	nodes, err := ds.FindWorkerNodesByDockerHost(event.WorkerInfo.DockerHost)
	if err != nil || len(nodes) != 1 {
		log.Errorf("find worker node err: %v", err)
	} else {
		nodes[0].LeftResource.Memory += event.WorkerInfo.UsedResource.Memory
		nodes[0].LeftResource.CPU += event.WorkerInfo.UsedResource.CPU
		_, err := ds.UpsertWorkerNodeDocument(&(nodes[0]))
		if err != nil {
			log.Errorf("release worker node resource err: %v", err)
		}
	}
}

// createServiceHander is the create service handler.
func createServiceHandler(event *api.Event) error {
	log.Infof("create service handler")
	w, err := NewWorker(event)
	if err != nil {
		return err
	}

	return w.DoWork(event)
}

// createServicePostHook is the create service post hook.
func createServicePostHook(event *api.Event) {
	log.Infof("create service post hook")
	if event.Status == api.EventStatusSuccess {
		event.Service.Repository.Status = api.RepositoryHealthy
	} else {
		event.Service.Repository.Status = api.RepositoryMissing
	}
	ds := store.NewStore()
	defer ds.Close()

	if err := ds.UpdateRepositoryStatus(event.Service.ServiceID, event.Service.Repository.Status); err != nil {
		event.Service.Repository.Status = api.RepositoryInternalError
		log.Errorf("Unable to update repository status in post hook for %+v: %v\n", event.Service, err)
	}

	remote, err := remoteManager.FindRemote(event.Service.Repository.Webhook)
	if err != nil {
		log.ErrorWithFields("Unable to get remote according code repository", log.Fields{"user_id": event.Service.UserID})
		return
	}
	if event.Service.Repository.Status == api.RepositoryHealthy {
		if err := remote.CreateHook(&event.Service); err != nil {
			log.ErrorWithFields("create hook failed", log.Fields{"user_id": event.Service.UserID, "error": err})
		}
	}

}

// createVersionHandler is the create version handler.
func createVersionHandler(event *api.Event) error {
	log.Infof("create version handler")
	w, err := NewWorker(event)
	if err != nil {
		return err
	}

	err = w.DoWork(event)
	if err != nil {
		return err
	}

	if event.Service.Repository.Webhook == api.GITHUB {
		remote, err := remoteManager.FindRemote(event.Service.Repository.Webhook)
		if err != nil {
			log.ErrorWithFields("Unable to get remote according coderepository", log.Fields{"user_id": event.Service.UserID})
		} else {
			if err = remote.PostCommitStatus(&event.Service, &event.Version); err != nil {
				log.Errorf("Unable to post commit status to github: %v", err)
			}
		}
	}
	return nil
}

// createVersionPostHook is the create version post hook.
func createVersionPostHook(event *api.Event) {
	log.Infof("create version post hook")
	if event.Status == api.EventStatusSuccess {
		event.Version.Status = api.VersionHealthy
	} else if event.Status == api.EventStatusCancel {
		event.Version.Status = api.VersionCancel
	} else {
		event.Version.Status = api.VersionFailed
		event.Version.ErrorMessage = event.ErrorMessage
	}

	operation := string(event.Version.Operation)
	// Record that whether this event is a deploy for project. According this flag, we will make some special operations.
	DeployInProject := false
	if (operation == string(api.DeployOperation)) && (event.Version.ProjectVersionID != "") {
		DeployInProject = true
	}

	ds := store.NewStore()
	defer ds.Close()

	if err := ds.UpdateVersionDocument(event.Version.VersionID, event.Version); err != nil {
		log.Errorf("Unable to update version status post hook for %+v: %v", event.Version, err)
	}

	remote, err := remoteManager.FindRemote(event.Service.Repository.Webhook)
	if err != nil {
		log.ErrorWithFields("Unable to get remote according coderepository", log.Fields{"user_id": event.Service.UserID})
	} else {
		if err := remote.PostCommitStatus(&event.Service, &event.Version); err != nil {
			log.Errorf("Unable to post commit status to %s: %v", event.Service.Repository.Webhook, err)
		}
	}

	if DeployInProject == false {
		if event.Version.Status == api.VersionHealthy {
			if err := ds.AddNewVersion(event.Version.ServiceID, event.Version.VersionID); err != nil {
				log.Errorf("Unable to add new version in post hook for %+v: %v", event.Version, err)
			}
		} else {
			if err := ds.AddNewFailVersion(event.Version.ServiceID, event.Version.VersionID); err != nil {
				log.Errorf("Unable to add new version in post hook for %+v: %v", event.Version, err)
			}
		}

	}
	if err := ds.UpdateServiceLastInfo(event.Version.ServiceID, event.Version.CreateTime, event.Version.Name); err != nil {
		log.Errorf("Unable to update new version info in service %+v: %v", event.Version, err)
	}

	// Use for checking project's version deploy status.
	if DeployInProject == true {
		event.Version.FinalStatus = "finished"
	}

	// TODO: poll version log, not query once.
	versionLog, err := ds.FindVersionLogByVersionID(event.Version.VersionID)
	if err != nil {
		log.Warnf("Notify error, getting version failed: %v", err)
		return
	}
	notify.Notify(&event.Service, &event.Version, versionLog.Logs)
}
