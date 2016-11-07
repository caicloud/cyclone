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

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
)

// setResource set a resource, validates and saves it..
//
// PUT: /api/v0.1/:uid/resources
//
// RESPONSE: (ResourceSetResponse)
//  {
//    "result": (string) set IFF set is accepted.
//    "error_msg": (string) set IFF the request fails.
//  }
func setResource(request *restful.Request, response *restful.Response) {
	resource := api.Resource{}
	err := request.ReadEntity(&resource)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}
	var setResponse api.ResourceSetResponse
	userID := request.PathParameter("user_id")

	ds := store.NewStore()
	defer ds.Close()

	resourcePre, err := ds.FindResourceByID(userID)
	if err != nil {
		resource.UserID = userID
		resource.LeftResource.Memory = resource.TotalResource.Memory
		resource.LeftResource.CPU = resource.TotalResource.CPU
	} else {
		resource.UserID = userID
		resource.LeftResource.Memory = resourcePre.LeftResource.Memory + resource.TotalResource.Memory -
			resourcePre.TotalResource.Memory
		resource.LeftResource.CPU = resourcePre.LeftResource.CPU + resource.TotalResource.CPU - resourcePre.TotalResource.CPU
		if resource.LeftResource.CPU < 0 || resource.LeftResource.Memory < 0 {
			response.AddHeader("Content-Type", "text/plain")
			response.WriteErrorString(http.StatusBadRequest, "the total cpu and memory < used cpu and memory ")
			return
		}
	}
	if err := ds.UpdateResourceDocument(&resource); err != nil {
		setResponse.ErrorMessage = "Unable to update new resource document"
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	// Return resource set response.
	setResponse.Result = "success"
	response.WriteHeaderAndEntity(http.StatusAccepted, setResponse)
}

// getResource finds a resource by user ID.
//
// GET: /api/v0.1/:uid/resources
//
// RESPONSE: (ResourceGetResponse)
//  {
//    "resource": (object) api.Resource object.
//    "error_msg": (string) set IFF the request fails.
//  }
func getResource(request *restful.Request, response *restful.Response) {
	userID := request.PathParameter("user_id")
	var getResponse api.ResourceGetResponse

	ds := store.NewStore()
	defer ds.Close()
	result, err := ds.FindResourceByID(userID)
	if err != nil {
		message := fmt.Sprintf("Unable to find resource")
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		resource := api.Resource{}
		resource.UserID = userID
		resource.TotalResource.CPU = resourceManager.GetCpuuser()
		resource.TotalResource.Memory = resourceManager.GetMemoryuser()
		resource.PerResource.CPU = resourceManager.GetCpucontainer()
		resource.PerResource.Memory = resourceManager.GetMemorycontainer()
		resource.LeftResource.CPU = resource.TotalResource.CPU
		resource.LeftResource.Memory = resource.TotalResource.Memory
		if err := ds.UpdateResourceDocument(&resource); err != nil {
			log.Errorf("Unable to update new resource document")
		}
		getResponse.Resource = resource
	} else {
		resource := api.Resource{}
		resource.UserID = result.UserID
		resource.TotalResource.CPU = result.TotalResource.CPU
		resource.TotalResource.Memory = result.TotalResource.Memory
		resource.PerResource.CPU = result.PerResource.CPU
		resource.PerResource.Memory = result.PerResource.Memory
		resource.LeftResource.CPU = result.LeftResource.CPU
		resource.LeftResource.Memory = result.LeftResource.Memory
		getResponse.Resource = resource
	}

	response.WriteEntity(getResponse)
}
