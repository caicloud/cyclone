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
	"time"

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

	// pipelineIDPathParameterName represents the name of the path parameter for pipeline id.
	pipelineIDPathParameterName = "pipelineid"

	// pipelineRecordPathParameterName represents the name of the path parameter for pipeline record.
	pipelineRecordPathParameterName = "recordid"

	// pipelineRecordStagePathParameterName represents the name of the query parameter for pipeline record stage.
	pipelineRecordStageQueryParameterName = "stage"

	// pipelineRecordTaskQueryParameterName represents the name of the query parameter for pipeline record task.
	pipelineRecordTaskQueryParameterName = "task"

	// eventPathParameterName represents the name of the path parameter for event.
	eventPathParameterName = "eventid"

	// cloudPathParameterName represents the name of the path parameter for cloud.
	cloudPathParameterName = "cloud"

	// namespaceQueryParameterName represents the k8s cluster namespce of the query parameter for cloud.
	namespaceQueryParameterName = "namespace"
)

// router represents the router to distribute the REST requests.
type router struct {
	// dataStore represents the manager for data store.
	dataStore *store.DataStore

	// projectManager represents the project manager.
	projectManager manager.ProjectManager

	// pipelineManager represents the pipeline manager.
	pipelineManager manager.PipelineManager

	// pipelineRecordManager represents the pipeline record manager.
	pipelineRecordManager manager.PipelineRecordManager

	// eventManager represents the event manager.
	eventManager manager.EventManager

	// cloudManager represents the cloud manager.
	cloudManager manager.CloudManager
}

// InitRouters initializes the router for REST APIs.
func InitRouters(dataStore *store.DataStore, recordRotationThreshold int) error {
	// New pipeline record manager
	pipelineRecordManager, err := manager.NewPipelineRecordManager(dataStore, recordRotationThreshold)
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

	// New event manager
	eventManager, err := manager.NewEventManager(dataStore)
	if err != nil {
		return err
	}

	// New cloud manager
	cloudManager, err := manager.NewCloudManager(dataStore)
	if err != nil {
		return err
	}

	router := &router{
		dataStore,
		projectManager,
		pipelineManager,
		pipelineRecordManager,
		eventManager,
		cloudManager,
	}

	ws := new(restful.WebService)
	ws.Filter(NCSACommonLogFormatLogger())

	router.registerProjectAPIs(ws)
	router.registerPipelineAPIs(ws)
	router.registerPipelineRecordAPIs(ws)
	router.registerEventAPIs(ws)
	router.registerCloudAPIs(ws)
	router.registerHealthCheckAPI(ws)
	router.registerWebhookAPIs(ws)
	router.registerOauthAPI(ws)
	restful.Add(ws)

	return nil
}

// NCSACommonLogFormatLogger add filter to produce NCSA standard log.
func NCSACommonLogFormatLogger() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		t := time.Now()

		chain.ProcessFilter(req, resp)
		log.Infof("%d \"%s %s %s\" - %s %d %.2fs\n",
			resp.StatusCode(),
			req.Request.Method,
			req.Request.URL.RequestURI(),
			req.Request.Proto,
			req.Request.RemoteAddr,
			resp.ContentLength(),
			time.Since(t).Seconds(),
		)
	}
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

	// GET /api/v1/projects/{project}/tags
	ws.Route(ws.GET("/projects/{project}/tags").To(router.listTags).
		Doc("List tags of the repo for the project").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.QueryParameter("repo", "the repo to list branches for").Required(true)))

	// GET /api/v1/projects/{project}/stats
	ws.Route(ws.GET("/projects/{project}/stats").To(router.getProjectStatistics).
		Doc("Get statistics of the project").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.QueryParameter("startTime", "the start time of statistics").Required(false)).
		Param(ws.QueryParameter("endTime", "the end time of statistics").Required(false)))
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

	// GET /api/v1/projects/{projects}/pipelines/{pipeline}/stats
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/stats").To(router.getPipelineStatistics).
		Doc("Get statistics of the pipeline").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.QueryParameter("startTime", "the start time of statistics").Required(false)).
		Param(ws.QueryParameter("endTime", "the end time of statistics").Required(false)))
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

	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records/{recordid}").To(router.getPipelineRecord).
		Doc("Get the pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordid", "id of the pipeline record").DataType("string")))

	// DELETE /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}
	ws.Route(ws.DELETE("/projects/{project}/pipelines/{pipeline}/records/{recordid}").To(router.deletePipelineRecord).
		Doc("Delete a pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordid", "id of the pipeline record").DataType("string")))

	// PATCH /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/status
	ws.Route(ws.PATCH("/projects/{project}/pipelines/{pipeline}/records/{recordid}/status").To(router.updatePipelineRecordStatus).
		Doc("Update the status of pipeline record, only support to set the status as Aborted for running pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordid", "id of the pipeline record").DataType("string")))

	log.Info("Register pipeline records logs APIs")

	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/logs
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records/{recordid}/logs").To(router.getPipelineRecordLogs).
		Doc("Get the pipeline record log").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordid", "id of the pipeline record").DataType("string")))

	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/logstream
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records/{recordid}/logstream").
		To(router.getPipelineRecordLogStream).
		Doc("Get log stream of pipeline record").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordid", "identifier of the pipeline record").DataType("string")))

	// TODO (robin) gorilla/websocket only supports GET method. This a workaround as this API is only used by workers,
	// but still need a better way.
	// GET /api/v1/projects/{project}/pipelines/{pipeline}/records/{recordid}/stagelogstream
	ws.Route(ws.GET("/projects/{project}/pipelines/{pipeline}/records/{recordid}/stagelogstream").
		To(router.receivePipelineRecordLogStream).
		Doc("Receive log stream of pipeline record from worker").
		Param(ws.PathParameter("project", "name of the project").DataType("string")).
		Param(ws.PathParameter("pipeline", "name of the pipeline").DataType("string")).
		Param(ws.PathParameter("recordid", "id of the pipeline record").DataType("string")))
}

