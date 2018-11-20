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

	"github.com/caicloud/nirvana/service"
	opentracing "github.com/opentracing/opentracing-go"
)

// Hook allows you to custom information for span.
type Hook interface {
	// Exec before request processing
	Before(ctx context.Context, span opentracing.Span)
	// Exec after request processing
	After(ctx context.Context, span opentracing.Span)
}

// DefaultHook is an default hook, it will record X-Request-ID from http headers
type DefaultHook struct{}

const (
	// HeaderRequestID is request id
	HeaderRequestID = "X-Request-ID"
)

// Before request processing
func (h *DefaultHook) Before(ctx context.Context, span opentracing.Span) {
	httpCtx := service.HTTPContextFrom(ctx)
	req := httpCtx.Request()

	span.SetTag(HeaderRequestID, req.Header.Get(HeaderRequestID))
}

// After Do nothing, because HTTP response can't gets payload.
func (h *DefaultHook) After(ctx context.Context, span opentracing.Span) {
}
