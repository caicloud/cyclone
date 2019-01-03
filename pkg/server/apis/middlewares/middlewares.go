package middlewares

import (
	"context"
	"time"

	def "github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

// Middlewares returns a list of middlewares.
func Middlewares() []def.Middleware {
	return []def.Middleware{newLogMiddleware()}
}

func newLogMiddleware() def.Middleware {
	return func(ctx context.Context, next def.Chain) error {
		startTime := time.Now()
		err := next.Continue(ctx)

		httpCtx := service.HTTPContextFrom(ctx)
		req := httpCtx.Request()
		resp := httpCtx.ResponseWriter()

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

		return err
	}
}