// registerEventAPIs registers event related endpoints.
func (router *router) registerEventAPIs(ws *restful.WebService) {
	log.Info("Register event APIs")

	ws.Path(APIVersion).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	// GET /api/v1/events/{eventid}
	ws.Route(ws.GET("/events/{eventid}").To(router.getEvent).
		Doc("Get event by id").
		Param(ws.PathParameter("eventid", "id of the event").DataType("string")))

	// PUT /api/v1/events/{eventid}
	ws.Route(ws.PUT("/events/{eventid}").To(router.setEvent).
		Doc("Set the event by id").
		Param(ws.PathParameter("eventid", "id of the event").DataType("string")))
}

// registerWebhookAPIs registers webhook related endpoints.
func (router *router) registerWebhookAPIs(ws *restful.WebService) {
	log.Info("Register webhook APIs")

	ws.Path(APIVersion).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	// POST /api/v1/pipelines/{pipelineid}/githubwebhook
	ws.Route(ws.POST("/pipelines/{pipelineid}/githubwebhook").To(router.handleGithubWebhook).
		Doc("Trigger the pipeline by github webhook").
		Param(ws.PathParameter("pipelineid", "id of the pipeline").DataType("string")))

	// POST /api/v1/pipelines/{pipelineid}/gitlabwebhook
	ws.Route(ws.POST("/pipelines/{pipelineid}/gitlabwebhook").To(router.handleGitlabWebhook).
		Doc("Trigger the pipeline by gitlab webhook").
		Param(ws.PathParameter("pipelineid", "id of the pipeline").DataType("string")))
}

// registerCloudAPIs registers cloud related endpoints.
func (router *router) registerCloudAPIs(ws *restful.WebService) {
	log.Info("Register cloud APIs")

	ws.Path(APIVersion).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)
	// POST /api/v1/clouds
	ws.Route(ws.POST("/clouds").To(router.createCloud).
		Doc("Add a cloud").
		Reads(api.Cloud{}))

	// GET /api/v1/clouds
	ws.Route(ws.GET("/clouds").To(router.listClouds).
		Doc("Get all clouds"))

	// DELETE /api/v1/clouds/{cloud}
	ws.Route(ws.DELETE("/clouds/{cloud}").To(router.deleteCloud).
		Doc("Delete the cloud").
		Param(ws.PathParameter("cloud", "name of the cloud").DataType("string")))

	// GET /api/v1/clouds/{cloud}/ping
	ws.Route(ws.GET("/clouds/{cloud}/ping").To(router.pingCloud).
		Doc("Ping the cloud to check its health").
		Param(ws.PathParameter("cloud", "name of the cloud").DataType("string")))

	// GET /api/v1/clouds/{cloud}/workers
	ws.Route(ws.GET("/clouds/{cloud}/workers").To(router.listWorkers).
		Doc("Get all cyclone workers in the cloud").
		Param(ws.PathParameter(cloudPathParameterName, "name of the cloud").DataType("string")).
		Param(ws.QueryParameter(namespaceQueryParameterName, "namespace of kubernetes cluster").DataType("string")))
}

// registerHealthCheckAPI registers health check API.
func (router *router) registerHealthCheckAPI(ws *restful.WebService) {
	log.Info("Register health check API")

	// GET /api/v1/healthcheck
	ws.Route(ws.GET("/healthcheck").To(router.healthCheck).
		Doc("Health check for Cyclone server"))
}

// registerOauthAPI registers gitlab oauth API.
func (router *router) registerOauthAPI(ws *restful.WebService) {
	log.Info("Register gitlab oauth API")

	ws.Path(APIVersion).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	// GET /api/v1/scm/{type}/code
	ws.Route(ws.GET("/scm/{type}/code").To(router.getOauthCode).
		Doc("Get gitlab oauth code").
		Param(ws.PathParameter("type", "type of scm").DataType("string")))

	// GET /api/v1/scm/{type}/callback
	ws.Route(ws.GET("/scm/{type}/callback").To(router.getOauthToken).
		Doc("Oauth callback to access token").
		Param(ws.PathParameter("type", "type of scm").DataType("string")).
		Param(ws.QueryParameter("code", "result of first oauth request")).
		Param(ws.QueryParameter("state", "the random string to avoid csrf")))
}
