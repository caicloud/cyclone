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
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/caicloud/nirvana/definition"
)

// Error is a common interface for error.
// If an error implements the interface, type handlers can
// use Code() to get a specified HTTP status code.
type Error interface {
	// Code is a HTTP status code.
	Code() int
	// Message is an object which contains information of the error.
	Message() interface{}
}

const (
	// HighPriority for error type.
	// If an error occurs, ignore meta and data.
	HighPriority int = 100
	// MediumPriority for meta type.
	MediumPriority int = 200
	// LowPriority for data type.
	LowPriority int = 300
)

// DestinationHandler is used to handle the results from API handlers.
type DestinationHandler interface {
	// Type returns definition.Type which the type handler can handle.
	Destination() definition.Destination
	// Priority returns priority of the type handler. Type handler with higher priority will prior execute.
	Priority() int
	// Validate validates whether the type handler can handle the target type.
	Validate(target reflect.Type) error
	// Handle handles a value. If the handler has something wrong, it should return an error.
	// The handler descides how to deal with value by producers and status code.
	// The status code is a success status code. If everything is ok, the handler should use the status code.
	//
	// There are three cases for return values (goon means go on or continue):
	// 1. go on is true, err is nil.
	//    It means that current type handler did nothing (or looks like did nothing) and next type handler
	//    should take the context.
	// 2. go on is false, err is nil.
	//    It means that current type handler has finished the context and next type handler should not run.
	// 3. err is not nil
	//    It means that current type handler handled the context but something wrong. All subsequent type
	//    handlers should not run.
	Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error)
}

var handlers = map[definition.Destination]DestinationHandler{
	definition.Meta:  &MetaDestinationHandler{},
	definition.Data:  &DataDestinationHandler{},
	definition.Error: &ErrorDestinationHandler{},
}

// DestinationHandlerFor gets a type handler for specified type.
func DestinationHandlerFor(typ definition.Destination) DestinationHandler {
	return handlers[typ]
}

// RegisterDestinationHandler registers a type handler.
func RegisterDestinationHandler(handler DestinationHandler) error {
	handlers[handler.Destination()] = handler
	return nil
}

// MetaDestinationHandler writes metadata to http.ResponseWriter.Header and value type should be map[string]string.
// If value type is not map, the handler will stop the handlers chain and return an error.
// If there is no error, it always expect that the next handler goes on.
type MetaDestinationHandler struct{}

func (h *MetaDestinationHandler) Destination() definition.Destination { return definition.Meta }
func (h *MetaDestinationHandler) Priority() int                       { return MediumPriority }
func (h *MetaDestinationHandler) Validate(target reflect.Type) error  { return nil }
func (h *MetaDestinationHandler) Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error) {
	if value == nil {
		return true, nil
	}
	if values, ok := value.(map[string]string); ok {
		headers := HTTPContextFrom(ctx).ResponseWriter().Header()
		for key, value := range values {
			headers.Set(key, value)
		}
		return true, nil
	}
	return false, invalidMetaType.Error(reflect.TypeOf(value))
}

// DataDestinationHandler writes value to http.ResponseWriter. The type handler handle object value.
// If value is nil, the handler does nothing.
type DataDestinationHandler struct{}

func (h *DataDestinationHandler) Destination() definition.Destination { return definition.Data }
func (h *DataDestinationHandler) Priority() int                       { return LowPriority }
func (h *DataDestinationHandler) Validate(target reflect.Type) error  { return nil }
func (h *DataDestinationHandler) Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error) {
	if value == nil {
		return true, nil
	}
	err = WriteData(ctx, producers, code, value)
	return err == nil, err
}

// ErrorDestinationHandler writes error to http.ResponseWriter.
// If there is no error, the handler does nothing.
type ErrorDestinationHandler struct{}

func (h *ErrorDestinationHandler) Destination() definition.Destination { return definition.Error }
func (h *ErrorDestinationHandler) Priority() int                       { return HighPriority }
func (h *ErrorDestinationHandler) Validate(target reflect.Type) error  { return nil }
func (h *ErrorDestinationHandler) Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error) {
	if value == nil {
		return true, nil
	}
	return false, writeError(ctx, producers, value)
}

func writeError(ctx context.Context, producers []Producer, err interface{}) error {
	httpCtx := HTTPContextFrom(ctx)
	ats, e := AcceptTypes(httpCtx.Request())
	if e != nil {
		return e
	}
	if len(producers) <= 0 {
		return noProducerToWrite.Error(ats)
	}
	code := http.StatusInternalServerError
	msg := interface{}(nil)
	switch e := err.(type) {
	case Error:
		code = e.Code()
		msg = e.Message()
	case error:
		msg = e.Error()
	default:
		msg = err
	}

	producer := chooseProducer(ats, producers)
	if producer == nil {
		// Choose the first producer
		producer = producers[0]
	}
	resp := httpCtx.ResponseWriter()
	if resp.HeaderWritable() {
		// Error always has highest priority. So it can override "Content-Type".
		resp.Header().Set("Content-Type", producer.ContentType())
		resp.WriteHeader(code)
	}
	return producer.Produce(resp, msg)
}

// WriteData chooses right producer by "Accrpt" header and writes data to context.
// You should never call the function except you are writing a type handler.
func WriteData(ctx context.Context, producers []Producer, code int, data interface{}) error {
	httpCtx := HTTPContextFrom(ctx)
	ats, err := AcceptTypes(httpCtx.Request())
	if err != nil {
		return err
	}
	if len(producers) <= 0 {
		return noProducerToWrite.Error(ats)
	}
	producer := chooseProducer(ats, producers)
	if producer == nil {
		return noProducerToWrite.Error(ats)
	}
	resp := httpCtx.ResponseWriter()
	if resp.HeaderWritable() {
		// If "Content-Type" has been set, ignore producer's.
		ctype := resp.Header().Get("Content-Type")
		if strings.TrimSpace(ctype) == "" {
			resp.Header().Set("Content-Type", producer.ContentType())
		}
		resp.WriteHeader(code)
	}
	return producer.Produce(resp, data)
}

func chooseProducer(acceptTypes []string, producers []Producer) Producer {
	if len(acceptTypes) <= 0 || len(producers) <= 0 {
		return nil
	}
	for _, v := range acceptTypes {
		if v == definition.MIMEAll {
			return producers[0]
		}
		for _, p := range producers {
			if p.ContentType() == v {
				return p
			}
		}
	}
	return nil
}
