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

	"github.com/caicloud/cyclone/pkg/api"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	"github.com/emicklei/go-restful"
)

// createProject handles the request to create a project.
func (router *router) createProject(request *restful.Request, response *restful.Response) {
	project := &api.Project{}
	if err := httputil.ReadEntityFromRequest(request, response, project); err != nil {
		return
	}

	createdProject, err := router.projectManager.CreateProject(project)
	if err != nil {
		httputil.ResponseWithError(response, http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, createdProject)
}

// getProject handles the request to get a project.
func (router *router) getProject(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)

	project, err := router.projectManager.GetProject(name)
	if err != nil {
		httputil.ResponseWithError(response, http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, project)
}

// listProjects handles the request to list projects.
func (router *router) listProjects(request *restful.Request, response *restful.Response) {
	queryParams := httputil.QueryParamsFromRequest(request)
	projects, count, err := router.projectManager.ListProjects(queryParams)
	if err != nil {
		httputil.ResponseWithError(response, http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, httputil.ResponseWithList(projects, len(projects), count))
}

// updateProject handles the request to update a project.
func (router *router) updateProject(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)
	project := &api.Project{}
	if err := httputil.ReadEntityFromRequest(request, response, project); err != nil {
		return
	}

	updatedProject, err := router.projectManager.UpdateProject(name, project)
	if err != nil {
		httputil.ResponseWithError(response, http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusOK, updatedProject)
}

// deleteProject handles the request to delete a project.
func (router *router) deleteProject(request *restful.Request, response *restful.Response) {
	name := request.PathParameter(projectPathParameterName)

	if err := router.projectManager.DeleteProject(name); err != nil {
		httputil.ResponseWithError(response, http.StatusInternalServerError, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusNoContent, nil)
}
