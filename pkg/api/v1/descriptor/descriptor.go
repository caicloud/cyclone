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

package descriptor

import (
	"context"
	"net/http"
	"time"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

var descriptors = []definition.Descriptor{}

func register(ds ...definition.Descriptor) {
	descriptors = append(descriptors, ds...)
}

func Descriptor() definition.Descriptor {
	return definition.Descriptor{
		Path:        "/api/v1",
		Description: "It contains all APIs in v1",
		Consumes:    []string{definition.MIMEAll},
		Produces:    []string{definition.MIMEJSON},
		Middlewares: []definition.Middleware{newLogMiddleware()},
		Children:    descriptors,
	}
}

func newLogMiddleware() definition.Middleware {
	return func(ctx context.Context, next definition.Chain) error {
		startTime := time.Now()
		err := next.Continue(ctx)

		httpCtx := service.HTTPContextFrom(ctx)
		req := httpCtx.Request()
		resp := httpCtx.ResponseWriter()

		if req.Method != http.MethodGet {
			log.Infof("%s - [%s] \"%s %s %s\" %d %d %v",
				req.RemoteAddr,
				time.Now().Format("02/Jan/2006:15:04:05 -0700"),
				req.Method,
				req.URL.RequestURI(),
				req.Proto,
				resp.StatusCode,
				resp.ContentLength,
				time.Since(startTime),
			)
		} else {
			//log.V(log.LevelDebug).Infof("%s - [%s] \"%s %s %s\" %d %d %v",
			//	req.RemoteAddr,
			//	time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			//	req.Method,
			//	req.URL.RequestURI(),
			//	req.Proto,
			//	resp.StatusCode,
			//	resp.ContentLength,
			//	time.Since(startTime),
			//)
		}

		return err
	}
}
