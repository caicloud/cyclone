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
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/caicloud/nirvana/service"
	"github.com/emicklei/go-restful"
	"github.com/zoumo/logdog"
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/cyclone/pkg/api"
	httperror "github.com/caicloud/cyclone/pkg/util/http/errors"
)

const (

	// APIVersion is the version of API.
	APIVersion = "/api/v1"

	// ProjectPathParameterName represents the name of the path parameter for project.
	ProjectPathParameterName = "project"

	// PipelinePathParameterName represents the name of the path parameter for pipeline.
	PipelinePathParameterName = "pipeline"

	// PipelineIDPathParameterName represents the name of the path parameter for pipeline id.
	PipelineIDPathParameterName = "pipelineid"

	// PipelineRecordPathParameterName represents the name of the path parameter for pipeline record.
	PipelineRecordPathParameterName = "recordid"

	// PipelineRecordStagePathParameterName represents the name of the query parameter for pipeline record stage.
	PipelineRecordStageQueryParameterName = "stage"

	// PipelineRecordTaskQueryParameterName represents the name of the query parameter for pipeline record task.
	PipelineRecordTaskQueryParameterName = "task"

	// PipelineRecordDownloadQueryParameter represents a download flag of the query parameter for pipeline record task.
	PipelineRecordDownloadQueryParameter = "download"

	// EventPathParameterName represents the name of the path parameter for event.
	EventPathParameterName = "eventid"

	// CloudPathParameterName represents the name of the path parameter for cloud.
	CloudPathParameterName = "cloud"

	// NamespaceQueryParameterName represents the k8s cluster namespce of the query parameter for cloud.
	NamespaceQueryParameter = "namespace"

	// RepoQueryParameterName represents the repo name of the query parameter.
	RepoQueryParameter = "repo"

	// StartTimeQueryParameter represents the query param start time.
	StartTimeQueryParameter string = "startTime"

	// EndTimeQueryParameter represents the query param end time.
	EndTimeQueryParameter string = "endTime"

	// HeaderUser represents the the key of user in request header.
	HeaderUser = "X-User"

	HEADER_ContentType = "Content-Type"
)

// GetHttpRequest gets request from context.
func GetHttpRequest(ctx context.Context) *http.Request {
	return service.HTTPContextFrom(ctx).Request()
}

// ReadEntityFromRequest reads the entity from request body.
func ReadEntityFromRequest(request *restful.Request, entityPointer interface{}) error {
	if err := request.ReadEntity(entityPointer); err != nil {
		logdog.Errorf("Fail to read request entity as %s", err.Error())
		return httperror.ErrorUnknownRequest.Format(err.Error())
	}

	return nil
}

// ResponseWithError responses the request with error.
func ResponseWithError(response *restful.Response, err error) {
	switch err := err.(type) {
	case *httperror.Error:
		response.WriteHeaderAndEntity(err.Code, api.ErrorResponse{
			Message: err.Error(),
			Reason:  err.Reason,
		})
	case error:
		response.WriteHeaderAndEntity(http.StatusInternalServerError, api.ErrorResponse{
			Message: err.Error(),
			Reason:  httperror.ReasonInternal,
		})
	default:
		// should not come here
		logdog.Fatalf("%s is an unknown error type", err)
	}
}

// ResponseWithList responses list with metadata.
func ResponseWithList(list interface{}, total int) api.ListResponse {
	return api.ListResponse{
		Meta: api.ListMeta{
			Total: total,
		},
		Items: list,
	}
}

// QueryParamsFromRequest reads the query params from request.
func QueryParamsFromRequest(request *restful.Request) (qp api.QueryParams, err error) {
	limitStr := request.QueryParameter(api.Limit)
	startStr := request.QueryParameter(api.Start)
	filterStr := request.QueryParameter(api.Filter)

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
	if filterStr != "" {
		// Support multiple conditions to filter, they are seperated with comma.
		conditions := strings.Split(filterStr, ",")
		filter := make(map[string]interface{})
		for _, c := range conditions {
			filterParts := strings.Split(c, "=")
			if len(filterParts) != 2 {
				return qp, httperror.ErrorValidationFailed.Format(api.Filter, "filter pattern is not correct")
			}

			if _, ok := filter[filterParts[0]]; ok {
				return qp, httperror.ErrorValidationFailed.Format(api.Filter, "filter pattern is not correct")
			}

			filter[filterParts[0]] = bson.M{"$regex": filterParts[1]}
		}

		qp.Filter = filter
	}

	return qp, nil
}

// QueryParamsFromContext reads the query params from context.
func QueryParamsFromContext(ctx context.Context) (qp api.QueryParams, err error) {
	request := GetHttpRequest(ctx)
	err = request.ParseForm()
	if err != nil {
		return
	}

	limitStr := request.Form.Get(api.Limit)
	startStr := request.Form.Get(api.Start)
	filterStr := request.Form.Get(api.Filter)

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
	if filterStr != "" {
		// Support multiple conditions to filter, they are seperated with comma.
		conditions := strings.Split(filterStr, ",")
		filter := make(map[string]interface{})
		for _, c := range conditions {
			filterParts := strings.Split(c, "=")
			if len(filterParts) != 2 {
				return qp, httperror.ErrorValidationFailed.Format(api.Filter, "filter pattern is not correct")
			}

			if _, ok := filter[filterParts[0]]; ok {
				return qp, httperror.ErrorValidationFailed.Format(api.Filter, "filter pattern is not correct")
			}

			filter[filterParts[0]] = bson.M{"$regex": filterParts[1]}
		}

		qp.Filter = filter
	}

	return qp, nil
}

// RecordCountQueryParamsFromRequest reads the query params of pipeline record count from request.
func RecordCountQueryParamsFromRequest(request *restful.Request) (recentCount, recentSuccessCount, recentFailedCount int, err error) {
	recentCountStr := request.QueryParameter(api.RecentPipelineRecordCount)
	recentSuccessCountStr := request.QueryParameter(api.RecentSuccessPipelineRecordCount)
	recentFailedCountStr := request.QueryParameter(api.RecentFailedPipelineRecordCount)

	if recentCountStr != "" {
		recentCount, err = strconv.Atoi(recentCountStr)
		if err != nil {
			return 0, 0, 0, httperror.ErrorParamTypeError.Format(api.RecentPipelineRecordCount, "number", "string")
		}
	}
	if recentSuccessCountStr != "" {
		recentSuccessCount, err = strconv.Atoi(recentSuccessCountStr)
		if err != nil {
			return 0, 0, 0, httperror.ErrorParamTypeError.Format(api.RecentSuccessPipelineRecordCount, "number", "string")
		}
	}
	if recentFailedCountStr != "" {
		recentFailedCount, err = strconv.Atoi(recentFailedCountStr)
		if err != nil {
			return 0, 0, 0, httperror.ErrorParamTypeError.Format(api.RecentFailedPipelineRecordCount, "number", "string")
		}
	}

	return
}

// DownloadQueryParamsFromRequest reads the query param whether download pipeline record logs from request.
func DownloadQueryParamsFromRequest(request *restful.Request) (bool, error) {
	downloadStr := request.QueryParameter(api.Download)

	if downloadStr != "" {
		download, err := strconv.ParseBool(downloadStr)
		if err != nil {
			logdog.Errorf("Download param's value is %s", downloadStr)
			return false, httperror.ErrorParamTypeError.Format(api.Download, "bool", "string")
		}

		return download, nil
	}

	return false, nil
}
