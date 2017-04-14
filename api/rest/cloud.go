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

	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/event"
	"github.com/caicloud/cyclone/store"
	restful "github.com/emicklei/go-restful"
	"github.com/zoumo/logdog"
)

// createCloud add a cloud to cyclone,
//
// POST: /api/v0.1/clouds
//
// PAYLOAD (cloud.Options):
//   {
//     	"kind": (string) cloud type such as docker and kubernetes
//     	"name": (string) cloud name
//     	"host": (string) cloud host url
//     	"insecure": (bool) Optional server should be accessed without verifying the TLS certificate. For testing only.
//     	"dockerCertPath": (string) Optional docker cert path
//  	"k8sInCluster": (bool) Optional. set true if cyclone runs in k8s cluster and
//                             use the same cluster as cyclone cloud provider
//		"k8sNamespace": (string) Optional k8s cloud namespace to use
// 		"k8sBearerToken": (string) Optional k8s bearer token
//   }
//
// RESPONSE: (cloud.Options)
//  {
//      just like payload
//  }
func createCloud(request *restful.Request, response *restful.Response) {
	cloudOpt := cloud.Options{}
	err := request.ReadEntity(&cloudOpt)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	_, ok := event.CloudController.GetCloud(cloudOpt.Name)
	if ok {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusConflict, fmt.Sprintf("%s cloud already exists", cloudOpt.Name))
		return
	}

	err = event.CloudController.AddClouds(cloudOpt)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, err.Error())

		return
	}

	cloud, _ := event.CloudController.GetCloud(cloudOpt.Name)
	opts := cloud.GetOptions()

	ds := store.NewStore()
	defer ds.Close()

	err = ds.InsertCloud(&opts)
	if err != nil {
		logdog.Error("Can not add cloud to database", logdog.Fields{"cloud": opts.Name, "kind": opts.Kind})
	}

	response.WriteHeaderAndJson(http.StatusCreated, opts, restful.MIME_JSON)
}

// listCloud get all clouds registed in cyclone
//
// GET: /api/v0.1/clouds
//
// RESPONSE (cloud.Options):
//   [
//   	{
//     		"kind": (string) cloud type such as docker and kubernetes
//     		"name": (string) cloud name
//     		"host": (string) cloud host url
//     		"insecure": (bool) Optional server should be accessed without verifying the TLS certificate. For testing only.
//     		"dockerCertPath": (string) Optional docker cert path
//  		"k8sInCluster": (bool) Optional. set true if cyclone runs in k8s cluster and
//                       	      use the same cluster as cyclone cloud provider
//			"k8sNamespace": (string) Optional k8s cloud namespace to use
// 			"k8sBearerToken": (string) Optional k8s bearer token
//   	},
//      ...
//   ]
//
func listCloud(request *restful.Request, response *restful.Response) {
	resp := make(map[string]cloud.Options)
	for name, cloud := range event.CloudController.Clouds {
		resp[name] = cloud.GetOptions()
	}
	response.WriteAsJson(resp)
}

// deleteCloud delete a cloud
//
// DELETE: /api/v0.1/clouds
//
// RESPONSE 204 NO CONTENT
func deleteCloud(request *restful.Request, response *restful.Response) {
	cloudName := request.PathParameter("cloudName")
	event.CloudController.DeleteCloud(cloudName)

	ds := store.NewStore()
	defer ds.Close()
	ds.DeleteCloudByName(cloudName)

	response.WriteHeader(http.StatusNoContent)
}

// upsertCloud upsert a cloud to cyclone,
//
// PUST: /api/v0.1/clouds
//
// PAYLOAD (cloud.Options):
//   {
//     	"kind": (string) cloud type such as docker and kubernetes
//     	"name": (string) cloud name
//     	"host": (string) cloud host url
//     	"insecure": (bool) Optional server should be accessed without verifying the TLS certificate. For testing only.
//     	"dockerCertPath": (string) Optional docker cert path
//  	"k8sInCluster": (bool) Optional. set true if cyclone runs in k8s cluster and
//                             use the same cluster as cyclone cloud provider
//		"k8sNamespace": (string) Optional k8s cloud namespace to use
// 		"k8sBearerToken": (string) Optional k8s bearer token
//   }
//
// RESPONSE: (cloud.Options)
//  {
//      just like payload
//  }
func upsertCloud(request *restful.Request, response *restful.Response) {
	cloudName := request.PathParameter("cloudName")
	cloudOpt := cloud.Options{}
	err := request.ReadEntity(&cloudOpt)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	cloudOpt.Name = cloudName
	event.CloudController.DeleteCloud(cloudOpt.Name)
	err = event.CloudController.AddClouds(cloudOpt)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, err.Error())
		return
	}

	cloud, _ := event.CloudController.GetCloud(cloudOpt.Name)
	opts := cloud.GetOptions()

	ds := store.NewStore()
	defer ds.Close()

	err = ds.UpsertCloud(&opts)
	if err != nil {
		logdog.Error("Can not add cloud to database", logdog.Fields{"cloud": opts.Name, "kind": opts.Kind})
	}

	response.WriteHeaderAndJson(http.StatusOK, opts, restful.MIME_JSON)
}
