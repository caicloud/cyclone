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

package handler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/caicloud/cyclone/pkg/api"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

// CreatePipeline handles the request to create a pipeline.
func CreatePipeline(ctx context.Context, projectName, user string) (*api.Pipeline, error) {
	project, err := projectManager.GetProject(projectName)
	if err != nil {
		return nil, err
	}

	pipeline := &api.Pipeline{}
	err = contextutil.GetJsonPayload(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	pipeline.Owner = user
	pipeline.ProjectID = project.ID
	createdPipeline, err := pipelineManager.CreatePipeline(projectName, pipeline)
	if err != nil {
		return nil, err
	}

	return createdPipeline, nil
}

// GetPipeline handles the request to get a pipeline.
func GetPipeline(ctx context.Context, projectName, pipelineName string, recentCount, recentSuccessCount, recentFailedCount int) (*api.Pipeline, error) {

	//recentCount, recentSuccessCount, recentFailedCount, err := httputil.RecordCountQueryParamsFromRequest(request)
	//if err != nil {
	//	httputil.ResponseWithError(response, err)
	//	return
	//}

	pipeline, err := pipelineManager.GetPipeline(projectName, pipelineName, recentCount, recentSuccessCount, recentFailedCount)
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

// ListPipelines handles the request to list pipelines.
func ListPipelines(ctx context.Context, projectName string, recentCount, recentSuccessCount, recentFailedCount int) (api.ListResponse, error) {
	queryParams, err := httputil.QueryParamsFromContext(ctx)
	if err != nil {
		return api.ListResponse{}, nil
	}
	//recentCount, recentSuccessCount, recentFailedCount, err := httputil.RecordCountQueryParamsFromRequest(request)
	//if err != nil {
	//	httputil.ResponseWithError(response, err)
	//	return
	//}

	pipelines, count, err := pipelineManager.ListPipelines(projectName, queryParams, recentCount, recentSuccessCount, recentFailedCount)
	if err != nil {
		return api.ListResponse{}, nil
	}

	return httputil.ResponseWithList(pipelines, count), nil
}

// UpdatePipeline handles the request to update a pipeline.
func UpdatePipeline(ctx context.Context, projectName, pipelineName string) (*api.Pipeline, error) {
	pipeline := &api.Pipeline{}
	err := contextutil.GetJsonPayload(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	updatedPipeline, err := pipelineManager.UpdatePipeline(projectName, pipelineName, pipeline)
	if err != nil {
		return nil, err
	}

	return updatedPipeline, nil
}

// DeletePipeline handles the request to delete a pipeline.
func DeletePipeline(ctx context.Context, projectName, pipelineName string) error {
	if err := pipelineManager.DeletePipeline(projectName, pipelineName); err != nil {
		return err
	}

	return nil
}

// GetPipelineStatistics handles the request to get a pipeline's statistics.
func GetPipelineStatistics(ctx context.Context, projectName, pipelineName string, start, end string) (*api.PipelineStatusStats, error) {
	startTime, endTime, err := checkAndTransTimes(start, end)
	if err != nil {
		return nil, err
	}

	stats, err := pipelineManager.GetStatistics(projectName, pipelineName, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func checkAndTransTimes(start, end string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time
	if start == "" || end == "" {
		err := fmt.Errorf("query parameters `startTime` and `endTime` can not be empty.")
		return startTime, endTime, err
	}

	startTime, endTime, err := transTimes(start, end)
	if err != nil {
		err := fmt.Errorf("query parameters `startTime` and `endTime` must be int positive integer.")
		return startTime, endTime, err
	}

	if startTime.After(endTime) {
		err := fmt.Errorf("query parameters `startTime` must less or equal than `endTime`.")
		return startTime, endTime, err
	}
	return startTime, endTime, nil
}

// transTimes trans startTime and endTime from string to time.Time.
func transTimes(start, end string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time

	startInt, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		return startTime, endTime, err
	}
	startTime = time.Unix(startInt, 0)

	endInt, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		return startTime, endTime, err
	}
	endTime = time.Unix(endInt, 0)

	return startTime, endTime, nil
}
