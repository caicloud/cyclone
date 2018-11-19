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

	// FileNamePathParameterName represents the name of the path parameter for file name.
	FileNamePathParameterName = "filename"

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

	// SVNRepoIDPathParameterName represents a svn repository's uuid.
	SVNRepoIDPathParameterName = "svnrepoid"

	// SVNRevisionQueryParameterName represents the svn commit revision.
	SVNRevisionQueryParameterName = "revision"
)

// GetHttpRequest gets request from context.
func GetHttpRequest(ctx context.Context) *http.Request {
	return service.HTTPContextFrom(ctx).Request()
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
			return qp, httperror.ErrorParamTypeError.Error(api.Limit, "number", "string")
		}
	}
	if startStr != "" {
		qp.Start, err = strconv.Atoi(startStr)
		if err != nil {
			return qp, httperror.ErrorParamTypeError.Error(api.Start, "number", "string")
		}
	}
	if filterStr != "" {
		// Support multiple conditions to filter, they are seperated with comma.
		conditions := strings.Split(filterStr, ",")
		filter := make(map[string]interface{})
		for _, c := range conditions {
			filterParts := strings.Split(c, "=")
			if len(filterParts) != 2 {
				return qp, httperror.ErrorValidationFailed.Error(api.Filter, "filter pattern is not correct")
			}

			if _, ok := filter[filterParts[0]]; ok {
				return qp, httperror.ErrorValidationFailed.Error(api.Filter, "filter pattern is not correct")
			}

			filter[filterParts[0]] = bson.M{"$regex": filterParts[1]}
		}

		qp.Filter = filter
	}

	return qp, nil
}
