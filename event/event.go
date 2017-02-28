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

	"golang.org/x/oauth2"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/notify"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/store"
	"github.com/caicloud/cyclone/websocket"
	"gopkg.in/mgo.v2/bson"
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

	// PostStartPhase hooks phase
	PostStartPhase = "postStart"
	// PreStopPhase hooks phase
	PreStopPhase = "preStop"
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

	// Include cancel manual or timeout
	if event.Status == api.EventStatusCancel {
		versionLog, err := websocket.StoreTopic(event.Service.UserID, event.Service.ServiceID, event.Version.VersionID)
		if err == nil {
			Log := api.VersionLog{
				VerisonID: event.Version.VersionID,
				Logs:      versionLog,
			}

			ds := store.NewStore()
			defer ds.Close()
			_, err = ds.NewVersionLogDocument(&Log)
			if err != nil {
				log.Error(err)
			}
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
	if event.Service.Repository.Status == api.RepositoryAccepted {
		if event.Status == api.EventStatusSuccess {
			event.Service.Repository.Status = api.RepositoryHealthy
		} else {
			event.Service.Repository.Status = api.RepositoryMissing
		}
	}

	ds := store.NewStore()
	defer ds.Close()

	if err := ds.UpdateRepositoryStatus(event.Service.ServiceID, event.Service.Repository.Status); err != nil {
		event.Service.Repository.Status = api.RepositoryInternalError
		log.Errorf("Unable to update repository status in post hook for %+v: %v\n", event.Service, err)
	}

	if remote, err := remoteManager.FindRemote(event.Service.Repository.Webhook); err == nil &&
		event.Service.Repository.Status == api.RepositoryHealthy {
		if err := remote.CreateHook(&event.Service); err != nil {
			log.ErrorWithFields("create hook failed", log.Fields{"user_id": event.Service.UserID, "error": err})
		}
	}

	autoPublishVersion(&event.Service)
}

// autoPublishVersion create a new version with
// publish opration automatically after service created successfully
func autoPublishVersion(service *api.Service) {
	if service == nil {
		return
	}
	if service.PublishNow {
		ds := store.NewStore()
		defer ds.Close()
		version := api.Version{}
		version.ServiceID = service.ServiceID
		version.Name = bson.NewObjectId().Hex()
		version.Description = "trigger by auto publish"
		version.CreateTime = time.Now()
		version.Status = api.VersionPending
		version.URL = service.Repository.URL
		version.SecurityCheck = false
		version.Operation = "publish"
		_, err := ds.NewVersionDocument(&version)
		if err != nil {
			message := "Unable to create version document in database"
			log.ErrorWithFields(message, log.Fields{"error": err})
			return
		}

		err = SendCreateVersionEvent(service, &version)
		if err != nil {
			message := "Unable to create build version job"
			log.ErrorWithFields(message, log.Fields{"service": service, "version": version, "error": err})
			return
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

	// trigger after get an valid worker
	triggerHooks(event, PostStartPhase)

	err = w.DoWork(event)
	if err != nil {
		return err
	}

	webhook := event.Service.Repository.Webhook
	if remote, err := remoteManager.FindRemote(webhook); webhook == api.GITHUB && err == nil {
		if err = remote.PostCommitStatus(&event.Service, &event.Version); err != nil {
			log.Errorf("Unable to post commit status to github: %v", err)
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

	if remote, err := remoteManager.FindRemote(event.Service.Repository.Webhook); err == nil {
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
	} else {
		notify.Notify(&event.Service, &event.Version, versionLog.Logs)
	}

	// trigger after create end
	triggerHooks(event, PreStopPhase)
}

// triggerHooks triggers webhooks for specific phase
func triggerHooks(event *api.Event, phase string) {

	hooks := event.Service.Hooks

	if hooks == nil {
		return
	}

	for _, hook := range hooks {
		if hook.Phase == phase {
			log.InfoWithFields("trigger version hook", log.Fields{"phase": phase})
			data := map[string]string{
				"status":      "failed",
				"serviceId":   event.Service.ServiceID,
				"serviceName": event.Service.Name,
				"versionId":   event.Version.VersionID,
				"versionName": event.Version.Name,
				"image":       "",
			}
			registryLocation := osutil.GetStringEnv(WORK_REGISTRY_LOCATION, "")
			if phase == PostStartPhase {
				data["status"] = "building"
			} else if event.Status == api.EventStatusSuccess {
				data["status"] = "success"
				// registry/username/service_name:version_name
				data["image"] = fmt.Sprintf("%s/%s/%s:%s",
					registryLocation,
					strings.ToLower(event.Service.Username),
					strings.ToLower(event.Service.Name),
					event.Version.Name)
			}
			jsonStr, _ := json.Marshal(data)

			client := getClientWithOauth2(hook.Token)
			_, err := client.Post(hook.Callback, "application/json", bytes.NewBuffer(jsonStr))
			if err != nil {
				log.ErrorWithFields("error occur when callback", log.Fields{"url": hook.Callback, "err": err})
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
