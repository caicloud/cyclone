/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package v1

import (
	"time"

	"github.com/caicloud/devops-admin/pkg/api/definition"
	"github.com/caicloud/devops-admin/pkg/api/middleware"
	"github.com/caicloud/devops-admin/pkg/api/v1/descriptor"
	"github.com/golang/glog"
	"github.com/emicklei/go-restful"
)

// InstallRouters installs api WebService
func InstallRouters(containers *restful.Container) *restful.WebService {
	service := (&restful.WebService{}).
		ApiVersion("v1").
		Path("/api/v1").
		Doc("v1 API").
		Consumes("*/*", restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// Add filters
	service.Filter(NCSACommonLogFormatLogger())
	service.Filter(middleware.APIMetrics("v1"))
	service = definition.GenerateRoutes(service, descriptor.Descriptors)
	containers.Add(service)
	return service
}

// NCSACommonLogFormatLogger adds logs for every request using common log format.
func NCSACommonLogFormatLogger() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		r := req.Request
		start := time.Now()
		glog.Infof("Started %s - [%s] %s %s",
			req.Request.RemoteAddr,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			req.Request.URL.RequestURI(),
		)
		chain.ProcessFilter(req, resp)
		glog.Infof("%s - [%s] \"%s %s %s\" %d %d %v",
			req.Request.RemoteAddr,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			req.Request.Method,
			req.Request.URL.RequestURI(),
			req.Request.Proto,
			resp.StatusCode(),
			resp.ContentLength(),
			time.Since(start),
		)
	}
}
