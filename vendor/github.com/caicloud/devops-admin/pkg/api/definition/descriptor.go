/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package definition

import (
	"github.com/emicklei/go-restful"
)

// Descriptor defines a http api
type Descriptor struct {
	// Path describes a url path that the handler can handle it
	Path string
	// Handlers describes an array of http handlers
	Handlers []Handler
}

// Handler defines a http handler
type Handler struct {

	// HTTPMethod should be a valid http method (eg. GET POST PUT DELETE ...)
	HTTPMethod string

	// Handler provides a handler for current url path
	Handler restful.RouteFunction

	// Filters describes an array of filters
	Filters []restful.FilterFunction

	// Doc provides a short document for describing current descriptor
	Doc string

	// Note is a verbose explanation of the current descriptor
	Note string

	// PathParams describes url params in a request
	PathParams []Param

	// QueryParams describes query params in a request
	QueryParams []Param

	// HeaderParams describes params in request headers
	HeaderParams []Param

	// StatusCode describes all Status Codes from the handler
	StatusCode []StatusCode
}

// Param describes detail infomation of a param
type Param struct {
	// Name provides a name of param
	Name string

	// Type cover the type of current param
	Type string

	// Doc provides a short document of current param
	Doc string

	// Required shows whether the param is required
	Required bool

	// Default holds a default value of current param
	Default interface{}
}

// StatusCode describes detail infomation of a status code
type StatusCode struct {
	// Code provides a status code of http
	Code int

	// Message provides a short message of the current code
	Message string

	// Sample shows a result example for the current code
	Sample interface{}
}
