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

package tracing

import (
	"context"
	"io"
	"time"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tlog "github.com/opentracing/opentracing-go/log"
	tconfig "github.com/uber/jaeger-client-go/config"
)

func init() {
	nirvana.RegisterConfigInstaller(&tracingInstaller{})
}

// ExternalConfigName is the external config name of tracing.
const ExternalConfigName = "tracing"

// config is tracing config.
type config struct {
	serviceName   string
	agentHostPort string
	tracer        opentracing.Tracer
	closer        io.Closer
	hook          Hook
}

type tracingInstaller struct{}

// Name is the external config name.
func (i *tracingInstaller) Name() string {
	return ExternalConfigName
}

// Install installs config to builder.
func (i *tracingInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		if c.tracer == nil {
			tcfg := tconfig.Configuration{
				ServiceName: c.serviceName,
				Sampler: &tconfig.SamplerConfig{
					Type:  "const",
					Param: 1,
				},
				Reporter: &tconfig.ReporterConfig{
					LogSpans:            false,
					BufferFlushInterval: 1 * time.Second,
					LocalAgentHostPort:  c.agentHostPort,
				},
			}
			c.tracer, c.closer, err = tcfg.NewTracer(
				tconfig.Logger(&loggerAdapter{cfg.Logger()}),
			)
			if err != nil {
				return
			}
		}
		err = builder.AddDescriptor(descriptor(c.tracer, c.hook))
	})
	return err

}

// Uninstall uninstalls stuffs after server terminating.
func (i *tracingInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		if c.closer != nil {
			err = c.closer.Close()
		}
	})
	return err
}

// Disable returns a configurer to disable tracing.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}

func wrapper(c *nirvana.Config, f func(c *config)) {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// CustomTracer returns a configurer to set custom tracer.
func CustomTracer(tracer opentracing.Tracer) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.tracer = tracer
		})
		return nil
	}
}

// DefaultTracer returns a configurer to create default tracer by service and port.
func DefaultTracer(serviceName string, agentHostPort string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.serviceName = serviceName
			c.agentHostPort = agentHostPort
		})
		return nil
	}
}

// AddHook returns a configurer to set request hook.
func AddHook(hook Hook) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.hook = hook
		})
		return nil
	}
}

type event string

const (
	eventRequest  event = "request"
	eventResponse event = "response"
)

// descriptor creates descriptor for middleware.
func descriptor(tracer opentracing.Tracer, hook Hook) definition.Descriptor {
	return definition.Descriptor{
		Path:        "/",
		Middlewares: []definition.Middleware{middleware(tracer, hook)},
	}
}

// middleware created definition middleware for tracing.
func middleware(tracer opentracing.Tracer, hook Hook) definition.Middleware {
	return func(ctx context.Context, next definition.Chain) error {
		httpCtx := service.HTTPContextFrom(ctx)
		req := httpCtx.Request()
		// extract span context from HTTP Headers
		// Don't check errors. if "spanContext" is nil, it will start a new trace id.
		spanContext, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)) // #nosec

		span := tracer.StartSpan(httpCtx.RoutePath(), ext.RPCServerOption(spanContext))
		defer span.Finish()

		// set standard tags
		ext.HTTPUrl.Set(span, req.URL.String())
		ext.HTTPMethod.Set(span, req.Method)
		ext.Component.Set(span, "nirvana/middlewares/trace")

		span.LogFields(
			tlog.String("event", string(eventRequest)),
		)

		if hook != nil {
			hook.Before(ctx, span)
		}
		ctx = opentracing.ContextWithSpan(ctx, span)

		defer func() {
			span.LogFields(
				tlog.String("event", string(eventResponse)),
			)
			if hook != nil {
				hook.After(ctx, span)
			}
		}()
		if err := next.Continue(ctx); err != nil {
			ext.HTTPStatusCode.Set(span, 500)
			ext.Error.Set(span, true)
			return err
		}

		resp := httpCtx.ResponseWriter()
		code := resp.StatusCode()
		ext.HTTPStatusCode.Set(span, uint16(code))
		if code >= 400 {
			ext.Error.Set(span, true)
		}

		return nil
	}
}

type loggerAdapter struct {
	logger log.Logger
}

// Error logs a message at error priority
func (l *loggerAdapter) Error(msg string) {
	l.logger.Error(msg)
}

// Infof logs a message at info priority
func (l *loggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

// Option contains basic configurations of tracing.
type Option struct {
	// ServiceName is the trace service name
	ServiceName string
	// AgentHostPort instructs reporter to send spans to jaeger-agent at this address
	AgentHostPort string
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		ServiceName:   "",
		AgentHostPort: "",
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	cfg.Configure(
		DefaultTracer(p.ServiceName, p.AgentHostPort),
	)
	return nil
}
