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

func listCloud(request *restful.Request, response *restful.Response) {
	resp := make(map[string]cloud.Options)
	for name, cloud := range event.CloudController.Clouds {
		resp[name] = cloud.GetOptions()
	}
	response.WriteAsJson(resp)
}

func deleteCloud(request *restful.Request, response *restful.Response) {
	cloudName := request.PathParameter("cloudName")
	event.CloudController.DeleteCloud(cloudName)
	response.WriteHeader(http.StatusNoContent)
}
