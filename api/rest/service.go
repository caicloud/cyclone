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
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
)

// createService creates a service, validates and saves it. The operation is asynchronous,
// meaning that when creating a service, its service ID is returned and saved in database,
// but the service related resources are not created yet. For example, the service's
// codebase url hasn't been verified. Client should use service ID to query the status
// of the service via getService API below.
//
// POST: /api/v0.1/:uid/services/
//
// PAYLOAD (api.Service):
//   {
//     "name": (string) the service name to create with
//     "description": (string) a short description of the service
//     "username": (string) the username
//     "repository": {
//       "url": (string) url path of the service repository
//       "vcs": (string) version control tool used to host the repository, options:
//          git, fake (for testing)
//     }
//     "build_path": (string) Path of the file used to create service version. By default,
//        Cyclone will create service version using "docker build", assming there is a
//        Dockerfile at top of the repository.
//     ...
//   }
//
// RESPONSE: (ServiceCreationResponse)
//  {
//    "service_id": (string) set IFF creation is accepted.
//    "error_msg": (string) set IFF the request fails.
//  }
func createService(request *restful.Request, response *restful.Response) {
	// Read out posted service information.
	service := api.Service{}
	err := request.ReadEntity(&service)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}
	var createResponse api.ServiceCreationResponse
	// Request looks good, now fill up initial service status.
	userID := request.PathParameter("user_id")
	service.UserID = userID
	service.Repository.Status = api.RepositoryAccepted
	service.LastCreateTIme = time.Now()

	ds := store.NewStore()
	defer ds.Close()

	// Find the target service entity by UserID.
	services, err := ds.FindServiceByCondition(userID, service.Name)
	if err == nil && len(services) > 0 {
		message := fmt.Sprintf("Name of service %s is existed", service.Name)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	// Need Jenkins according Jconfig.Address, so here should check if username
	// or password is empty.
	if service.Jconfig.Address != "" {
		if service.Jconfig.Username == "" || service.Jconfig.Password == "" {
			message := "username or password in Jconfig is empty!"
			log.ErrorWithFields(message, log.Fields{"username": service.Jconfig.Username,
				"password": service.Jconfig.Password})
			createResponse.ErrorMessage = message
			response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
			return
		}
	}

	log.InfoWithFields("Cyclone receives creating service request",
		log.Fields{"user_id": userID, "service_name": service.Name})

	// Create service in database (but not ready to be used yet).
	serviceID, err := ds.NewServiceDocument(&service)
	if err != nil {
		message := "Unable to create service document in database"
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "service": service, "error": err})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	// Start creating the service asynchronously, and make sure event is
	// successfully acked before return.
	err = sendCreateServiceEvent(&service)

	if err != nil {
		message := "Unable to create new service job"
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "service": service, "error": err})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	// Return service creation response.
	createResponse.ServiceID = serviceID
	response.WriteHeaderAndEntity(http.StatusAccepted, createResponse)
}

// getService finds a service from service ID and user ID.
//
// GET: /api/v0.1/:uid/services/:service_id
//
// RESPONSE: (ServiceGetResponse)
//  {
//    "service": (object) api.Service object.
//    "error_msg": (string) set IFF the request fails.
//  }
func getService(request *restful.Request, response *restful.Response) {
	serviceID := request.PathParameter("service_id")
	userID := request.PathParameter("user_id")

	var getResponse api.ServiceGetResponse
	ds := store.NewStore()
	defer ds.Close()

	result, err := ds.FindServiceByID(serviceID)
	if err != nil {
		message := fmt.Sprintf("Unable to find service %v", serviceID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		getResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	}
	getResponse.Service = *result

	response.WriteEntity(getResponse)
}

// listServices returns all services belong to a user.
//
// GET: /api/v0.1/:uid/services
//
// RESPONSE: (ServiceListResponse)
//  {
//    "services": (array) a list of api.Service object.
//    "error_msg": (string) set IFF the request fails.
//  }
func listServices(request *restful.Request, response *restful.Response) {
	userID := request.PathParameter("user_id")
	ds := store.NewStore()
	defer ds.Close()

	var listResponse api.ServiceListResponse
	result, err := ds.FindServicesByUserID(userID)

	if err != nil {
		message := "Unable to list service"
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		listResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	}
	listResponse.Services = result

	response.WriteEntity(listResponse)
}

// deleteService that delete the service by service_id.
//
// DELETE: /api/v0.1/:uid/services/:service_id
//
// RESPONSE: (ServiceDelResponse)
//  {
//    "result": (string) the result of deleting service
//    "error_msg": (string) set IFF the request fails.
//  }
func deleteService(request *restful.Request, response *restful.Response) {
	serviceID := request.PathParameter("service_id")

	var deleteResponse api.ServiceDelResponse
	ds := store.NewStore()
	defer ds.Close()

	service, err := ds.FindServiceByID(serviceID)
	if err != nil {
		message := "Unable to find service"
		log.ErrorWithFields(message, log.Fields{"service_id": serviceID, "error": err})
		deleteResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	}

	// // Check if this service has referenced in project.
	// projects, err := ds.FindProjectsRelateService(service_id)
	// if err != nil {
	// 	message := "Unable to find relate projects"
	// 	log.ErrorWithFields(message, log.Fields{"service_id": serviceID})
	// 	deleteResponse.ErrorMessage = message
	// 	response.WriteHeaderAndEntity(http.StatusInternalServerError, message)
	// 	return
	// }
	// if len(projects) != 0 {
	// 	message := "Has relate projects"
	// 	log.ErrorWithFields(message, log.Fields{"service_id": serviceID})
	// 	deleteResponse.ErrorMessage = message
	// 	response.WriteHeaderAndEntity(http.StatusInternalServerError, message)
	// 	return
	// }

	remote, err := remoteManager.FindRemote(service.Repository.Webhook)
	if err != nil {
		log.ErrorWithFields("Unable to get remote according coderepository", log.Fields{"user_id": service.UserID})
	} else {
		if err := remote.DeleteHook(service); err != nil {
			log.ErrorWithFields("delete hook fail", log.Fields{"service_id": serviceID, "error": err})
		}
	}

	versions, err := ds.FindVersionsByServiceID(serviceID)
	if err != nil {
		message := "Unable to find version"
		log.ErrorWithFields(message, log.Fields{"service_id": serviceID, "error": err})
		deleteResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, message)
		return
	}
	// Delete version one by one.
	for _, version := range versions {
		err := ds.DeleteVersionByID(version.VersionID)
		if err != nil {
			message := "Unable to delete version"
			log.ErrorWithFields(message, log.Fields{"service_id": serviceID, "version_id": version.VersionID})
			deleteResponse.ErrorMessage = message
			response.WriteHeaderAndEntity(http.StatusInternalServerError, message)
			return
		}
	}

	// Delete the service in DB.
	err = ds.DeleteServiceByID(serviceID)
	if err != nil {
		message := "Unable to delete service"
		log.ErrorWithFields(message, log.Fields{"service_id": serviceID, "error": err})
		deleteResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, message)
		return
	}

	deleteResponse.Result = "success"
	response.WriteEntity(deleteResponse)
}

