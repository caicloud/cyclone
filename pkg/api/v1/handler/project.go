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

	"github.com/caicloud/cyclone/pkg/api"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

// CreateProject handles the request to create a project.
func CreateProject(ctx context.Context, user string) (*api.Project, error) {
	project := &api.Project{}

	err := contextutil.GetJsonPayload(ctx, project)
	if err != nil {
		return nil, err
	}

	project.Owner = user

	createdProject, err := projectManager.CreateProject(project)
	if err != nil {
		return nil, err
	}

	return createdProject, nil
}

// GetProject handles the request to get a project.
func GetProject(ctx context.Context, name string) (*api.Project, error) {
	project, err := projectManager.GetProject(name)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// ListProjects handles the request to list projects.
func ListProjects(ctx context.Context) (api.ListResponse, error) {
	queryParams, err := httputil.QueryParamsFromContext(ctx)
	if err != nil {
		return api.ListResponse{}, err
	}

	projects, count, err := projectManager.ListProjects(queryParams)
	if err != nil {
		return api.ListResponse{}, err
	}

	return httputil.ResponseWithList(projects, count), nil
}

// UpdateProject handles the request to update a project.
func UpdateProject(ctx context.Context, name string) (*api.Project, error) {
	project := &api.Project{}
	err := contextutil.GetJsonPayload(ctx, project)
	if err != nil {
		return nil, err
	}

	updatedProject, err := projectManager.UpdateProject(name, project)
	if err != nil {
		return nil, err
	}

	return updatedProject, nil
}

// DeleteProject handles the request to delete a project.
func DeleteProject(ctx context.Context, name string) error {
	if err := projectManager.DeleteProject(name); err != nil {
		return err
	}

	return nil
}

// ListRepos handles the request to list repositories.
func ListRepos(ctx context.Context, name string) (api.ListResponse, error) {
	_, err := projectManager.GetProject(name)
	if err != nil {
		return api.ListResponse{}, err
	}

	repos, err := projectManager.ListRepos(name)
	if err != nil {
		return api.ListResponse{}, err
	}

	return httputil.ResponseWithList(repos, len(repos)), nil
}

// ListBranches handles the request to list branches for SCM repositories.
func ListBranches(ctx context.Context, name, repo string) (api.ListResponse, error) {
	_, err := projectManager.GetProject(name)
	if err != nil {
		return api.ListResponse{}, err
	}

	branches, err := projectManager.ListBranches(name, repo)
	if err != nil {
		return api.ListResponse{}, err
	}

	return httputil.ResponseWithList(branches, len(branches)), nil
}

// ListBranches handles the request to list branches for SCM repositories.
func ListTags(ctx context.Context, name, repo string) (api.ListResponse, error) {
	_, err := projectManager.GetProject(name)
	if err != nil {
		return api.ListResponse{}, err
	}

	tags, err := projectManager.ListTags(name, repo)
	if err != nil {
		return api.ListResponse{}, err
	}

	return httputil.ResponseWithList(tags, len(tags)), nil
}

// ListDockerfiles handles the request to list dockerfiles for SCM repositories.
func ListDockerfiles(ctx context.Context, name, repo string) (api.ListResponse, error) {
	_, err := projectManager.GetProject(name)
	if err != nil {
		return api.ListResponse{}, err
	}

	dockerfiles, err := projectManager.ListDockerfiles(name, repo)
	if err != nil {
		return api.ListResponse{}, err
	}

	return httputil.ResponseWithList(dockerfiles, len(dockerfiles)), nil
}

// GetTemplateType handles the request to get project type for SCM repositories.
func GetTemplateType(ctx context.Context, name, repo string) (*api.TemplateType, error) {
	templateType := &api.TemplateType{}
	_, err := projectManager.GetProject(name)
	if err != nil {
		return templateType, err
	}

	tt, err := projectManager.GetTemplateType(name, repo)
	if err != nil {
		return templateType, err
	}

	templateType.Type = tt

	return templateType, nil
}

// GetProjectStatistics handles the request to get a project's statistics.
func GetProjectStatistics(ctx context.Context, name, start, end string) (*api.PipelineStatusStats, error) {
	startTime, endTime, err := checkAndTransTimes(start, end)
	if err != nil {
		return nil, err
	}

	stats, err := projectManager.GetStatistics(name, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
