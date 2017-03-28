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
	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/event"
	"github.com/emicklei/go-restful"
	"github.com/zoumo/logdog"
)

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
	// userID := request.PathParameter("user_id")
	var getResponse api.ResourceGetResponse

	res, err := event.CloudController.Resources()
	if err != nil {
		logdog.Error("Unable to find resource", logdog.Fields{"error": err})
	}
	getResponse.Resource = res
	response.WriteEntity(getResponse)
}
