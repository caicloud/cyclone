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

package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/event"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
)

// createVersion creates a new version from service codebase master branch/trunk,
// validates and saves it. The operation is asynchronous, meaning that when creating
// a service, its service ID is returned and saved in database, but the version
// related resources are not created yet. Cyclone will do the following to actually
// create the version:
//  1. Runs user specified unittest/integration if any (or hook up with jenkins);
//  2. Runs user specified script to build docker container or just run docker build;
// To query build progress, logs, etc, use getVersionLogs API below.
//
// POST: /api/v0.1/:uid/versions/
//
// PAYLOAD (Version):
//   {
//     "name": (string) the version name to create with, e.g. v0.1.0
//     "description": (string) a short description of the version
//     "service_id": (string) service associated with the version
//   }
//
// RESPONSE: (VersionCreationResponse)
//  {
//    "version_id": (string) set IFF creation is accepted
//    "error_msg": (string) set IFF the request fails.
//  }
func createVersion(request *restful.Request, response *restful.Response) {
	// Read out version information.
	version := api.Version{}
	version.Operator = api.APIOperator
	err := request.ReadEntity(&version)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	userID := request.PathParameter("user_id")
	log.InfoWithFields("Cyclone receives creating version request", log.Fields{"user_id": userID, "version": version})

	// Find the target service entity.
	var createResponse api.VersionCreationResponse
	ds := store.NewStore()
	defer ds.Close()

	service, err := ds.FindServiceByID(version.ServiceID)
	if err != nil {
		message := fmt.Sprintf("Unable to find service %v", version.ServiceID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	// To create a version, we must first make sure repository is healthy.
	if service.Repository.Status != api.RepositoryHealthy {
		message := fmt.Sprintf("Repository of service %s is not healthy, current status %s", service.Name, service.Repository.Status)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusNotAcceptable, createResponse)
		return
	}

	versions, err := ds.FindVersionsByCondition(version.ServiceID, version.Name)
	if err == nil && len(versions) > 0 {
		message := fmt.Sprintf("Name of version %s is existed", version.Name)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	// Request looks good, now fill up initial version status.
	version.CreateTime = time.Now()
	version.Status = api.VersionPending
	if "" == version.URL {
		version.URL = service.Repository.URL
	}

	version.YamlDeployStatus = api.DeployNoRun
	for i := 0; i < len(version.DeployPlansStatuses); i++ {
		version.DeployPlansStatuses[i].Status = api.DeployNoRun
	}

	version.SecurityCheck = false

	// Create a new version in database. Note the version is NOT the final version:
	// there can be error when running tests or building docker image. The version
	// ID is only a record that a version build has occurred. If the version build
	// succeeds, it'll be added to the service and is considered as a final version;
	// otherwise, it is just a version recorded in database.
	versionID, err := ds.NewVersionDocument(&version)
	if err != nil {
		message := "Unable to create version document in database"
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	// Start building the version asynchronously, and make sure event is successfully
	// created before return.
	err = sendCreateVersionEvent(service, &version)
	if err != nil {
		message := "Unable to create build version job"
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "service": service, "version": version, "error": err})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	createResponse.VersionID = versionID
	response.WriteEntity(createResponse)
}

// findVersionAndService finds service and version entity based on version id.
func findServiceAndVersion(versionID string) (*api.Service, *api.Version, error) {
	ds := store.NewStore()
	defer ds.Close()

	version, err := ds.FindVersionByID(versionID)
	if err != nil {
		return nil, nil, err
	}
	service, err := ds.FindServiceByID(version.ServiceID)
	if err != nil {
		return nil, nil, err
	}
	return service, version, nil
}

// getVersion finds an version from ID.
//
// GET: /api/v0.1/:uid/versions/:versionID
//
// RESPONSE: (VersionGetResponse)
//  {
//    "version": (object) api.Version object.
//    "error_msg": (string) set IFF the request fails.
//  }
func getVersion(request *restful.Request, response *restful.Response) {
	versionID := request.PathParameter("version_id")
	userID := request.PathParameter("user_id")

	var getResponse api.VersionGetResponse
	ds := store.NewStore()
	defer ds.Close()

	result, err := ds.FindVersionByID(versionID)
	if err != nil {
		message := fmt.Sprintf("Unable to find version %v", versionID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		getResponse.ErrorMessage = message
	} else {
		getResponse.Version = *result
	}

	response.WriteEntity(getResponse)
}

// listVersions returns all versions belong to a user and service.
//
// GET: /api/v0.1/:uid/services/:service_id/versions
//
// RESPONSE: (VersionListResponse)
//  {
//    "versions": (array) a list of api.Version object.
//    "error_msg": (string) set IFF the request fails.
//  }
func listVersions(request *restful.Request, response *restful.Response) {
	serviceID := request.PathParameter("service_id")
	userID := request.PathParameter("user_id")

	var listResponse api.VersionListResponse
	ds := store.NewStore()
	defer ds.Close()

	result, err := ds.FindVersionsByServiceID(serviceID)

	if err != nil {
		message := "Unable to list version"
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		listResponse.ErrorMessage = message
	} else {
		listResponse.Versions = result
	}

	response.WriteEntity(listResponse)
}

//
// POST: /api/v0.1/:uid/versions/:versionID/cancelbuild
//
// RESPONSE: (VersionConcelResponse)
//  {
//    "result": (string) success.
//    "error_msg": (string) set IFF the request fails.
//  }
func cancelVersion(request *restful.Request, response *restful.Response) {
	versionID := request.PathParameter("version_id")
	userID := request.PathParameter("user_id")

	var cancelresponse api.VersionConcelResponse
	log.Infof("user(%s) cance build version %s", userID, versionID)
	e, err := event.LoadEventFromEtcd(api.EventID(versionID))
	if err != nil {
		message := fmt.Sprintf("Unable to find event by versonID %v", versionID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		cancelresponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, cancelresponse)
		return
	}
	w, err := event.LoadWorker(e)
	if err != nil {
		message := fmt.Sprintf("Unable to load worker by event")
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		cancelresponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, cancelresponse)
		return
	}

	err = w.Fire()
	if err != nil {
		message := fmt.Sprintf("Unable to cancel event")
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		cancelresponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, cancelresponse)
		return
	}

	cancelresponse.Result = "success"
	response.WriteEntity(cancelresponse)
}
