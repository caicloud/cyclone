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
	"github.com/emicklei/go-restful"
	log "github.com/golang/glog"

	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/server/manager"
	"github.com/caicloud/cyclone/pkg/store"
)

const (
	// APIVersion is the version of API.
	APIVersion = "/api/v1"

	// projectPathParameterName represents the name of the path parameter for project.
	projectPathParameterName = "project"

	// pipelinePathParameterName represents the name of the path parameter for pipeline.
	pipelinePathParameterName = "pipeline"

	// pipelineRecordPathParameterName represents the name of the path parameter for pipeline record.
	pipelineRecordPathParameterName = "recordId"

	// eventPathParameterName represents the name of the path parameter for event.
	eventPathParameterName = "eventId"
)

// router represents the router to distribute the REST requests.
type router struct {
	// projectManager represents the project manager.
	projectManager manager.ProjectManager

	// pipelineManager represents the pipeline manager.
	pipelineManager manager.PipelineManager

	// pipelineRecordManager represents the pipeline record manager.
	pipelineRecordManager manager.PipelineRecordManager

	// eventManager represents the event manager.
	eventManager manager.EventManager
}

// InitRouters initializes the router for REST APIs.
func InitRouters(dataStore *store.DataStore) error {
	// New pipeline record manager
	pipelineRecordManager, err := manager.NewPipelineRecordManager(dataStore)
	if err != nil {
		return err
	}

	// New pipeline manager
	pipelineManager, err := manager.NewPipelineManager(dataStore, pipelineRecordManager)
	if err != nil {
		return err
	}

	// New project manager
	projectManager, err := manager.NewProjectManager(dataStore, pipelineManager)
	if err != nil {
		return err
	}

	// New project manager
	eventManager, err := manager.NewEventManager(dataStore)
	if err != nil {
		return err
	}

	router := &router{
		projectManager,
		pipelineManager,
		pipelineRecordManager,
		eventManager,
	}

	ws := new(restful.WebService)

	router.registerProjectAPIs(ws)
	router.registerPipelineAPIs(ws)
	router.registerPipelineRecordAPIs(ws)
	router.registerEventAPIs(ws)

	restful.Add(ws)

	return nil
}

// registerProjectAPIs registers project related endpoints.
func (router *router) registerProjectAPIs(ws *restful.WebService) {
	// TODO (robin) Update the go-restful to support API tags.
	// projectTags := []string{"Project"}

	log.Info("Register project APIs")

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

	// GET /api/v1/projects/{project}/repos
	ws.Route(ws.GET("/projects/{project}/repos").To(router.listRepos).
		Doc("List accessible repos of the project").
		Param(ws.PathParameter("project", "name of the project").DataType("string")))

	// GET /api/v1/projects/{project}/branches
	ws.Route(ws.GET("/projects/{project}/branches").To(router.listBranches).
		Doc("List branches of the repo for the project").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.QueryParameter("repo", "the repo to list branches for").Required(true)))
}

// registerPipelineAPIs registers pipeline related endpoints.
func (router *router) registerPipelineAPIs(ws *restful.WebService) {
	log.Info("Register pipeline APIs")

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

// registerPipelineRecordAPIs registers pipeline record related endpoints.
func (router *router) registerPipelineRecordAPIs(ws *restful.WebService) {
	log.Info("Register pipeline record APIs")

	ws.Path(APIVersion).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	// POST /api/v1/projects/{project}/pipelines/{pipeline}/records
	ws.Route(ws.POST("/projects/{project}/pipelines/{pipeline}/records").To(router.createPipelineRecord).
		Doc("Perform pipeline, which will create a pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Reads(api.PipelinePerformParams{}))

	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records").To(router.listPipelineRecords).
		Doc("Get all pipeline records of one pipeline").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")))

	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records/{recordId}").To(router.getPipelineRecord).
		Doc("Get the pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordId", "id of the pipeline record").DataType("string")))

	// DELETE /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}
	ws.Route(ws.DELETE("/projects/{project}/pipelines/{pipeline}/records/{recordId}").To(router.deletePipelineRecord).
		Doc("Delete a pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordId", "id of the pipeline record").DataType("string")))

	// PATCH /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}/status
	ws.Route(ws.PATCH("/projects/{project}/pipelines/{pipeline}/records/{recordId}/status").To(router.updatePipelineRecordStatus).
		Doc("Update the status of pipeline record, only support to set the status as Aborted for running pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordId", "id of the pipeline record").DataType("string")))

	log.Info("Register pipeline records logs APIs")

	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordId}/logs
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records/{recordId}/logs").To(router.getPipelineRecordLogs).
		Doc("Get the pipeline record log").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordId", "id of the pipeline record").DataType("string")))

	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}/logstream
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records/{recordId}/logstream").
		To(router.getPipelineRecordLogStream).
		Doc("Get log stream of pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordID", "identifier of the pipeline record").DataType("string")))

	// TODO (robin) gorilla/websocket only supports GET method. This a workaround as this API is only used by workers,
	// but still need a better way.
	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordID}/stagelogstream
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records/{recordID}/stagelogstream").
		To(router.receivePipelineRecordLogStream).
		Doc("Receive log stream of pipeline record from worker").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordId", "id of the pipeline record").DataType("string")))
}

// registerEventAPIs registers event related endpoints.
func (router *router) registerEventAPIs(ws *restful.WebService) {
	log.Info("Register event APIs")

	ws.Path(APIVersion).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	// GET /api/v1/events/{eventID}
	ws.Route(ws.GET("/events/{eventId}").To(router.getEvent).
		Doc("Get event by id").
		Param(ws.PathParameter("eventId", "id of the event").DataType("string")))

	// PUT /api/v1/events/{eventID}
	ws.Route(ws.PUT("/events/{eventId}").To(router.setEvent).
		Doc("Set the event by id").
		Param(ws.PathParameter("eventId", "id of the event").DataType("string")))
}
