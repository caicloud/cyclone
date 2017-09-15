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

	"strconv"

	"github.com/caicloud/cyclone/pkg/api"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
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
	switch err := err.(type) {
	case *httperror.Error:
		response.WriteHeaderAndEntity(err.Code, api.ErrorResponse{
			Message: err.Error(),
			Reason: err.Reason,
		})
	case error:
		response.WriteHeaderAndEntity(http.StatusInternalServerError, api.ErrorResponse{
			Message: err.Error(),
			Reason: httperror.ReasonInternal,
		})
	default:
		// should not come here
		logdog.Fatalf("%s is an unknown error type", err)
	}
}

// ResponseWithList responses list with metadata.
func ResponseWithList(list interface{}, itemsLength int, total int) api.ListResponse {
	return api.ListResponse{
		Meta: api.ListMeta{
			Total:       total,
			ItemsLength: itemsLength,
		},
		Items: list,
	}
}

// QueryParamsFromRequest reads the query params from request body.
func QueryParamsFromRequest(request *restful.Request) (qp api.QueryParams, err error) {
	limitStr := request.QueryParameter(api.Limit)
	startStr := request.QueryParameter(api.Start)

	if limitStr != "" {
		qp.Limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return qp, httperror.ErrorParamTypeError.Format(api.Limit, "number", "string")
		}
	}
	if startStr != "" {
		qp.Start, err = strconv.Atoi(startStr)
		if err != nil {
			return qp, httperror.ErrorParamTypeError.Format(api.Start, "number", "string")
		}
	}

	return qp, nil
}