// setService set a service, validates and saves it.
//
// PUT: /api/v0.1/:uid/services/:service_id
//
// PAYLOAD (Service):
//   {
//     "name": (string) the service name to create with
//     "description": (string) a short description of the service
//     "username": (string) the username
//     "repository": {
//       "url": (string) url path of the service repository
//       "vcs": (string) version control tool used to host the repository, options:
//          git, fake (for testing)
//     }
//     "build_path": (string) Path of the file used to create service version. By default,
//        Cyclone will create service version using "docker build", assming there is a
//        Dockerfile at top of the repository.
//   }
//
// RESPONSE: (ServiceSetResponse)
//  {
//    "service_id": (string) set IFF creation is accepted.
//    "error_msg": (string) set IFF the request fails.
//  }
func setService(request *restful.Request, response *restful.Response) {
	// Read out posted service information.
	service := api.Service{}
	err := request.ReadEntity(&service)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}
	var setResponse api.ServiceSetResponse
	// Request looks good, now fill up initial service status.
	userID := request.PathParameter("user_id")
	serviceID := request.PathParameter("service_id")

	ds := store.NewStore()
	defer ds.Close()

	// Find the target service by ServiceID.
	servicePre, err := ds.FindServiceByID(serviceID)
	if nil != err {
		message := fmt.Sprintf("Find service %s err: %v", serviceID, err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	// Now, can only set description, webhook, profile, deploy_plans.
	servicePre.Description = service.Description
	if servicePre.Repository.Webhook != service.Repository.Webhook {
		remote, err := remoteManager.FindRemote(servicePre.Repository.SubVcs)
		if err != nil {
			message := "Unable to get remote according coderepository"
			log.ErrorWithFields(message, log.Fields{"service_id": serviceID})
			setResponse.ErrorMessage = message
			response.WriteHeaderAndEntity(http.StatusInternalServerError, message)
			return
		}
		if service.Repository.Webhook != "" {
			servicePre.Repository.Webhook = service.Repository.Webhook
			remote.CreateHook(servicePre)
		} else {
			remote.DeleteHook(servicePre)
			servicePre.Repository.Webhook = service.Repository.Webhook
		}

	}
	servicePre.Profile = service.Profile

	// projects, err := ds.FindProjectsRelateService(serviceID)
	// if err != nil {
	// 	message := "Unable to find relate projects"
	// 	log.ErrorWithFields(message, log.Fields{"service_id": serviceID})
	// 	setResponse.ErrorMessage = message
	// 	response.WriteHeaderAndEntity(http.StatusInternalServerError, message)
	// 	return
	// }
	// for _, project := range projects {
	// 	changeDeployinProject(&project.Services, serviceID, service.DeployPlans)
	// 	_, err = ds.UpsertProjectDocument(&project)
	// 	if err != nil {
	// 		message := "Unable to set project document in database"
	// 		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
	// 		setResponse.ErrorMessage = message
	// 		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
	// 		return
	// 	}
	// }

	servicePre.DeployPlans = service.DeployPlans
	_, err = ds.UpsertServiceDocument(servicePre)
	if nil != err {
		message := fmt.Sprintf("Set service %s err: %v", serviceID, err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	// Return service creation response.
	setResponse.ServiceID = serviceID
	response.WriteHeaderAndEntity(http.StatusAccepted, setResponse)
}
