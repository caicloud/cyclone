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

package http

import (
	"net/http"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/emicklei/go-restful"
	"github.com/zoumo/logdog"
)

// ReadEntityFromRequest reads the entity from request body.
func ReadEntityFromRequest(request *restful.Request, response *restful.Response, entityPointer interface{}) error {
	if err := request.ReadEntity(entityPointer); err != nil {
		logdog.Errorf("Fail to read request entity as %s", err.Error())
		ResponseWithError(response, http.StatusBadRequest, err)
		return err
	}

	return nil
}

// ResponseWithError responses the request with error.
func ResponseWithError(response *restful.Response, statusCode int, err error) {
	errResp := api.ErrorResponse{Message: err.Error()}
	response.WriteHeaderAndEntity(statusCode, errResp)
}

// ResponseWithList responses the list request with metedata.
func ResponseWithList(list interface{}, total int, itemsLength int) api.ListResponse {
	return api.ListResponse{
		Meta: api.ListMeta{
			Total:       total,
			ItemsLength: itemsLength,
		},
		Items: list,
	}
}
