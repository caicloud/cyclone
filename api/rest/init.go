/*
Copyright 2016 caicloud authors. All rights reserved.

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

package rest

import (
	"fmt"
	"path/filepath"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/remote"
	"github.com/emicklei/go-restful"
)

var (
	// remoteManager is remote api manager.
	remoteManager *remote.Manager
	// resourceManager *resource.Manager
)

// Initialize initializes rest endpoints and all Cyclone managers. It register REST
// APIs to restful.WebService, and creates a global remote manager.
func Initialize() {

	// Register all rest endpoints.
	ws := &restful.WebService{}
	ws.Path(fmt.Sprintf("/api/%s", api.APIVersion)).
		Consumes(restful.MIME_JSON, "text/plain", "text/event-stream").
		Produces(restful.MIME_JSON, "text/plain", "text/event-stream")

	// Register APIs to the web service.
	registerWebhookAPIs(ws)
	registerHealthCheckAPIs(ws)
	registerServiceAPIs(ws)
	registerEventAPIs(ws)
	registerVersionAPIs(ws)
	// TODO Register project API.
	//registerProjectAPIs(ws)
	registerRemoteAPIs(ws)
	registerVersionLogAPIs(ws)
	registerResourceAPIs(ws)
	registerDeployAPIs(ws)

	registerTemplateAPIs(ws)
	registerCloudsAPIs(ws)

	restful.Add(ws)

	// Add container filter to enable CORS and respond to OPTIONS.
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders: []string{"X-My-Header"},
		// The header "token" is not a standard header. It's used in auth server
		// to handle authentication.
		AllowedHeaders: []string{"Content-Type", "Accept", "token"},
		CookiesAllowed: false,
		Container:      restful.DefaultContainer,
	}
	restful.Filter(cors.Filter)
	restful.Filter(restful.OPTIONSFilter())

	remoteManager = remote.NewManager()
}

// GetManager gets a remote manager.
func GetManager() *remote.Manager {
	if nil == remoteManager {
		remoteManager = remote.NewManager()
	}

	return remoteManager
}

// registerHealthCheckAPIs registers health check related endpoints.
func registerHealthCheckAPIs(ws *restful.WebService) {
	ws.Route(ws.GET("/healthcheck").
		To(healthCheck).
		Doc("health check").
		Writes(api.HealthCheckResponse{}))
}

// registerServiceAPIs registers service related endpoints.
func registerServiceAPIs(ws *restful.WebService) {
	ws.Route(ws.POST("/{user_id}/services").
		To(createService).
		Doc("create a service for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Reads(api.Service{}).
		Writes(api.ServiceCreationResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.GET("/{user_id}/services/{service_id}").
		Filter(checkACLForService).
		To(getService).
		Doc("find a service by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("service_id", "identifier of the service").DataType("string")).
		Writes(api.ServiceGetResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.GET("/{user_id}/services/{service_id}/badge").
		To(getServiceBadge).
		Doc("get the badge for given service").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("service_id", "identifier of the service").DataType("string")))

	ws.Route(ws.GET("/{user_id}/services").
		To(listServices).
		Doc("list all services of a given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Writes([]api.ServiceListResponse{}))

	ws.Route(ws.DELETE("/{user_id}/services/{service_id}").
		Filter(checkACLForService).
		To(deleteService).
		Doc("delete a service by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("service_id", "identifier of the service").DataType("string")).
		Writes(api.ServiceDelResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.PUT("/{user_id}/services/{service_id}").
		Filter(checkACLForService).
		To(setService).
		Doc("set a service by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("service_id", "identifier of the service").DataType("string")).
		Writes(api.ServiceSetResponse{}))
}

// registerVersionAPIs registers version related endpoints.
func registerVersionAPIs(ws *restful.WebService) {
	ws.Route(ws.POST("/{user_id}/versions").
		To(createVersion).
		Doc("create a version for given user of a specific service").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Reads(api.Version{}).
		Writes(api.VersionCreationResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.GET("/{user_id}/versions/{version_id}").
		Filter(checkACLForVersion).
		To(getVersion).
		Doc("find a version by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("version_id", "identifier of the version").DataType("string")).
		Writes(api.Version{}))

	ws.Route(ws.GET("/{user_id}/services/{service_id}/versions").
		Filter(checkACLForService).
		To(listVersions).
		Doc("list all versions of a given user and service").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("service_id", "identifier of the service").DataType("string")).
		Writes([]api.VersionListResponse{}))

	ws.Route(ws.POST("/{user_id}/versions/{version_id}/cancelbuild").
		Filter(checkACLForVersion).
		To(cancelVersion).
		Doc("cancel a version by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("version_id", "identifier of the version").DataType("string")).
		Writes(api.VersionConcelResponse{}))

}

// registerEventAPIs registers event related endpoints.
func registerEventAPIs(ws *restful.WebService) {
	ws.Route(ws.GET("/events/{event_id}").
		To(getEvent).
		Doc("get a event info").
		Param(ws.PathParameter("event_id", "identifier of the event").DataType("string")).
		Writes(api.GetEventResponse{}))

	ws.Route(ws.PUT("/events/{event_id}").
		To(setEvent).
		Doc("set a event info").
		Param(ws.PathParameter("event_id", "identifier of the event").DataType("string")).
		Reads(api.SetEvent{}).
		Writes(api.SetEventResponse{}))
}

// registerResourceAPIs registers resource related endpoints.
func registerRemoteAPIs(ws *restful.WebService) {
	ws.Route(ws.GET("/{user_id}/remotes/{code_repository}/requesttoken").To(requesttoken).
		Doc("request token for private repository of a given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("code_repository", "identifier of the code repository").DataType("string")).
		Writes(api.RequestTokenResponse{}))

	ws.Route(ws.GET("/remotes/{code_repository}/authcallback").To(authcallback).
		Doc("auth callback by code_repository").
		Param(ws.PathParameter("code_repository", "identifier of the code repository").DataType("string")).
		Writes(api.AuthCallbackResponse{}))

	ws.Route(ws.GET("/{user_id}/remotes/{code_repository}/listrepo").To(listrepo).
		Doc("list repo by using token").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("code_repository", "identifier of the code repository").DataType("string")).
		Writes(api.ListRepoResponse{}))

	ws.Route(ws.GET("/{user_id}/remotes/{code_repository}/logout").To(logout).
		Doc("logout").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("code_repository", "identifier of the code repository").DataType("string")).
		Writes(api.LogoutResponse{}))
}

// registerWebhookAPIs registers webhook related endpoints.
func registerWebhookAPIs(ws *restful.WebService) {
	ws.Route(ws.POST("/{service_id}/webhook_github").
		To(webhookGithub).
		Doc("webhook from github").
		Param(ws.PathParameter("service_id", "identifier of the service").DataType("string")).
		Reads(api.WebhookGithub{}).
		Writes(api.WebhookResponse{}))

	ws.Route(ws.POST("/{service_id}/webhook_gitlab").
		To(webhookGitLab).
		Doc("webhook from gitlab").
		Param(ws.PathParameter("service_id", "identifier of the service").DataType("string")).
		Reads(api.WebhookGitlab{}).
		Writes(api.WebhookResponse{}))

	ws.Route(ws.POST("/{service_id}/webhook_svn").
		To(webhookSVN).
		Doc("webhook from svn").
		Param(ws.PathParameter("service_id", "identifier of the service").DataType("string")).
		Reads(api.WebhookSVN{}).
		Writes(api.WebhookResponse{}))
}

// registerResourceAPIs registers resource related endpoints.
func registerResourceAPIs(ws *restful.WebService) {
	// Filter the unauthorized operation.
	// ws.Route(ws.PUT("/{user_id}/resources").
	// 	To(setResource).
	// 	Doc("set a resource by id for given user").
	// 	Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
	// 	Writes(api.ResourceSetResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.GET("/{user_id}/resources").
		To(getResource).
		Doc("find a service by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Writes(api.ResourceGetResponse{}))
}

// registerVersionLogAPIs registers log related endpoints.
func registerVersionLogAPIs(ws *restful.WebService) {
	ws.Route(ws.GET("/{user_id}/versions/{version_id}/logs").
		Filter(checkACLForVersion).
		To(getVersionLog).
		Doc("find version log by given version id").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("version_id", "identifier of the version").DataType("string")).
		Writes(api.VersionLogGetResponse{}))

	// Notive: If you modify here, you also need to update the code in worker/helper/output.go.
	ws.Route(ws.POST("/{user_id}/versions/{version_id}/logs").
		Filter(checkACLForVersion).
		To(createVersionLog).
		Doc("createVersionLog creates an version log").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("version_id", "identifier of the version").DataType("string")).
		Writes(api.VersionLogCreateResponse{}))
}

// registerDeployAPIs registers deploy related endpoints.
func registerDeployAPIs(ws *restful.WebService) {
	ws.Route(ws.POST("/{user_id}/deploys").
		To(createDeploy).
		Doc("create a deploy for given user of a specific service").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Reads(api.Deploy{}).
		Writes(api.DeployCreationResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.GET("/{user_id}/deploys/{deploy_id}").
		To(getDeploy).
		Doc("find a deploy by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("deploy_id", "identifier of the deploy").DataType("string")).
		Writes(api.DeployGetResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.GET("/{user_id}/deploys").
		To(listDeploy).
		Doc("find deploys by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Writes(api.DeployListResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.PUT("/{user_id}/deploys/{deploy_id}").
		To(setDeploy).
		Doc("set a deploy by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("deploy_id", "identifier of the deploy").DataType("string")).
		Writes(api.DeploySetResponse{}))

	// Filter the unauthorized operation.
	ws.Route(ws.DELETE("/{user_id}/deploys/{deploy_id}").
		To(delDeploy).
		Doc("delete a deploy by id for given user").
		Param(ws.PathParameter("user_id", "identifier of the user").DataType("string")).
		Param(ws.PathParameter("deploy_id", "identifier of the deploy").DataType("string")).
		Writes(api.DeployDelResponse{}))
}

// registerWebhookAPIs registers templates api
func registerTemplateAPIs(ws *restful.WebService) {

	err := filepath.Walk(DockerfilePath, walkDockerfiles)

	if err != nil {
		log.WarnWithFields("error occur when walk dockerfile path, ", log.Fields{"path": DockerfilePath, "err": err})
	}

	err = filepath.Walk(YamlPath, walkYamlfiles)

	if err != nil {
		log.WarnWithFields("error occur when walk yamlfile path, ", log.Fields{"path": YamlPath, "err": err})
	}

	ws.Route(ws.GET("/templates/yamls").
		To(listYamlfiles).
		Doc("list all yaml templates"))

	ws.Route(ws.GET("/templates/yamls/{yamlfile}").
		To(getYamlfile).
		Doc("get one yaml template").
		Param(ws.PathParameter("yamlfile", "yaml file name").DataType("string")))

	ws.Route(ws.GET("/templates/dockerfiles").
		To(listDockerfiles).
		Doc("list all docekrfile templates"))

	ws.Route(ws.GET("/templates/dockerfiles/{dockerfile}").
		To(getDockerfile).
		Doc("get one docekrfile template").
		Param(ws.PathParameter("dockerfile", "dockerfile name").DataType("string")))

}

// registerCloudsAPIs ...
func registerCloudsAPIs(ws *restful.WebService) {
	ws.Route(ws.GET("/clouds").
		To(listCloud).
		Doc("list all clouds"))

	ws.Route(ws.POST("/clouds").
		To(createCloud).
		Doc("add a cloud to cyclone"))

	ws.Route(ws.DELETE("/clouds/{cloudName}").
		To(deleteCloud).
		Doc("add a cloud to cyclone").
		Param(ws.PathParameter("cloudName", "cloud name").DataType("string")))

	ws.Route(ws.GET("/clouds/{cloudName}/ping").
		To(pingCloud).
		Doc("test a cloud ping").
		Param(ws.PathParameter("cloudName", "cloud name").DataType("string")))

}
