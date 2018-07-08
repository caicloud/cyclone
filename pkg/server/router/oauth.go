/*
Copyright 2018 caicloud authors. All rights reserved.

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
	"github.com/caicloud/cyclone/pkg/scm"
	httputil "github.com/caicloud/cyclone/pkg/util/http"
	"github.com/emicklei/go-restful"
	"net/http"
)

// test state value
var state = "fake-state"

func (router *router) getOauthCode(request *restful.Request, response *restful.Response) {

	provider, err := scm.GetSCMProvider(api.Gitlab)
	if err != nil {
		httputil.ResponseWithError(response, err)
	}
	url, err := provider.GetAuthCodeURL(state, api.Gitlab)
	if err != nil {
		httputil.ResponseWithError(response, err)
	}
	response.WriteEntity(url)

}

func (router *router) getOauthToken(request *restful.Request, response *restful.Response) {
	code := request.QueryParameter("code")
	state = request.QueryParameter("state")

	provider, err := scm.GetSCMProvider(api.Gitlab)
	if err != nil {
		httputil.ResponseWithError(response, err)
	}

	redirectURL, err := provider.Authcallback(code, state)
	if err != nil {
		httputil.ResponseWithError(response, err)
		return
	}
	response.AddHeader("Location", redirectURL)
	response.WriteHeaderAndEntity(http.StatusMovedPermanently, redirectURL)
}
