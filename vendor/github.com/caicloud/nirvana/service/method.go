/*
Copyright 2017 Caicloud Authors

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

package service

import (
	"net/http"

	"github.com/caicloud/nirvana/definition"
)

// statusMapping contains a binding of HTTP method and success status code.
type statusMapping struct {
	// HTTPMethod is HTTP method.
	HTTPMethod string
	// Code means a success status code.
	Code int
}

var mappings = map[definition.Method]statusMapping{
	definition.List:        {http.MethodGet, http.StatusOK},
	definition.Get:         {http.MethodGet, http.StatusOK},
	definition.Create:      {http.MethodPost, http.StatusCreated},
	definition.Update:      {http.MethodPut, http.StatusOK},
	definition.Patch:       {http.MethodPatch, http.StatusOK},
	definition.Delete:      {http.MethodDelete, http.StatusNoContent},
	definition.AsyncCreate: {http.MethodPost, http.StatusAccepted},
	definition.AsyncUpdate: {http.MethodPut, http.StatusAccepted},
	definition.AsyncPatch:  {http.MethodPatch, http.StatusAccepted},
	definition.AsyncDelete: {http.MethodDelete, http.StatusAccepted},
}

// HTTPMethodFor gets a HTTP method for specified definition method.
func HTTPMethodFor(m definition.Method) string {
	return mappings[m].HTTPMethod
}

// HTTPCodeFor gets a success status code for specified definition method.
func HTTPCodeFor(m definition.Method) int {
	return mappings[m].Code
}

// RegisterMethod registers a HTTP method and a success status code for a definition method.
func RegisterMethod(method definition.Method, httpMethod string, code int) error {
	validHTTPMethod := []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace}
	found := false
	for _, m := range validHTTPMethod {
		if m == httpMethod {
			found = true
			break
		}
	}
	if !found {
		return invalidMethod.Error(httpMethod)
	}
	if code < 100 || code >= 600 {
		return invalidStatusCode.Error()
	}
	mappings[method] = statusMapping{
		Code:       code,
		HTTPMethod: httpMethod,
	}
	return nil
}
