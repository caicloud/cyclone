/*
Copyright 2014 The Kubernetes Authors.

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

package filters

import (
	"fmt"
	"net/http"

	"k8s.io/kubernetes/pkg/auth/authorizer"
	"k8s.io/kubernetes/pkg/util/runtime"
)

// badGatewayError renders a simple bad gateway error.
func badGatewayError(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusBadGateway)
	fmt.Fprintf(w, "Bad Gateway: %#v", req.RequestURI)
}

// forbidden renders a simple forbidden error
func forbidden(attributes authorizer.Attributes, w http.ResponseWriter, req *http.Request, reason string) {
	msg := forbiddenMessage(attributes)
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprintf(w, "%s: %q", msg, reason)
}

func forbiddenMessage(attributes authorizer.Attributes) string {
	username := ""
	if user := attributes.GetUser(); user != nil {
		username = user.GetName()
	}

	resource := attributes.GetResource()
	if group := attributes.GetAPIGroup(); len(group) > 0 {
		resource = resource + "." + group
	}

	if ns := attributes.GetNamespace(); len(ns) > 0 {
		return fmt.Sprintf("User %q cannot %s %s in the namespace %q.", username, attributes.GetVerb(), resource, ns)
	}

	return fmt.Sprintf("User %q cannot %s %s at the cluster scope.", username, attributes.GetVerb(), resource)
}

// internalError renders a simple internal error
func internalError(w http.ResponseWriter, req *http.Request, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Internal Server Error: %#v", req.RequestURI)
	runtime.HandleError(err)
}
