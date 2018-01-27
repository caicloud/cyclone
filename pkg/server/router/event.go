/*
Copyright 2017 caicloud authors. All rights reserved.

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

package router

import (
	"fmt"
	"net/http"

	"github.com/caicloud/cyclone/pkg/api"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	restful "github.com/emicklei/go-restful"
)

// getEvent handles the request to get a event.
func (router *router) getEvent(request *restful.Request, response *restful.Response) {
	eventID := request.PathParameter(eventPathParameterName)

	event, err := router.eventManager.GetEvent(eventID)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, event)
}

// setEvent handles the request to set the event.
func (router *router) setEvent(request *restful.Request, response *restful.Response) {
	eventID := request.PathParameter(eventPathParameterName)
	event := &api.Event{}
	if err := httputil.ReadEntityFromRequest(request, event); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}
	if eventID != event.ID {
		err := fmt.Errorf("The event IDs in the request path and request body are not same")
		httputil.ResponseWithError(response, err)
		return
	}

	event, err := router.eventManager.SetEvent(event)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, event)
}
