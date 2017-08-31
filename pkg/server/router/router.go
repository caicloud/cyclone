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
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/server/manager"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
	"github.com/zoumo/logdog"
)

const (
	// APIVersion is the version of API.
	APIVersion = "/api/v1"

	// projectPathParameterName represents the name of the path parameter for project.
	projectPathParameterName = "project"

	// pipelinePathParameterName represents the name of the path parameter for pipeline.
	pipelinePathParameterName = "pipeline"
)

// router represents the router to distribute the REST requests.
type router struct {
	// projectManager represents the project manager.
	projectManager manager.ProjectManager

	// pipelineManager represents the pipeline manager.
	pipelineManager manager.PipelineManager
}

// InitRouters initializes the router for REST APIs.
func InitRouters(dataStore *store.DataStore) error {
	// New pipeline manager
	pipelineManager, err := manager.NewPipelineManager(dataStore)
	if err != nil {
		return err
	}

	// New project manager
	projectManager, err := manager.NewProjectManager(dataStore, pipelineManager)
	if err != nil {
		return err
	}

	router := &router{
		projectManager,
		pipelineManager,
	}

	ws := new(restful.WebService)

	router.registerProjectAPIs(ws)
	router.registerPipelineAPIs(ws)

	restful.Add(ws)

	return nil
}

// registerProjectAPIs registers project related endpoints.
func (router *router) registerProjectAPIs(ws *restful.WebService) {
	// TODO (robin) Update the go-restful to support API tags.
	// projectTags := []string{"Project"}

	logdog.Info("Register project APIs")

	ws.Path(APIVersion).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	// POST /api/v1/projects
	ws.Route(ws.POST("/projects").To(router.createProject).
		Doc("Add a project").
		Reads(api.Project{}))

	// GET /api/v1/projects
	ws.Route(ws.GET("/projects").To(router.listProjects).
		Doc("Get all projects"))

	// PUT /api/v1/projects/{project}
	ws.Route(ws.PUT("/projects/{project}").To(router.updateProject).
		Doc("Update the project").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).Reads(api.Project{}))

	// GET /api/v1/projects/{project}
	ws.Route(ws.GET("/projects/{project}").To(router.getProject).
		Doc("Get the project").
		Param(ws.PathParameter("project", "name of the project").DataType("string")))

	// DELETE /api/v1/projects/{project}
	ws.Route(ws.DELETE("/projects/{project}").To(router.deleteProject).
		Doc("Delete the project").
		Param(ws.PathParameter("project", "name of the project").DataType("string")))
}

// registerPipelineAPIs registers pipeline related endpoints.
func (router *router) registerPipelineAPIs(ws *restful.WebService) {
	logdog.Info("Register pipeline APIs")

	ws.Path(APIVersion).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	// POST /api/v1/projects/{project}/pipelines
	ws.Route(ws.POST("/projects/{project}/pipelines").To(router.createPipeline).
		Doc("Add a pipeline").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Reads(api.Pipeline{}))

	// GET /api/v1/projects/{project}/pipelines
	ws.Route(ws.GET("/projects/{project}/pipelines").To(router.listPipelines).
		Doc("Get all pipelines").
		Param(ws.PathParameter("project", "name of the project").DataType("string")))

	// PUT /api/v1/projects/{project}/pipelines/{pipeline}
	ws.Route(ws.PUT("/projects/{project}/pipelines/{pipeline}").To(router.updatePipeline).
		Doc("Update the pipeline").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Reads(api.Pipeline{}))

	// GET /api/v1/projects/{project}/pipelines/{pipeline}
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}").To(router.getPipeline).
		Doc("Get the pipeline").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")))

	// DELETE /api/v1/projects/{project}/pipelines/{pipeline}
	ws.Route(ws.DELETE("/projects/{project}/pipelines/{pipeline}").To(router.deletePipeline).
		Doc("Delete a pipeline").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")))
}
