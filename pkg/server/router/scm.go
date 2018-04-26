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

	"fmt"
	"github.com/caicloud/cyclone/cloud"
	"github.com/caicloud/cyclone/pkg/api"
	"github.com/caicloud/cyclone/pkg/osutil"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	restful "github.com/emicklei/go-restful"
)

// test state value
var state = "fake-state"

func (router *router) acceptCode(request *restful.Request, response *restful.Response) {
	code := request.QueryParameter("code")
	state := request.QueryParameter("state")
	cyclonePath := osutil.GetStringEnv(cloud.CycloneServer, "http://127.0.0.1:7099")
	redirectURL := fmt.Sprintf("%s/%s/scm/%s/token?code=%s&state=%s", cyclonePath, "api/v1", api.GITLAB, code, state)
	response.AddHeader("Location", redirectURL)
	response.WriteHeaderAndEntity(http.StatusMovedPermanently, redirectURL)
}

// getAuthCodeURL handles the request to get the oauth url.
func (router *router) getAuthCodeURL(request *restful.Request, response *restful.Response) {
	scmType := request.PathParameter("type")

	scm, err := router.scmManager.FindScm(scmType)
	if err != nil {
		httputil.ResponseWithError(response, err)
	}

	url, err := scm.GetAuthCodeURL(state, scmType)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.WriteEntity(url)
}

// authcallback handles the request auth back from git.
func (router *router) authcallback(request *restful.Request, response *restful.Response) {
	code := request.QueryParameter("code")
	state := request.QueryParameter("state")
	scmType := request.PathParameter("type")

	scm, err := router.scmManager.FindScm(scmType)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	redirectURL, err := scm.Authcallback(code, state)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	response.AddHeader("Location", redirectURL)
	response.WriteHeaderAndEntity(http.StatusMovedPermanently, redirectURL)
}

// listrepos handles the request to list repositories
func (router *router) listrepos(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter("project")
	scmType := request.PathParameter("type")

	project, err := router.projectManager.GetProject(projectName)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	scm, err := router.scmManager.FindScm(scmType)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	repos, usernname, avatarURL, err := scm.GetRepos(project.ID)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	var responseResult api.ListReposResponse

	responseResult.Repos = repos
	responseResult.Username = usernname
	responseResult.AvatarURL = avatarURL

	response.WriteEntity(responseResult)
}

// logout handles the request to log out and delete related token in db.
func (router *router) logout(request *restful.Request, response *restful.Response) {
	projectName := request.PathParameter("project")
	scmType := request.PathParameter("type")
	project, err := router.projectManager.GetProject(projectName)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	scm, err := router.scmManager.FindScm(scmType)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}

	if err = scm.LogOut(project.ID); err != nil {
		httputil.ResponseWithError(response, err)
		return
	}
}
