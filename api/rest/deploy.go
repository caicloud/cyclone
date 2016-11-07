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
//     "deploy_plans": (DeployPlan) deploy plan config
//     "service_id": (string) service associated with the version
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

	if len(deploy.DeployPlans) <= 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to use request body")
		return
	}

	log.Info("fornax receives creating deploy request")

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

	service, err := ds.FindServiceByID(deploy.ServiceID)
	if err != nil {
		message := fmt.Sprintf("Unable to find service %v", deploy.ServiceID)
		log.ErrorWithFields(message, log.Fields{"deploy": deploy, "error": err})
		createResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, createResponse)
		return
	}

	service.DeployID = deployID
	_, err = ds.UpsertServiceDocument(service)
	if nil != err {
		message := fmt.Sprintf("Set service %s err: %v", service.ServiceID, err)
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
// GET: /api/v0.1/:uid/deploy/:deploy_id
//
// RESPONSE: (DeployGetResponse)
//  {
//    "deploy": (object) api.Service object.
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

// setDeploy set a deploy, validates and saves it.
//
// PUT: /api/v0.1/:uid/deploy/:deploy_id
//
// PAYLOAD (Deploy):
//   {
//        "deploy_plans": (DeployPlan) deploy plan config
//   }
//
// RESPONSE: (DeploySetResponse)
//  {
//    "deploy_id": (string) set IFF setting is accepted.
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
	if len(deploy.DeployPlans) <= 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to use request body")
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

	var deleteFlag = false
	for _, deployPrePlan := range deployPre.DeployPlans {
		if deployPrePlan.PlanName == deploy.DeployPlans[0].PlanName {
			deleteFlag = true
		}
	}

	log.Infof("gao %+v", deploy)
	// delete element
	if deleteFlag {
		for _, deployPlan := range deploy.DeployPlans {
			for index, deployPrePlan := range deployPre.DeployPlans {
				if deployPrePlan.PlanName == deployPlan.PlanName {
					deployPre.DeployPlans = append(deployPre.DeployPlans[:index], deployPre.DeployPlans[index+1:]...)
					break
				}
			}
		}
	} else {
		log.Info("gao jinqiao ")
		deployPre.DeployPlans = append(deployPre.DeployPlans, deploy.DeployPlans...)
	}
	log.Infof("gao jianqiao %+v", deployPre.DeployPlans)

	_, err = ds.UpsertDeployDocument(deployPre)
	if nil != err {
		message := fmt.Sprintf("upsert deploy plan %v err: %v", deployPre, err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		setResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, setResponse)
		return
	}

	// Return deploy set response.
	setResponse.DeployID = deployID
	response.WriteHeaderAndEntity(http.StatusAccepted, setResponse)
}
