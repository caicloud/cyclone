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

	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

// CreateCloud handles the request to create a cloud.
func CreateCloud(ctx context.Context) (*api.Cloud, error) {
	cloud := &api.Cloud{}
	err := contextutil.GetJsonPayload(ctx, cloud)
	if err != nil {
		return nil, err
	}

	createdCloud, err := cloudManager.CreateCloud(cloud)
	if err != nil {
		return nil, err
	}

	return createdCloud, nil
}

// ListClouds handles the request to list all clouds.
func ListClouds(ctx context.Context) (api.ListResponse, error) {
	clouds, err := cloudManager.ListClouds()
	if err != nil {
		return api.ListResponse{}, nil
	}

	return httputil.ResponseWithList(clouds, len(clouds)), nil
}

// DeleteCloud handles the request to delete the cloud.
func DeleteCloud(ctx context.Context, name string) error {
	if err := cloudManager.DeleteCloud(name); err != nil {
		return err
	}

	return nil
}

// PingCloud handles the request to ping a cloud to check its health.
func PingCloud(ctx context.Context, name string) map[string]string {
	resp := make(map[string]string)
	err := cloudManager.PingCloud(name)
	if err != nil {
		resp["status"] = err.Error()
	} else {
		resp["status"] = "ok"
	}

	return resp
}

// ListWorkers handles the request to list all workers.
func ListWorkers(ctx context.Context, cloudName, namespace string) (api.ListResponse, error) {
	workers, err := cloudManager.ListWorkers(cloudName, namespace)
	if err != nil {
		log.Errorf("list worker error:%v", err)
		return api.ListResponse{}, err
	}

	return httputil.ResponseWithList(workers, len(workers)), nil
}
