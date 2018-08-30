/*
Copyright 2018 Caicloud Authors

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

package reqlog

import (
	"context"
	"time"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

func init() {
	nirvana.RegisterConfigInstaller(&reqlogInstaller{})
}

// ExternalConfigName is the external config name of request logger.
const ExternalConfigName = "reqlog"

// config is reqlog config.
type config struct {
	doubleLog  bool
	sourceAddr bool
	requestKey string
	requestID  bool
	logger     log.Logger
}

type reqlogInstaller struct{}

// Name is the external config name.
func (i *reqlogInstaller) Name() string {
	return ExternalConfigName
}

// Install installs stuffs before server starting.
func (i *reqlogInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		begin, end := i.buildPrinters(c)
		err = builder.AddDescriptor(definition.Descriptor{
			Path: "/",
			Middlewares: []definition.Middleware{
				func(ctx context.Context, next definition.Chain) error {
					start := time.Now()
					httpCtx := service.HTTPContextFrom(ctx)
					begin(httpCtx, nil)

					err := next.Continue(ctx)
					end(httpCtx, map[string]interface{}{
						intervalDuration: time.Since(start),
						responseError:    err,
					})
					return err
				},
			},
		})
	})
	return err
}

type printer func(ctx service.HTTPContext, data map[string]interface{})

const (
	intervalDuration = "intervalDuration"
	responseError    = "responseError"
)

func (i *reqlogInstaller) buildPrinters(c *config) (begin printer, end printer) {
	logger := c.logger
	output := func(ctx service.HTTPContext, data map[string]interface{},
		components []func(ctx service.HTTPContext, data map[string]interface{}) interface{}) {
		results := make([]interface{}, 0, len(components))
		for _, c := range components {
			result := c(ctx, data)
			if result == nil {
				continue
			}
			results = append(results, result)
		}
		if logger != nil {
			logger.Infoln(results...)
		} else {
			log.Infoln(results...)
		}
	}
	method := func(ctx service.HTTPContext, data map[string]interface{}) interface{} {
		return ctx.Request().Method
	}
	url := func(ctx service.HTTPContext, data map[string]interface{}) interface{} {
		return ctx.Request().URL.String()
	}
	clientAddr := func(ctx service.HTTPContext, data map[string]interface{}) interface{} {
		return ctx.Request().RemoteAddr
	}
	statusCode := func(ctx service.HTTPContext, data map[string]interface{}) interface{} {
		return ctx.ResponseWriter().StatusCode()
	}
	contentLength := func(ctx service.HTTPContext, data map[string]interface{}) interface{} {
		return ctx.ResponseWriter().ContentLength()
	}
	requestID := func(ctx service.HTTPContext, data map[string]interface{}) interface{} {
		id := ctx.Request().Header.Get(c.requestKey)
		if id == "" {
			return nil
		}
		return id
	}
	interval := func(ctx service.HTTPContext, data map[string]interface{}) interface{} {
		if data != nil {
			if result, ok := data[intervalDuration]; ok {
				return result
			}
		}
		return nil
	}
	respErr := func(ctx service.HTTPContext, data map[string]interface{}) interface{} {
		if data != nil {
			if result, ok := data[responseError]; ok {
				return result
			}
		}
		return nil
	}

	beginning := []func(ctx service.HTTPContext, data map[string]interface{}) interface{}{}
	if c.doubleLog {
		beginning = append(beginning, method, url)
		if c.requestID {
			beginning = append(beginning, requestID)
		}
		if c.sourceAddr {
			beginning = append(beginning, clientAddr)
		}
	}

	ending := []func(ctx service.HTTPContext, data map[string]interface{}) interface{}{
		method,
		statusCode,
		contentLength,
		interval,
		url,
	}
	if c.requestID {
		ending = append(ending, requestID)
	}
	if c.sourceAddr {
		ending = append(ending, clientAddr)
	}
	ending = append(ending, respErr)
	return func(ctx service.HTTPContext, data map[string]interface{}) {
			if c.doubleLog {
				output(ctx, data, beginning)
			}
		},
		func(ctx service.HTTPContext, data map[string]interface{}) {
			output(ctx, data, ending)
		}
}

// Uninstall uninstalls stuffs after server terminating.
func (i *reqlogInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	return nil
}

// Disable returns a configurer to disable reqlog.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}

// Default Configurer does nothing but ensure default config was set.
func Default() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.logger = nil
		})
		return nil
	}
}

// Logger Configurer sets logger.
func Logger(l log.Logger) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.logger = l
		})
		return nil
	}
}

// DoubleLog returns a configurer to enable or
// disable double log. If it's enabled, every
// request outputs two entries. One for starting
// and another for ending. If it's disabled, only
// outputs ending entry.
// Defaults to false.
func DoubleLog(enable bool) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.doubleLog = enable
		})
		return nil
	}
}

// SourceAddr returns a configurer to enable or
// disable showing source addr.
// Defaults to false.
func SourceAddr(enable bool) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.sourceAddr = enable
		})
		return nil
	}
}

// RequestID returns a configurer to enable or
// disable showing request id.
// Defaults to false.
func RequestID(enable bool) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.requestID = enable
		})
		return nil
	}
}

// RequestIDKey returns a configurer to set header key
// of request id.
// Defaults to X-Request-Id.
func RequestIDKey(key string) nirvana.Configurer {
	if key == "" {
		key = "X-Request-Id"
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.requestKey = key
		})
		return nil
	}
}

func wrapper(c *nirvana.Config, f func(c *config)) {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{
			requestKey: "X-Request-Id",
		}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// Option contains basic configurations of reqlog.
type Option struct {
	DoubleLog    bool   `desc:"Output two entries for every request"`
	SourceAddr   bool   `desc:"Output source addr for request log"`
	RequestID    bool   `desc:"Output request id for request log"`
	RequestIDKey string `desc:"Request header key for request id"`
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		DoubleLog:    false,
		SourceAddr:   false,
		RequestID:    false,
		RequestIDKey: "X-Request-Id",
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	cfg.Configure(
		DoubleLog(p.DoubleLog),
		SourceAddr(p.SourceAddr),
		RequestID(p.RequestID),
		RequestIDKey(p.RequestIDKey),
	)
	return nil
}
