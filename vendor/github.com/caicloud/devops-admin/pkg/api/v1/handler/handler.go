/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	log "github.com/golang/glog"

	"github.com/caicloud/devops-admin/pkg/api/v1/interceptor"
)

// handler handles the request through the reverse proxy.
type handler struct {
	proxy *httputil.ReverseProxy
}

// NewHandler creates the handler with reverse proxy.
func NewHandler(cycloneURL string) (http.Handler, error) {
	targetURL, err := url.Parse(cycloneURL)
	if err != nil {
		return nil, err
	}

	return handler{httputil.NewSingleHostReverseProxy(targetURL)}, nil
}

// ServeHTTP handles and proxies the request.
func (h handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	if interceptor.IsWorkspaceOperation(path) {
		interceptor.HandleWorkspace(rw, req, h.proxy)
	} else if interceptor.IsPipelineOperation(path) {
		interceptor.HandlePipeline(rw, req, h.proxy)
	} else {
		// Proxy other requests.
		log.Infof("Proxy other request: %s %s", req.Method, req.URL.Path)

		h.proxy.ServeHTTP(rw, req)
	}
}
