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

package utils

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
)

// StartSpanFromContext starts and returns a Span with `operationName`, using
// any Span found within `ctx` as a ChildOfRef.
func StartSpanFromContext(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span = span.Tracer().StartSpan(operationName, opentracing.ChildOf(span.Context()))
		ctx = opentracing.ContextWithSpan(ctx, span)
		return span, ctx
	}

	return nil, ctx
}
