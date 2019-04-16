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
	"fmt"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/caicloud/nirvana/definition"
)

// Filter can filter request. It has the highest priority in a request
// lifecycle. It runs before router matching.
// If a filter return false, that means the request should be filtered.
// If a filter want to filter a request, it should handle the request
// by itself.
type Filter func(resp http.ResponseWriter, req *http.Request) bool

// RedirectTrailingSlash returns a filter to redirect request.
// If a request has trailing slash like `some-url/`, the filter will
// redirect the request to `some-url`.
func RedirectTrailingSlash() Filter {
	return func(resp http.ResponseWriter, req *http.Request) bool {
		path := req.URL.Path
		if len(path) > 1 && path[len(path)-1] == '/' {
			req.URL.Path = strings.TrimRight(path, "/")
			// Redirect to path without trailing slash.
			http.Redirect(resp, req, req.URL.String(), http.StatusTemporaryRedirect)
			return false
		}
		return true
	}
}

// FillLeadingSlash returns a pseudo filter to fill a leading slash when
// a request path does not have a leading slash.
// The filter won't filter anything.
func FillLeadingSlash() Filter {
	return func(resp http.ResponseWriter, req *http.Request) bool {
		path := req.URL.Path
		if len(path) <= 0 || path[0] != '/' {
			// Relative path may omit leading slash.
			req.URL.Path = "/" + path
		}
		return true
	}
}

// ParseRequestForm returns a filter to parse request form when content
// type is "application/x-www-form-urlencoded" or "multipart/form-data".
// The filter won't filter anything unless some error occurs in parsing.
func ParseRequestForm() Filter {
	return func(resp http.ResponseWriter, req *http.Request) bool {
		ct, err := ContentType(req)
		if err == nil {
			switch ct {
			case definition.MIMEURLEncoded:
				err = req.ParseForm()
			case definition.MIMEFormData:
				err = req.ParseMultipartForm(32 << 20)
			default:
				req.Form = req.URL.Query()
			}
		}
		if err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return false
		}
		return true
	}
}

func isGTZero(length string) bool {
	if length == "" {
		return false
	}
	i, err := strconv.Atoi(length)
	if err != nil {
		return false
	}
	return i > 0
}

// ContentType is a util to get content type from a request.
func ContentType(req *http.Request) (string, error) {
	ct := req.Header.Get("Content-Type")
	if ct == "" {
		length := req.Header.Get("Content-Length")
		transfer := req.Header.Get("Transfer-Encoding")
		if isGTZero(length) || transfer != "" {
			return definition.MIMEOctetStream, nil
		}
		return definition.MIMENone, nil
	}
	result, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return "", invalidContentType.Error(ct)
	}
	return result, nil
}

// AcceptTypes is a util to get accept types from a request.
// Accept types are sorted by q.
func AcceptTypes(req *http.Request) ([]string, error) {
	ct := req.Header.Get("Accept")
	if ct == "" {
		return []string{definition.MIMEAll}, nil
	}
	return parseAcceptTypes(ct)
}

func parseAcceptTypes(v string) ([]string, error) {
	types := []string{}
	factors := []float64{}
	strs := strings.Split(v, ",")
	for _, str := range strs {
		fields := strings.Split(str, ";")
		factor := 1.0
		ctFields := make([]string, 0, len(fields))
		for _, field := range fields {
			index := strings.IndexByte(field, '=')
			key := ""
			value := ""
			if index >= 0 {
				key = strings.TrimSpace(field[:index])
				value = strings.TrimSpace(field[index+1:])
				if key == "q" && len(value) > 0 {
					q, err := strconv.ParseFloat(value, 32)
					if err != nil {
						return nil, err
					}
					factor = q
					continue
				}
			} else {
				key = strings.TrimSpace(field)
			}
			if value == "" {
				ctFields = append(ctFields, key)
			} else {
				ctFields = append(ctFields, fmt.Sprintf("%s=%s", key, value))
			}
		}
		types = append(types, strings.Join(ctFields, ";"))
		factors = append(factors, factor)
	}
	if len(types) <= 1 {
		return types, nil
	}
	// In most cases, bubble sort is enough.
	// Can optimize here.
	exchanged := true
	for exchanged {
		exchanged = false
		for i := 1; i < len(factors); i++ {
			if factors[i] > factors[i-1] {
				types[i-1], types[i] = types[i], types[i-1]
				factors[i-1], factors[i] = factors[i], factors[i-1]
				exchanged = true
			}
		}
	}
	return types, nil
}
