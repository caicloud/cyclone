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

package http

import (
	"net/http"
	"net/http/httptrace"

	"github.com/caicloud/nirvana/plugins/tracing/utils"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	tlog "github.com/opentracing/opentracing-go/log"
)

// Transport wraps a RoundTripper. If a request is being traced with
// Tracer, Transport will inject the current span into the headers,
// and set HTTP related tags on the span.
type Transport struct {
	// The actual RoundTripper to use for the request. A nil
	// RoundTripper defaults to http.DefaultTransport.
	http.RoundTripper
	// Enable the ClientTrace, See https://blog.golang.org/http-tracing for more.
	EnableHTTPtrace bool
	// The default is HTTP method
	OperationName string
}

// RoundTrip implements the RoundTripper interface.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}

	operationName := t.OperationName
	if operationName == "" {
		operationName = req.Method
	}

	span, ctx := utils.StartSpanFromContext(req.Context(), "HTTP: "+operationName)
	if span == nil {
		return rt.RoundTrip(req)
	}
	defer span.Finish()

	if t.EnableHTTPtrace {
		ht := tracer{span: span}
		req = req.WithContext(httptrace.WithClientTrace(ctx, ht.clientTrace()))
	}

	ext.HTTPMethod.Set(span, req.Method)
	ext.HTTPUrl.Set(span, req.URL.String())

	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	_ = span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier) // #nosec

	resp, err := rt.RoundTrip(req)
	if err != nil {
		ext.Error.Set(span, true)
		span.LogFields(
			tlog.Error(err),
		)
		return resp, err
	}

	ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
	return resp, err
}

type tracer struct {
	span opentracing.Span
}

func (t *tracer) clientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn:              t.getConn,
		GotConn:              t.gotConn,
		PutIdleConn:          t.putIdleConn,
		GotFirstResponseByte: t.gotFirstResponseByte,
		Got100Continue:       t.got100Continue,
		DNSStart:             t.dnsStart,
		DNSDone:              t.dnsDone,
		ConnectStart:         t.connectStart,
		ConnectDone:          t.connectDone,
		WroteHeaders:         t.wroteHeaders,
		Wait100Continue:      t.wait100Continue,
		WroteRequest:         t.wroteRequest,
	}
}

func (t *tracer) getConn(hostPort string) {
	ext.HTTPUrl.Set(t.span, hostPort)
	t.span.LogFields(tlog.String("event", "GetConn"))
}

func (t *tracer) gotConn(info httptrace.GotConnInfo) {
	t.span.SetTag("net/http.reused", info.Reused)
	t.span.SetTag("net/http.was_idle", info.WasIdle)
	t.span.LogFields(tlog.String("event", "GotConn"))
}

func (t *tracer) putIdleConn(error) {
	t.span.LogFields(tlog.String("event", "PutIdleConn"))
}

func (t *tracer) gotFirstResponseByte() {
	t.span.LogFields(tlog.String("event", "GotFirstResponseByte"))
}

func (t *tracer) got100Continue() {
	t.span.LogFields(tlog.String("event", "Got100Continue"))
}

func (t *tracer) dnsStart(info httptrace.DNSStartInfo) {
	t.span.LogFields(
		tlog.String("event", "DNSStart"),
		tlog.String("host", info.Host),
	)
}

func (t *tracer) dnsDone(info httptrace.DNSDoneInfo) {
	fields := []tlog.Field{tlog.String("event", "DNSDone")}
	for _, addr := range info.Addrs {
		fields = append(fields, tlog.String("addr", addr.String()))
	}
	if info.Err != nil {
		fields = append(fields, tlog.Error(info.Err))
	}
	t.span.LogFields(fields...)
}

func (t *tracer) connectStart(network, addr string) {
	t.span.LogFields(
		tlog.String("event", "ConnectStart"),
		tlog.String("network", network),
		tlog.String("addr", addr),
	)
}

func (t *tracer) connectDone(network, addr string, err error) {
	if err != nil {
		t.span.LogFields(
			tlog.String("message", "ConnectDone"),
			tlog.String("network", network),
			tlog.String("addr", addr),
			tlog.String("event", "error"),
			tlog.Error(err),
		)
	} else {
		t.span.LogFields(
			tlog.String("event", "ConnectDone"),
			tlog.String("network", network),
			tlog.String("addr", addr),
		)
	}
}

func (t *tracer) wroteHeaders() {
	t.span.LogFields(tlog.String("event", "WroteHeaders"))
}

func (t *tracer) wait100Continue() {
	t.span.LogFields(tlog.String("event", "Wait100Continue"))
}

func (t *tracer) wroteRequest(info httptrace.WroteRequestInfo) {
	if info.Err != nil {
		t.span.LogFields(
			tlog.String("message", "WroteRequest"),
			tlog.String("event", "error"),
			tlog.Error(info.Err),
		)
		ext.Error.Set(t.span, true)
	} else {
		t.span.LogFields(tlog.String("event", "WroteRequest"))
	}
}
