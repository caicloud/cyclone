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

// getVersionLog finds an version log from versionID.
//
// GET: /api/v0.1/:uid/versions/:versionID/logs
//
// RESPONSE: (VersionLogGetResponse)
//  {
//    "logs": (string) log
//    "error_msg": (string) set IFF the request fails.
//  }
func getVersionLog(request *restful.Request, response *restful.Response) {
	versionID := request.PathParameter("version_id")
	userID := request.PathParameter("user_id")

	var getResponse api.VersionLogGetResponse

	ds := store.NewStore()
	defer ds.Close()
	result, err := ds.FindVersionLogByVersionID(versionID)
	if err != nil {
		message := fmt.Sprintf("Unable to find version log by versionID %v", versionID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		getResponse.ErrorMessage = message
	} else {
		getResponse.Logs = result.Logs
	}

	response.WriteEntity(getResponse)
}

// createVersionLog creates an version log.
//
// POST: /api/v0.1/:uid/versions/:versionID/logs
//
// RESPONSE: (VersionLogGetResponse)
//  {
//    "error_msg": (string) set IFF the request fails.
//  }
func createVersionLog(request *restful.Request, response *restful.Response) {
	versionID := request.PathParameter("version_id")
	userID := request.PathParameter("user_id")

	versionLog := api.VersionLog{}
	err := request.ReadEntity(&versionLog)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, "Unable to parse request body")
		return
	}

	var createResponse api.VersionLogCreateResponse

	ds := store.NewStore()
	defer ds.Close()
	result, err := ds.NewVersionLogDocument(&versionLog)
	if err != nil {
		message := fmt.Sprintf("Unable to create version log by versionID %v", versionID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		createResponse.ErrorMessage = message
	} else {
		createResponse.LogID = result
	}

	response.WriteEntity(createResponse)
}
