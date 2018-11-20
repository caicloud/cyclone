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

package definition

import (
	"context"
	"reflect"
)

// Chain contains all subsequent actions.
type Chain interface {
	// Continue continues to execute the next subsequent actions.
	Continue(context.Context) error
}

// Middleware describes the form of middlewares. If you want to
// carry on, call Chain.Continue() and pass the context.
type Middleware func(context.Context, Chain) error

// Operator is used to operate an object and return an replacement object.
//
// For example:
//  A converter:
//    type ConverterForAnObject struct{}
//    func (c *ConverterForAnObject) Kind() {return "converter"}
//    func (c *ConverterForAnObject) In() reflect.Type {return definition.TypeOf(&ObjectV1{})}
//    func (c *ConverterForAnObject) Out() reflect.Type {return definition.TypeOf(&ObjectV2{})}
//    func (c *ConverterForAnObject) Operate(ctx context.Context, object interface{}) (interface{}, error) {
//        objV2, err := convertObjectV1ToObjectV2(object.(*ObjectV1))
//        return objV2, err
//    }
//
//  A validator:
//    type ValidatorForAnObject struct{}
//    func (c *ValidatorForAnObject) Kind() {return "validator"}
//    func (c *ValidatorForAnObject) In() reflect.Type {return definition.TypeOf(&Object{})}
//    func (c *ValidatorForAnObject) Out() reflect.Type {return definition.TypeOf(&Object{})}
//    func (c *ValidatorForAnObject) Operate(ctx context.Context, object interface{}) (interface{}, error) {
//        if err := validate(object.(*Object)); err != nil {
//            return nil, err
//        }
//        return object, nil
//    }
type Operator interface {
	// Kind indicates operator type.
	Kind() string
	// In returns the type of the only object parameter of operator.
	// The type must be a concrete struct or built-in type. It should
	// not be interface unless it's a file or stream reader.
	In() reflect.Type
	// Out returns the type of the only object result of operator.
	// The type must be a concrete struct or built-in type. It should
	// not be interface unless it's a file or stream reader.
	Out() reflect.Type
	// Operate operates an object and return one.
	Operate(ctx context.Context, field string, object interface{}) (interface{}, error)
}

// Method is an alternative of HTTP method. It's more clearer than HTTP method.
// A definition method binds a certain HTTP method and a success status code.
type Method string

const (
	// List binds to http.MethodGet and code http.StatusOK(200).
	List Method = "List"
	// Get binds to http.MethodGet and code http.StatusOK(200).
	Get Method = "Get"
	// Create binds to http.MethodPost and code http.StatusCreated(201).
	Create Method = "Create"
	// Update binds to http.MethodPut and code http.StatusOK(200).
	Update Method = "Update"
	// Patch binds to http.MethodPatch and code http.StatusOK(200).
	Patch Method = "Patch"
	// Delete binds to http.MethodDelete and code http.StatusNoContent(204).
	Delete Method = "Delete"
	// AsyncCreate binds to http.MethodPost and code http.StatusAccepted(202).
	AsyncCreate Method = "AsyncCreate"
	// AsyncUpdate binds to http.MethodPut and code http.StatusAccepted(202).
	AsyncUpdate Method = "AsyncUpdate"
	// AsyncPatch binds to http.MethodPatch and code http.StatusAccepted(202).
	AsyncPatch Method = "AsyncPatch"
	// AsyncDelete binds to http.MethodDelete and code http.StatusAccepted(202).
	AsyncDelete Method = "AsyncDelete"
)

// Source indicates which place a value is from.
type Source string

const (
	// Path means value is from URL path.
	Path Source = "Path"
	// Query means value is from URL query string.
	Query Source = "Query"
	// Header means value is from request header.
	Header Source = "Header"
	// Form means value is from request body and content type must be
	// "application/x-www-form-urlencoded" and "multipart/form-data".
	Form Source = "Form"
	// File means value is from request body and content type must be
	// "multipart/form-data".
	File Source = "File"
	// Body means value is from request body.
	Body Source = "Body"
	// Auto identifies a struct and generate field values by field tag.
	//
	// Tag name is "source". Its value format is "Source,Name".
	//
	// ex.
	// type Example struct {
	//     Start       int    `source:"Query,start"`
	//     ContentType string `source:"Header,Content-Type"`
	// }
	Auto Source = "Auto"
	// Prefab means value is from a prefab generator.
	// A prefab combines data to generate value.
	Prefab Source = "Prefab"
)

// Destination indicates the target type to place function results.
type Destination string

const (
	// Meta means result will be set into the header of response.
	Meta Destination = "Meta"
	// Data means result will be set into the body of response.
	Data Destination = "Data"
	// Error means the result is an error and should be treated specially.
	// An error occurs indicates that there is no data to return. So the
	// error should be treated as data and be writed back to client.
	Error Destination = "Error"
)

// Example is just an example.
type Example struct {
	// Description describes the example.
	Description string
	// Instance is a custom data.
	Instance interface{}
}

// Parameter describes a function parameter.
type Parameter struct {
	// Source is the parameter value generated from.
	Source Source
	// Name is the name to get value from a request.
	// ex. a query name, a header key, etc.
	Name string
	// Default value is used when a request does not provide a value
	// for the parameter.
	Default interface{}
	// Operators can modify and validate the target value.
	// Parameter value is passed to the first operator, then
	// previous operator's result is as next operator's parameter.
	// The result of last operator will be passed to target function.
	Operators []Operator
	// Description describes the parameter.
	Description string
}

// Result describes how to handle a result from function results.
type Result struct {
	// Destination is the target for the result. Different types make different behavior.
	Destination Destination
	// Operators can modify the result value.
	// Result value is passed to the first operator, then
	// previous operator's result is as next operator's parameter.
	// The result of last operator will be passed to destination handler.
	Operators []Operator
	// Description describes the result.
	Description string
}

// Definition defines an API handler.
type Definition struct {
	// Method is definition method.
	Method Method
	// Consumes indicates how many content types the handler can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates how many content types the handler can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// ErrorProduces is used to generate data for error. If this field is empty,
	// it means that this field equals to Produces.
	// In some cases, succeessful data and error data should be generated in
	// different ways.
	ErrorProduces []string
	// Function is a function handler. It must be func type.
	Function interface{}
	// Parameters describes function parameters.
	Parameters []Parameter
	// Results describes function retrun values.
	Results []Result
	// Summary is a one-line brief description of this definition.
	Summary string
	// Description describes the API handler.
	Description string
	// Examples contains many examples for the API handler.
	Examples []Example
}

// Descriptor describes a descriptor for API definitions.
type Descriptor struct {
	// Path is the url path. It will inherit parent's path.
	//
	// If parent path is "/api/v1", current is "/some",
	// It means current definitions handles "/api/v1/some".
	Path string
	// Consumes indicates content types that current definitions
	// and child definitions can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates content types that current definitions
	// and child definitions can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Middlewares contains path middlewares.
	Middlewares []Middleware
	// Definitions contains definitions for current path.
	Definitions []Definition
	// Children is used to place sub-descriptors.
	Children []Descriptor
	// Description describes the usage of the path.
	Description string
}
