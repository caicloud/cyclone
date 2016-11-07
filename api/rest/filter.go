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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/caicloud/cyclone/pkg/log"
	"github.com/caicloud/cyclone/pkg/osutil"
	"github.com/caicloud/cyclone/store"
	"github.com/emicklei/go-restful"
)

// DefaultAuthAddress is the default address of auth server.
const DefaultAuthAddress = "https://default-auth-address"

// checkUserAuth checks if the user is logined.
func checkUserAuth(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	// Don't check Github or other auth callback.
	if strings.Contains(request.SelectedRoutePath(), "authcallback") {
		chain.ProcessFilter(request, response)
		return
	}

	userID := request.PathParameter("user_id")
	token := request.HeaderParameter("token")

	// TODO: SSE doesn't have good support for authentication.
	if strings.Contains(request.SelectedRoutePath(), "logs") && request.QueryParameter("logs") == "true" {
		// Validation passed and pass on to specific api operation.
		chain.ProcessFilter(request, response)
		return
	}

	if token == "" || userID == "" {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusUnauthorized, "auth failed ,maybe you have not already signed in")
		return
	}

	authhost := osutil.GetStringEnv("AUTH_HOST", DefaultAuthAddress)
	// Notice: The request URL format can be find at caicloud/auth repo.
	url := fmt.Sprintf("%s/api/v0.1/users/%s/tokens/authenticate", authhost, userID)
	payload := strings.NewReader(`{"token": "` + token + `"}`)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/json")

	// Initialize http client and send the request.
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	res, err := client.Do(req)
	if err != nil {
		message := fmt.Sprintf("Failed to send request to auth server: %v\n", err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "token": token, "error": err})
		response.WriteHeaderAndEntity(http.StatusUnauthorized, message)
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		message := fmt.Sprintf("Failed to get response from auth server: %v\n", err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "token": token, "error": err})
		response.WriteHeaderAndEntity(http.StatusUnauthorized, message)
		return
	}
	defer res.Body.Close()

	// Convert response body from binary to json.
	var dat map[string]string
	if err := json.Unmarshal(body, &dat); err != nil {
		message := fmt.Sprintf("Failed to parse response from auth server: %v\n", err)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "token": token, "error": err})
		response.WriteHeaderAndEntity(http.StatusUnauthorized, message)
		return
	}

	if dat["error"] == "" {
		// Validation passed and pass on to specific api operation.s
		chain.ProcessFilter(request, response)
	} else {
		message := fmt.Sprintf("Validation failed\n")
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "token": token, "error": err})
		response.WriteHeaderAndEntity(http.StatusUnauthorized, message)
		return
	}
}

// checkACLForService checks whether the user has access to a specific service.
// TODO: The filter function will query mongo, so now a single request will have more
//       than one query.
//       Possible solution: pass context to the next function.
func checkACLForService(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	userID := request.PathParameter("user_id")
	serviceID := request.PathParameter("service_id")

	ds := store.NewStore()
	defer ds.Close()

	service, err := ds.FindServiceByID(serviceID)
	if err != nil {
		message := fmt.Sprintf("Unable to find service %v", serviceID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	} else if service.UserID != userID {
		message := fmt.Sprintf("have no access to service %v", serviceID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		response.WriteHeaderAndEntity(http.StatusUnauthorized, message)
		return
	}

	// Validation passed and pass on to specific api operation.
	chain.ProcessFilter(request, response)
}

// checkACLForVersion checks whether the user has access to a specific version.
// TODO: The filter function will query mongo, so now a single request will have more
//       than one query.
//       Possible solution: pass context to the next function.
func checkACLForVersion(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	userID := request.PathParameter("user_id")
	versionID := request.PathParameter("version_id")

	service, _, err := findServiceAndVersion(versionID)
	if err != nil {
		message := fmt.Sprintf("Unable to find version %v", versionID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID, "error": err})
		response.WriteHeaderAndEntity(http.StatusNotFound, message)
		return
	} else if service.UserID != userID {
		message := fmt.Sprintf("have no access to version %v", versionID)
		log.ErrorWithFields(message, log.Fields{"user_id": userID})
		response.WriteHeaderAndEntity(http.StatusUnauthorized, message)
		return
	}

	// Validation passed and pass on to specific api operation.
	chain.ProcessFilter(request, response)
}
