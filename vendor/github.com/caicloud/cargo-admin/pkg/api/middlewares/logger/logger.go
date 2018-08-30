package logger

import (
	"context"
	"time"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

var (
	green        = "\x1b[42m"
	white        = "\x1b[47m"
	yellow       = "\x1b[43m"
	red          = "\x1b[41m"
	blue         = "\x1b[44m"
	magenta      = "\x1b[45m"
	cyan         = "\x1b[46m"
	reset        = "\x1b[0m"
	disableColor = false
)

type Logger interface {
	Infof(string, ...interface{})
}

func New(log Logger) func(context.Context, definition.Chain) error {
	return func(ctx context.Context, chain definition.Chain) error {
		var err error

		start := time.Now()
		req := service.HTTPContextFrom(ctx).Request()
		path := req.URL.Path
		raw := req.URL.RawPath

		err = chain.Continue(ctx)

		w := service.HTTPContextFrom(ctx).ResponseWriter()
		end := time.Now()
		latency := end.Sub(start)
		clientIP := req.RemoteAddr
		method := req.Method
		statusCode := w.StatusCode()
		if raw != "" {
			path = path + "?" + raw
		}

		comment := ""
		if err != nil {
			comment = err.Error()
		}
		log.Infof("%s %3d %s %13v | %15s | %s %-7s %s %s\n%s",
			colorForStatus(statusCode), statusCode, reset,
			latency,
			clientIP,
			colorForMethod(method), method, reset,
			path,
			comment,
		)

		return err
	}
}

func colorForMethod(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return green
	case code >= 300 && code < 400:
		return white
	case code >= 400 && code < 500:
		return yellow
	default:
		return red
	}
}
