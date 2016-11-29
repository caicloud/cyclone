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

// createDeploy creates a deploy, validates and saves it.
//
// POST: /api/v0.1/:uid/deploys/
//
// PAYLOAD (api.Deploy):
//   {
//     "deploy_plan": (DeployPlan) deploy plan config
//     "user_id": (string) user associated with the deploy
//   }
//
// RESPONSE: (DeployCreationResponse)
//  {
//    "deploy_id": (string) set IFF creation is accepted.
//    "error_msg": (string) set IFF the request fails.
//  }
func createDeploy(request *restful.Request, response *restful.Response) {
	// Read out posted deploy information
	deploy := api.Deploy{}
	err := request.ReadEntity(&deploy)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	if deploy.UserID == "" {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to get useID from request body")
		return
	}
	log.Info("receives creating deploy request")

	var createResponse api.DeployCreationResponse
	ds := store.NewStore()
	defer ds.Close()

	// Create deploy in database
	deployID, err := ds.NewDeployDocument(&deploy)
	if err != nil {
		message := "Unable to create deploy document in database"
		log.ErrorWithFields(message, log.Fields{"deploy": deploy, "error": err})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	// Return deploy creation response.
	createResponse.DeployID = deployID
	response.WriteHeaderAndEntity(http.StatusAccepted, createResponse)
}

// getDeploy finds a deploy from deploy ID and user ID.
//
// GET: /api/v0.1/:uid/deploys/:deploy_id
//
// RESPONSE: (DeployGetResponse)
//  {
//    "deploy": (object) api.Deploy object.
//    "error_msg": (string) set IFF the request fails.
//  }
func getDeploy(request *restful.Request, response *restful.Response) {
	userID := request.PathParameter("user_id")
	deployID := request.PathParameter("deploy_id")

	var getResponse api.DeployGetResponse
	ds := store.NewStore()
	defer ds.Close()

	result, err := ds.FindDeployByID(deployID)
	if err != nil {
		message := fmt.Sprintf("Unable to find deploy %v", deployID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		getResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	}
	getResponse.Deploy = *result

	response.WriteEntity(getResponse)
}

// listDeploy finds deploys by user ID.
//
// GET: /api/v0.1/:uid/deploys
//
// RESPONSE: (DeployGetResponse)
//  {
//    "deploys": (object) api.Deploy object.
//    "error_msg": (string) set IFF the request fails.
//  }
func listDeploy(request *restful.Request, response *restful.Response) {
	userID := request.PathParameter("user_id")

	var listResponse api.DeployListResponse
	ds := store.NewStore()
	defer ds.Close()

	result, err := ds.FindDeployByUserID(userID)
	if err != nil {
		message := fmt.Sprintf("Unable to find deploy, err: %v", err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		listResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, listResponse)
		return
	}
	listResponse.Deploys = result

	response.WriteEntity(listResponse)
}

// setDeploy set a deploy, validates and saves it.
//
// PUT: /api/v0.1/:uid/deploys/:deploy_id
//
// PAYLOAD (Deploy):
//   {
//        "deploy_plan": (DeployPlan) deploy plan config
//   }
//
// RESPONSE: (DeploySetResponse)
//  {
//    "result": (string) set IFF setting is accepted.
//    "error_msg": (string) set IFF the request fails.
//  }
func setDeploy(request *restful.Request, response *restful.Response) {
	// Read out posted deploy information.
	deploy := api.Deploy{}
	err := request.ReadEntity(&deploy)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	var setResponse api.DeploySetResponse
	userID := request.PathParameter("user_id")
	deployID := request.PathParameter("deploy_id")

	ds := store.NewStore()
	defer ds.Close()

	// Find the target deploy by deployID.
	deployPre, err := ds.FindDeployByID(deployID)
	if nil != err {
		message := fmt.Sprintf("Find deploy %s err: %v", deployID, err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	if deployPre.UserID != userID {
		message := fmt.Sprintf("UserID is not match")
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	deployPre.DeployPlan = deploy.DeployPlan
	_, err = ds.UpsertDeployDocument(deployPre)
	if nil != err {
		message := fmt.Sprintf("upsert deploy plan %v err: %v", deployPre, err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	// Return deploy set response.
	setResponse.Result = "success"
	response.WriteHeaderAndEntity(http.StatusAccepted, setResponse)
}

// delDeploy delete a deploy, validates and saves it.
//
// DELETE: /api/v0.1/:uid/deploys/:deploy_id
//
// RESPONSE: (DeployDelResponse)
//  {
//    "result": (string) set IFF setting is accepted.
//    "error_msg": (string) set IFF the request fails.
//  }
func delDeploy(request *restful.Request, response *restful.Response) {
	var delResponse api.DeployDelResponse
	userID := request.PathParameter("user_id")
	deployID := request.PathParameter("deploy_id")

	ds := store.NewStore()
	defer ds.Close()

	// Find the target deploy by deployID.
	_, err := ds.FindDeployByID(deployID)
	if nil != err {
		message := fmt.Sprintf("Find deploy %s err: %v", deployID, err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		delResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, delResponse)
		return
	}

	err = ds.DeleteDeployByID(deployID)
	if nil != err {
		message := fmt.Sprintf("delete deploy by %s err: %v", deployID, err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		delResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, delResponse)
		return
	}

	// Return deploy delete response.
	delResponse.Result = "success"
	response.WriteHeaderAndEntity(http.StatusAccepted, delResponse)
}
