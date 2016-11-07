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
	"net/http"

	"github.com/caicloud/cyclone/api"
	"github.com/caicloud/cyclone/pkg/log"
	"github.com/emicklei/go-restful"
)

// requesttoken returns the url address where owner can grant
// github's authentication to the applicant.
//
// GET: /api/v0.1/:uid/remotes/:code_repository/requesttoken
//
// RESPONSE: (RequestTokenResponse)
//  {
//    "enterpoint": (string) token request enterpoint
//    "error_msg": (string) set IF the request fails.
//  }
func requesttoken(request *restful.Request, response *restful.Response) {
	userID := request.PathParameter("user_id")
	coderepository := request.PathParameter("code_repository")
	var requestTResponse api.RequestTokenResponse

	remote, err := remoteManager.FindRemote(coderepository)
	if err != nil {
		message := "Unable to get remote according coderepository"
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		requestTResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, requestTResponse)
		return
	}

	url, err := remote.GetTokenQuestURL(userID)
	if err != nil {
		message := "Unable to get the request url"
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		requestTResponse.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, requestTResponse)
		return
	}

	requestTResponse.Enterpoint = url
	response.WriteEntity(requestTResponse)
}

// authcallback redirects to caicloud's web.
//
// GET: /api/v0.1/remotes/:code_repository/authcallback
//
// RESPONSE: (AuthCallbackResponse)
//  {
//    "result": (string) authcallback result
//    "error_msg": (string) set IF the request fails.
//  }
func authcallback(request *restful.Request, response *restful.Response) {
	code := request.QueryParameter("code")
	state := request.QueryParameter("state")
	coderepository := request.PathParameter("code_repository")

	var responseResult api.AuthCallbackResponse

	remote, err := remoteManager.FindRemote(coderepository)
	if err != nil {
		message := "Unable to get remote according coderepository"
		log.ErrorWithFields(message, log.Fields{"user_id": state})
		responseResult.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, responseResult)
		return
	}

	redirectURL, err := remote.Authcallback(code, state)
	if err != nil {
		message := "Unable to get token"
		log.ErrorWithFields(message, log.Fields{"user_id": state})
		responseResult.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, responseResult)
		return
	}

	response.AddHeader("Location", redirectURL)
	responseResult.Result = redirectURL
	response.WriteHeaderAndEntity(http.StatusMovedPermanently, responseResult)

	return
}

// listrepo finds the token in db and use it to get the
// repos according to the code_repository.
// GET: /api/v0.1/:uid/remotes/:code_repository/listrepo
//
// RESPONSE: (ListrepoResponse)
//  {
//    "result": (string) token request result
//    "error_msg": (string) set IF the request fails.
//  }
func listrepo(request *restful.Request, response *restful.Response) {
	coderepository := request.PathParameter("code_repository")
	userID := request.PathParameter("user_id")

	var responseResult api.ListRepoResponse
	remote, err := remoteManager.FindRemote(coderepository)
	if err != nil {
		message := "Unable to get remote according coderepository"
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		responseResult.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, responseResult)
		return
	}

	repos, usernname, avatarURL, err := remote.GetRepos(userID)
	if err != nil {
		message := "Unable to get remote repos"
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		responseResult.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, responseResult)
		return
	}

	responseResult.Repos = repos
	responseResult.Username = usernname
	responseResult.AvatarURL = avatarURL

	response.WriteEntity(responseResult)
}

// logout finds the token in db and remove it according the code_repository.
// GET: /api/v0.1/:uid/remotes/:code_repository/logout
//
// RESPONSE: (LogoutResponse)
//  {
//    "result": (string) logout result
//    "error_msg": (string) set IF the request fails.
//  }
func logout(request *restful.Request, response *restful.Response) {
	coderepository := request.PathParameter("code_repository")
	userID := request.PathParameter("user_id")

	var responseResult api.LogoutResponse
	remote, err := remoteManager.FindRemote(coderepository)
	if err != nil {
		message := "Unable to get remote according coderepository"
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		responseResult.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, responseResult)
		return
	}
	err = remote.LogOut(userID)
	if err != nil {
		message := "Unable to logout"
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		responseResult.ErrorMessage = message
		response.WriteHeaderAndEntity(http.StatusInternalServerError, responseResult)
		return
	}
	responseResult.Result = "success"
	response.WriteEntity(responseResult)
}
