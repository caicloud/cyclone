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

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/service/router"
)

type path struct {
	path     string
	segments []string
	names    map[string]int
}

func (p *path) String() string {
	return p.path
}

func (p *path) Path(values map[string]string) (string, error) {
	if len(p.names) == 0 {
		return strings.Join(p.segments, ""), nil
	}
	segments := make([]string, len(p.segments))
	copy(segments, p.segments)
	for key, index := range p.names {
		value, ok := values[key]
		if !ok {
			return "", noPathParameter.Error(key, p.path)
		}
		segments[index] = url.PathEscape(value)
	}
	return strings.Join(segments, ""), nil
}

// Client implements builder pattern for http client.
type Client struct {
	endpoint string
	config   *Config
	lock     sync.RWMutex
	paths    map[string]*path
}

// NewClient creates a client.
func NewClient(cfg *Config) (*Client, error) {
	cfg = cfg.DeepCopy()
	if err := cfg.Complete(); err != nil {
		return nil, err
	}
	client := &Client{
		endpoint: fmt.Sprintf("%s://%s/", cfg.Scheme, strings.TrimRight(cfg.Host, "/\\")),
		config:   cfg,
		paths:    map[string]*path{},
	}
	return client, nil
}

func (c *Client) parseURL(url string) (*path, error) {
	c.lock.RLock()
	p, ok := c.paths[url]
	c.lock.RUnlock()
	if ok {
		return p, nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	p, ok = c.paths[url]
	if ok {
		return p, nil
	}
	segments, err := router.Split(url)
	if err != nil {
		return nil, invalidPath.Error(url, err.Error())
	}
	p = &path{
		path:  url,
		names: map[string]int{},
	}
	for i, segment := range segments {
		if strings.HasPrefix(segment, "{") {
			// Remove "{" and "}".
			segment = segment[1 : len(segment)-1]
			index := strings.Index(segment, ":")
			if index > 0 {
				segment = segment[:index]
			}
			if _, ok := p.names[segment]; ok {
				return nil, duplicatedPathParameter.Error(segment, p.path)
			}
			p.names[segment] = i
		}
		p.segments = append(p.segments, segment)
	}
	c.paths[url] = p
	return p, nil
}

// Request creates an request with specific method and url path.
// The code is only for checking if status code of response is right.
func (c *Client) Request(method string, code int, url string) *Request {
	path, err := c.parseURL(url)
	req := &Request{
		err:      err,
		method:   method,
		code:     code,
		endpoint: c.endpoint,
		path:     path,
		client:   c.config.Executor,
		paths:    map[string]string{},
		queries:  map[string][]string{},
		headers:  map[string][]string{},
		forms:    map[string][]string{},
		files:    map[string]interface{}{},
	}
	return req
}

// Request describes a http request.
type Request struct {
	once            sync.Once
	err             error
	method          string
	code            int
	endpoint        string
	path            *path
	client          RequestExecutor
	paths           map[string]string
	queries         map[string][]string
	headers         map[string][]string
	forms           map[string][]string
	files           map[string]interface{}
	body            interface{}
	bodyContentType string
	meta            map[string]string
	data            interface{}
}

// Path sets path parameter.
func (r *Request) Path(name string, value interface{}) *Request {
	r.paths[name] = fmt.Sprint(value)
	return r
}

// Query sets query parameter.
func (r *Request) Query(name string, values ...interface{}) *Request {
	m := r.queries
	for _, value := range values {
		m[name] = append(m[name], fmt.Sprint(value))
	}
	return r
}

// Header sets header parameter.
func (r *Request) Header(name string, values ...interface{}) *Request {
	m := r.headers
	for _, value := range values {
		m[name] = append(m[name], fmt.Sprint(value))
	}
	return r
}

// Form sets form parameter.
func (r *Request) Form(name string, values ...interface{}) *Request {
	m := r.forms
	for _, value := range values {
		m[name] = append(m[name], fmt.Sprint(value))
	}
	return r
}

// File sets file parameter.
func (r *Request) File(name string, file interface{}) *Request {
	r.files[name] = file
	return r
}

// Body sets body parameter.
func (r *Request) Body(contentType string, value interface{}) *Request {
	r.body = value
	r.bodyContentType = contentType
	return r
}

// Meta sets header result.
func (r *Request) Meta(value *map[string]string) *Request {
	if *value == nil {
		*value = map[string]string{}
	}
	r.meta = *value
	return r
}

// Data sets body result. value must be a pointer.
func (r *Request) Data(value interface{}) *Request {
	r.data = value
	return r
}

// Do executes the request.
func (r *Request) Do(ctx context.Context) error {
	r.once.Do(func() {
		if err := r.do(ctx); err != nil {
			r.err = err
		}
	})
	return r.err
}

func (r *Request) do(ctx context.Context) error {
	if r.err != nil {
		return r.err
	}
	req, err := r.request(ctx)
	if err != nil {
		return err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	return r.finish(ctx, req, resp)
}

func (r *Request) request(ctx context.Context) (*http.Request, error) {
	if r.body != nil && (len(r.files) > 0 || len(r.forms) != 0) {
		return nil, conflictBodyParameter.Error(r.path.String())
	}
	urlPath, err := r.path.Path(r.paths)
	if err != nil {
		return nil, err
	}
	path := r.endpoint + strings.TrimLeft(urlPath, "/")
	urlVal := url.Values(r.queries)
	if len(urlVal) > 0 {
		path += "?" + urlVal.Encode()
	}
	contentType := r.bodyContentType
	buf := bytes.NewBuffer(nil)
	reader := io.Reader(buf)
	if r.body != nil {
		if body, ok := r.body.(io.Reader); ok {
			reader = body
		} else {
			// Write body to buffer.
			switch contentType {
			case definition.MIMEJSON:
				err = json.NewEncoder(buf).Encode(r.body)
			case definition.MIMEXML:
				err = xml.NewEncoder(buf).Encode(r.body)
			default:
				_, err = buf.WriteString(fmt.Sprint(r.body))
			}
			if err != nil {
				return nil, unconvertibleObject.Error(reflect.TypeOf(r.body).String(), r.path.String(), err.Error())
			}
		}
	} else {
		if len(r.files) > 0 {
			// Construct multipart form.
			parts := multipart.NewWriter(buf)
			for k, values := range r.forms {
				for _, value := range values {
					// Create parts for form values.
					w, err := parts.CreateFormField(k)
					if err == nil {
						_, err = io.WriteString(w, value)
					}
					if err != nil {
						return nil, unwritableForm.Error(k, r.path.String(), err.Error())
					}
				}
			}
			for k, v := range r.files {
				w, err := parts.CreateFormFile(k, k)
				if err == nil {
					// For io.Reader and []byte, write directly.
					if r, ok := v.(io.Reader); ok {
						_, err = io.Copy(w, r)
					} else {
						switch data := v.(type) {
						case []byte:
							_, err = w.Write(data)
						default:
							// For other types, print it.
							_, err = fmt.Fprint(w, v)
						}
					}
				}
				if err != nil {
					return nil, unwritableFile.Error(k, r.path.String(), err.Error())
				}
			}
			contentType = parts.FormDataContentType()
		} else if len(r.forms) > 0 {
			// Write form data to buffer.
			contentType = definition.MIMEURLEncoded
			formVal := url.Values(r.forms)
			_, err := buf.WriteString(formVal.Encode())
			if err != nil {
				return nil, unwritableForms.Error(r.path.String(), err.Error())
			}
		}
	}
	req, err := http.NewRequest(r.method, path, reader)
	if err != nil {
		return nil, invalidRequest.Error(r.path.String(), err.Error())
	}
	for k, values := range r.headers {
		for _, value := range values {
			req.Header.Add(k, value)
		}
	}
	// Reset Content-Type.
	req.Header.Set("Content-Type", contentType)
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	return req, nil
}

func (r *Request) finish(ctx context.Context, req *http.Request, resp *http.Response) (err error) {
	reader := &autocloser{resp.Body}
	defer func() {
		if err != nil {
			e := reader.Close()
			// Ignore error.
			_ = e
		}
	}()
	// Fill headers.
	if r.meta != nil {
		for k, v := range resp.Header {
			if len(v) > 0 {
				r.meta[k] = v[0]
			} else {
				r.meta[k] = ""
			}
		}
	}
	ct := resp.Header.Get("Content-Type")
	if resp.StatusCode >= 200 && resp.StatusCode < 299 {
		if resp.StatusCode != r.code {
			return unmatchedStatusCode.Error(r.path.String(), r.code, resp.StatusCode)
		}
		// Unmarshal body to target.
		if r.data != nil {
			contentType, _, err := mime.ParseMediaType(ct)
			if err != nil {
				return invalidContentType.Error(ct, r.path.String(), err.Error())
			}
			switch target := r.data.(type) {
			case *io.Reader:
				*target = reader
			case *io.ReadCloser:
				*target = reader
			case *[]byte:
				data, err := ioutil.ReadAll(reader)
				if err != nil {
					return unreadableBody.Error(r.path.String(), err.Error())
				}
				*target = data
			case *string:
				data, err := ioutil.ReadAll(reader)
				if err != nil {
					return unreadableBody.Error(r.path.String(), err.Error())
				}
				*target = string(data)
			default:
				switch contentType {
				case definition.MIMEJSON:
					if err := json.NewDecoder(reader).Decode(r.data); err != nil {
						return unreadableBody.Error(r.path.String(), err.Error())
					}
				case definition.MIMEXML:
					if err := xml.NewDecoder(reader).Decode(r.data); err != nil {
						return unreadableBody.Error(r.path.String(), err.Error())
					}
				}
				return unrecognizedBody.Error(r.path.String(), "no appropriate receiver")
			}
		}
	} else {
		contentType, _, err := mime.ParseMediaType(ct)
		if err != nil {
			return invalidContentType.Error(ct, r.path.String(), err.Error())
		}
		// Unmarshal body to error.
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return unreadableBody.Error(r.path.String(), err.Error())
		}
		dt := errors.DataTypePlain
		switch contentType {
		case definition.MIMEJSON:
			dt = errors.DataTypeJSON
		case definition.MIMEXML:
			dt = errors.DataTypeXML
		}
		e, err := errors.ParseError(resp.StatusCode, dt, data)
		if err != nil {
			return unreadableBody.Error(r.path.String(), err.Error())
		}
		return e
	}
	return nil
}

type autocloser struct {
	io.ReadCloser
}

func (ac *autocloser) Read(p []byte) (n int, err error) {
	count, err := ac.ReadCloser.Read(p)
	if err == io.EOF {
		if e := ac.ReadCloser.Close(); e != nil {
			return count, e
		}
	}
	return count, err
}
