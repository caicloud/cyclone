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
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/caicloud/cyclone/pkg/api"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
)

// createProject handles the request to create a project.
func (router *router) createProject(request *restful.Request, response *restful.Response) {
	project := &api.Project{}
	if err := httputil.ReadEntityFromRequest(request, project); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	project.Owner = request.Request.Header.Get(httputil.HeaderUser)

	createdProject, err := router.projectManager.CreateProject(project)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, createdProject)
}

// getProject handles the request to get a project.
func (router *router) getProject(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)

	project, err := router.projectManager.GetProject(name)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, project)
}

// listProjects handles the request to list projects.
func (router *router) listProjects(request *restful.Request, response *restful.Response) {
	queryParams, err := httputil.QueryParamsFromRequest(request)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	projects, count, err := router.projectManager.ListProjects(queryParams)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(projects, count))
}

// updateProject handles the request to update a project.
func (router *router) updateProject(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)
	project := &api.Project{}
	if err := httputil.ReadEntityFromRequest(request, project); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	updatedProject, err := router.projectManager.UpdateProject(name, project)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, updatedProject)
}

// deleteProject handles the request to delete a project.
func (router *router) deleteProject(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)

	if err := router.projectManager.DeleteProject(name); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusNoContent, nil)
}

// listRepos handles the request to list repositories.
func (router *router) listRepos(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)

	_, err := router.projectManager.GetProject(name)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	repos, err := router.projectManager.ListRepos(name)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(repos, len(repos)))
}

// listBranches handles the request to list branches for SCM repositories.
func (router *router) listBranches(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)
	repo := request.QueryParameter(api.Repo)

	_, err := router.projectManager.GetProject(name)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	branches, err := router.projectManager.ListBranches(name, repo)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(branches, len(branches)))
}

// listBranches handles the request to list branches for SCM repositories.
func (router *router) listTags(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)
	repo := request.QueryParameter(api.Repo)

	_, err := router.projectManager.GetProject(name)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	tags, err := router.projectManager.ListTags(name, repo)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(tags, len(tags)))
}

// getRepoType handles the request to get project type for SCM repositories.
func (router *router) getRepoType(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)
	repo := request.QueryParameter(api.Repo)

	_, err := router.projectManager.GetProject(name)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	repotype, err := router.projectManager.GetRepoType(name, repo)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	repoType := struct {
		Type string `json:"type"`
	}{
		Type: repotype,
	}

	response.WriteHeaderAndEntity(http.StatusOK, repoType)
}

// getProjectStatistics handles the request to get a project's statistics.
func (router *router) getProjectStatistics(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter(projectPathParameterName)
	start := request.QueryParameter(api.StartTime)
	end := request.QueryParameter(api.EndTime)

	startTime, endTime, err := checkAndTransTimes(start, end)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	stats, err := router.projectManager.GetStatistics(projectName, startTime, endTime)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, stats)
}
