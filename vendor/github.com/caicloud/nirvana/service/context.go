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

package service

import (
	"bufio"
	"context"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
)

var (
	// contextKeyUnderlyingHTTPContext is a key for context.
	// It's unique and point to httpCtx.
	contextKeyUnderlyingHTTPContext interface{} = new(byte)
)

// httpCtx contains a http.Request and a http.ResponseWriter for a request.
// It goes through the life cycle of a request.
type httpCtx struct {
	context.Context
	container container
	response  response
	path      string
}

func newHTTPContext(resp http.ResponseWriter, request *http.Request) *httpCtx {
	ctx := &httpCtx{}
	ctx.Context = request.Context()
	ctx.container.request = request
	ctx.container.params = make([]param, 0, 5)
	ctx.response.writer = resp
	return ctx
}

// Value returns itself when key is contextKeyUnderlyingHTTPContext.
func (c *httpCtx) Value(key interface{}) interface{} {
	if key == contextKeyUnderlyingHTTPContext {
		return c
	}
	return c.Context.Value(key)
}

// ValueContainer contains values from a request.
type ValueContainer interface {
	// Path returns path value by key.
	Path(key string) (string, bool)
	// Query returns value from query string.
	Query(key string) ([]string, bool)
	// Header returns value by header key.
	Header(key string) ([]string, bool)
	// Form returns value from request. It is valid when
	// http "Content-Type" is "application/x-www-form-urlencoded"
	// or "multipart/form-data".
	Form(key string) ([]string, bool)
	// File returns a file reader when "Content-Type" is "multipart/form-data".
	File(key string) (multipart.File, bool)
	// Body returns a reader to read data from request body.
	// The reader only can read once.
	Body() (reader io.ReadCloser, contentType string, ok bool)
}

type param struct {
	key   string
	value string
}

// container implements ValueContainer and provides methods to get values.
type container struct {
	request *http.Request
	params  []param
	query   url.Values
}

// Set sets path parameter key-value pairs.
func (c *container) Set(key, value string) {
	c.params = append(c.params, param{key, value})
}

// Get gets path value.
func (c *container) Get(key string) (string, bool) {
	for i := len(c.params) - 1; i >= 0; i-- {
		p := c.params[i]
		if p.key == key {
			return p.value, true
		}
	}
	return "", false
}

// Path returns path value by key. It's same as Get().
func (c *container) Path(key string) (string, bool) {
	return c.Get(key)
}

// Query returns value from query string.
func (c *container) Query(key string) ([]string, bool) {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.removeEmpties(c.query[key])
}

// Header returns value by header key.
func (c *container) Header(key string) ([]string, bool) {
	h := c.request.Header[textproto.CanonicalMIMEHeaderKey(key)]
	return c.removeEmpties(h)
}

// Form returns value from request. It is valid when
// http "Content-Type" is "application/x-www-form-urlencoded"
// or "multipart/form-data".
func (c *container) Form(key string) ([]string, bool) {
	return c.removeEmpties(c.request.PostForm[key])
}

// removeEmpties removes empty strings.
func (c *container) removeEmpties(values []string) ([]string, bool) {
	if len(values) <= 0 {
		return values, false
	}
	results := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			results = append(results, value)
		}
	}
	return results, len(results) > 0
}

// File returns a file reader when "Content-Type" is "multipart/form-data".
func (c *container) File(key string) (multipart.File, bool) {
	file, _, err := c.request.FormFile(key)
	return file, err == nil
}

// Body returns a reader to read data from request body.
// The reader only can read once.
func (c *container) Body() (reader io.ReadCloser, contentType string, ok bool) {
	contentType, err := ContentType(c.request)
	return c.request.Body, contentType, err == nil
}

// ResponseWriter extends http.ResponseWriter.
type ResponseWriter interface {
	http.ResponseWriter
	// HeaderWritable can check whether WriteHeader() has
	// been called. If the method returns false, you should
	// not recall WriteHeader().
	HeaderWritable() bool
	// StatusCode returns status code.
	StatusCode() int
	// ContentLength returns the length of written content.
	ContentLength() int
}

type response struct {
	writer        http.ResponseWriter
	statusCode    int
	contentLength int
}

// For http.HTTPResponseWriter and HTTPResponseInfo
func (c *response) Header() http.Header {
	return c.writer.Header()
}

// Write is a disguise of http.response.Write().
func (c *response) Write(data []byte) (int, error) {
	if c.statusCode <= 0 {
		c.WriteHeader(200)
	}
	length, err := c.writer.Write(data)
	c.contentLength += length
	return length, err
}

// WriteHeader is a disguise of http.response.WriteHeader().
func (c *response) WriteHeader(code int) {
	c.statusCode = code
	c.writer.WriteHeader(code)
}

// Flush is a disguise of http.response.Flush().
func (c *response) Flush() {
	c.writer.(http.Flusher).Flush()
}

// CloseNotify is a disguise of http.response.CloseNotify().
func (c *response) CloseNotify() <-chan bool {
	return c.writer.(http.CloseNotifier).CloseNotify()
}

// Hijack is a disguise of http.response.Hijack().
func (c *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := c.writer.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, noConnectionHijacker.Error()
}

// StatusCode returns status code.
func (c *response) StatusCode() int {
	if c.statusCode <= 0 {
		return http.StatusOK
	}
	return c.statusCode
}

// ContentLength returns the length of written content.
func (c *response) ContentLength() int {
	return c.contentLength
}

// HeaderWritable can check whether WriteHeader() has
// been called. If the method returns false, you should
// not recall WriteHeader().
func (c *response) HeaderWritable() bool {
	return c.statusCode <= 0
}

// HTTPContext describes an http context.
type HTTPContext interface {
	Request() *http.Request
	ResponseWriter() ResponseWriter
	ValueContainer() ValueContainer
	RoutePath() string
	setRoutePath(path string)
}

// HTTPContextFrom get http context from context.
func HTTPContextFrom(ctx context.Context) HTTPContext {
	value := ctx.Value(contextKeyUnderlyingHTTPContext)
	if value == nil {
		return nil
	}
	if c, ok := value.(*httpCtx); ok {
		return c
	}
	return nil
}

// Request gets http.Request.
func (c *httpCtx) Request() *http.Request {
	return c.container.request
}

// ResponseWriter gets ResponseWriter.
func (c *httpCtx) ResponseWriter() ResponseWriter {
	return &c.response
}

// ValueContainer gets ValueContainer.
func (c *httpCtx) ValueContainer() ValueContainer {
	return &c.container
}

// RoutePath is the abstract path which matches request URL.
func (c *httpCtx) RoutePath() string {
	return c.path
}

// setRoutePath sets the abstract path which matches request URL.
func (c *httpCtx) setRoutePath(path string) {
	c.path = path
}
