package gock

import (
	"io/ioutil"
	"net/http"
	"regexp"
)

// EOL represents the end of line character.
const EOL = 0xa

// BodyTypes stores the supported MIME body types for matching.
// Currently only text-based types.
var BodyTypes = []string{
	"text/html",
	"text/plain",
	"application/json",
	"application/xml",
	"multipart/form-data",
	"application/x-www-form-urlencoded",
}

// BodyTypeAliases stores a generic MIME type by alias.
var BodyTypeAliases = map[string]string{
	"html": "text/html",
	"text": "text/plain",
	"json": "application/json",
	"xml":  "application/xml",
	"form": "multipart/form-data",
	"url":  "application/x-www-form-urlencoded",
}

// MatchMethod matches the HTTP method of the given request.
func MatchMethod(req *http.Request, ereq *Request) (bool, error) {
	return ereq.Method == "" || req.Method == ereq.Method, nil
}

// MatchScheme matches the request URL protocol scheme.
func MatchScheme(req *http.Request, ereq *Request) (bool, error) {
	return ereq.URLStruct.Scheme == "" || req.URL.Scheme == "" || ereq.URLStruct.Scheme == req.URL.Scheme, nil
}

// MatchHost matches the HTTP host header field of the given request.
func MatchHost(req *http.Request, ereq *Request) (bool, error) {
	url := ereq.URLStruct
	if url.Host == req.URL.Host {
		return true, nil
	}
	return regexp.MatchString(url.Host, req.URL.Host)
}

// MatchPath matches the HTTP URL path of the given request.
func MatchPath(req *http.Request, ereq *Request) (bool, error) {
	return regexp.MatchString(ereq.URLStruct.Path, req.URL.Path)
}

// MatchHeaders matches the headers fields of the given request.
func MatchHeaders(req *http.Request, ereq *Request) (bool, error) {
	for key, value := range ereq.Header {
		var err error
		var match bool

		for _, field := range req.Header[key] {
			match, err = regexp.MatchString(value[0], field)
			if err != nil {
				return false, err
			}
			if match {
				break
			}
		}

		if !match {
			return false, nil
		}
	}
	return true, nil
}

// MatchQueryParams matches the URL query params fields of the given request.
func MatchQueryParams(req *http.Request, ereq *Request) (bool, error) {
	for key, value := range ereq.URLStruct.Query() {
		var err error
		var match bool

		for _, field := range req.URL.Query()[key] {
			match, err = regexp.MatchString(value[0], field)
			if err != nil {
				return false, err
			}
			if match {
				break
			}
		}

		if !match {
			return false, nil
		}
	}
	return true, nil
}

// MatchBody tries to match the request body.
// TODO: not too smart now, needs several improvements.
func MatchBody(req *http.Request, ereq *Request) (bool, error) {
	// If match body is empty, just continue
	if req.Method == "GET" || len(ereq.BodyBuffer) == 0 {
		return true, nil
	}

	// Only can match certain MIME body types
	if !supportedType(req) {
		return false, nil
	}

	// Read the whole request body
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return false, err
	}

	// Restore body reader stream
	req.Body = createReadCloser(body)

	// If empty, ignore the match
	if len(body) == 0 && len(ereq.BodyBuffer) != 0 {
		return false, nil
	}

	// Match body by atomic string comparison
	bodyStr := castToString(body)
	matchStr := castToString(ereq.BodyBuffer)
	if bodyStr == matchStr {
		return true, nil
	}

	// Match request body by regexp
	match, _ := regexp.MatchString(matchStr, bodyStr)
	return match, nil
}

func supportedType(req *http.Request) bool {
	mime := req.Header.Get("Content-Type")
	if mime == "" {
		return true
	}

	for _, kind := range BodyTypes {
		if match, _ := regexp.MatchString(kind, mime); match {
			return true
		}
	}
	return false
}

func castToString(buf []byte) string {
	str := string(buf)
	tail := len(str) - 1
	if str[tail] == EOL {
		str = str[:tail]
	}
	return str
}
