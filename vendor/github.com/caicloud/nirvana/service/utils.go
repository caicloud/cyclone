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
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
)

// WrapHTTPHandler wraps an http handler to definition function.
func WrapHTTPHandler(h http.Handler) func(ctx context.Context) {
	return func(ctx context.Context) {
		httpCtx := HTTPContextFrom(ctx)
		h.ServeHTTP(httpCtx.ResponseWriter(), httpCtx.Request())
	}
}

// WrapHTTPHandlerFunc wraps an http handler func to definition function.
func WrapHTTPHandlerFunc(f http.HandlerFunc) func(ctx context.Context) {
	return func(ctx context.Context) {
		httpCtx := HTTPContextFrom(ctx)
		f(httpCtx.ResponseWriter(), httpCtx.Request())
	}
}

// FileNotFound is an error factory to show why can't find a file.
// This error may contains private information. Don't return this error to end users directly.
var FileNotFound = errors.NotFound.Build("Nirvana:Service:FileNotFound", "can't find file ${path} because ${reason}")

// FileForbidden is an error factory to show why can't access a file.
// This error may contains private information. Don't return this error to end users directly.
var FileForbidden = errors.Forbidden.Build("Nirvana:Service:FileForbidden", "can't access file ${path} because ${reason}")

// UnreadableFile is an error factory to show why can't read a file.
// This error may contains private information. Don't return this error to end users directly.
var UnreadableFile = errors.InternalServerError.Build("Nirvana:Service:UnreadableFile", "can't read file ${path} because ${reason}")

// UnseekableFile is an error factory to show why can't seek a file.
// This error may contains private information. Don't return this error to end users directly.
var UnseekableFile = errors.InternalServerError.Build("Nirvana:Service:UnseekableFile", "can't seek file ${path} because ${reason}")

// ReadFile reads file and returns mime type.
func ReadFile(path string) (string, io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return "", nil, FileNotFound.Error(path, err)
		case os.IsPermission(err):
			return "", nil, FileForbidden.Error(path, err)
		}
		return "", nil, UnreadableFile.Error(path, err)
	}
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		// Read a chunk to decide between utf-8 text and binary
		var buf [512]byte
		n, _ := io.ReadFull(file, buf[:])
		ctype = http.DetectContentType(buf[:n])
		_, err := file.Seek(0, io.SeekStart)
		if err != nil {
			return "", nil, UnseekableFile.Error(path, err)
		}
	}
	if ctype == "" {
		ctype = definition.MIMEOctetStream
	}
	return ctype, file, nil
}

// MetaForContentType returns a meta for content type.
func MetaForContentType(ctype string) map[string]string {
	return map[string]string{"Content-Type": ctype}
}

// Internal error factories:
var noExecutorForMethod = errors.MethodNotAllowed.Build("Nirvana:Service:NoExecutorForMethod", "method not allowed")
var noExecutorForContentType = errors.UnsupportedMediaType.Build("Nirvana:Service:NoExecutorForContentType", "unsupported media type")
var noExecutorToProduce = errors.NotAcceptable.Build("Nirvana:Service:NoExecutorToProduce", "not acceptable")
var invalidContentType = errors.BadRequest.Build("Nirvana:Service:InvalidContentType", "invalid content type ${type}")
var invalidConversion = errors.BadRequest.Build("Nirvana:Service:InvalidConversion", "can't convert ${data} to ${type}")

var noRouter = errors.InternalServerError.Build("Nirvana:Service:NoRouter", "no router to build service")
var invalidService = errors.InternalServerError.Build("Nirvana:Service:NoResponse", "no response")
var invalidConsumer = errors.InternalServerError.Build("Nirvana:Service:InvalidConsumer", "${type} is invalid for consumer")
var invalidProducer = errors.InternalServerError.Build("Nirvana:Service:InvalidProducer", "${type} is invalid for producer")
var noConnectionHijacker = errors.InternalServerError.Build("Nirvana:Service:NoConnectionHijacker",
	"underlying http.ResponseWriter does not implement http.Hijacker")
var definitionNoMethod = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoMethod", "no http method in [${method}]${path}")
var definitionNoConsumes = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoConsumes", "no content type to consume in [${method}]${path}")
var definitionNoProduces = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoProduces", "no content type to produce in [${method}]${path}")
var definitionNoErrorProduces = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoErrorProduces",
	"no content type to produce error in [${method}]${path}")
var definitionNoFunction = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoFunction", "no function in [${method}]${path}")
var definitionInvalidFunctionType = errors.InternalServerError.Build("Nirvana:Service:DefinitionInvalidFunctionType",
	"${type} is not function in [${method}]${path}")

var definitionNoConsumer = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoConsumer",
	"no consumer for content type ${type} in [${method}]${path}")

var definitionNoProducer = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoProducer",
	"no producer for content type ${type} in [${method}]${path}")

var definitionConflict = errors.InternalServerError.Build("Nirvana:Service:DefinitionConflict",
	"consumer-producer pair ${key}:${value} conflicts in [http.${method}]${path}")

var definitionUnmatchedParameters = errors.InternalServerError.Build("Nirvana:Service:DefinitionUnmatchedParameters",
	"function ${function} has ${count} parameters but want ${desired} in ${path}, "+
		"you can define it with descriptor->definition[]->parameters[]")

var definitionUnmatchedResults = errors.InternalServerError.Build("Nirvana:Service:DefinitionUnmatchedResults",
	"function ${function} has ${count} results but want ${desired} in ${path}, "+
		"you can define it with descriptor->definition[]->results[]")

var noDestinationHandler = errors.InternalServerError.Build("Nirvana:Service:NoDestinationHandler", "no destination handler for destination ${destination}, "+
	"you can define it with descriptor->definition[]->results[]->destination")

var noContext = errors.InternalServerError.Build("Nirvana:Service:NoContext", "can't find http context, "+
	"you should define `ctx context.Context` as the first parameter of your handler function")

var requiredField = errors.InternalServerError.Build("Nirvana:Service:RequiredField", "required field ${field} in ${source} but got empty")
var invalidMetaType = errors.InternalServerError.Build("Nirvana:Service:InvalidMetaType", "can't recognize meta for type ${type}")
var noProducerToWrite = errors.NotAcceptable.Build("Nirvana:Service:NoProducerToWrite", "can't find producer for accept types ${types}")
var invalidMethod = errors.InternalServerError.Build("Nirvana:Service:InvalidMethod", "http method ${method} is invalid")
var invalidStatusCode = errors.InternalServerError.Build("Nirvana:Service:InvalidStatusCode", "http status code must be in [100,599]")
var unassignableType = errors.InternalServerError.Build("Nirvana:Service:UnassignableType", "type ${typeA} can't assign to ${typeB}")
var noConverter = errors.InternalServerError.Build("Nirvana:Service:UnassignableType", "no converter for type ${type}")
var invalidBodyType = errors.InternalServerError.Build("Nirvana:Service:InvalidBodyType", "${type} is not a valid type for body")
var noPrefab = errors.InternalServerError.Build("Nirvana:Service:NoPrefab", "no prefab named ${name}")
var invalidAutoParameter = errors.InternalServerError.Build("Nirvana:Service:InvalidAutoParameter", "${type} is not a struct or a pointer to struct")
var noParameterGenerator = errors.InternalServerError.Build("Nirvana:Service:NoParameterGenerator", "no parameter generator for source ${source}")
var invalidFieldTag = errors.InternalServerError.Build("Nirvana:Service:InvalidFieldTag", "filed tag ${tag} is invalid")
var noName = errors.InternalServerError.Build("Nirvana:Service:NoName", "${source} must have a name")
var invalidTypeForConsumer = errors.InternalServerError.Build("Nirvana:Service:InvalidTypeForConsumer",
	"consumer ${content} can't consume data for type ${type}")
var invalidTypeForProducer = errors.InternalServerError.Build("Nirvana:Service:InvalidTypeForProducer",
	"producer ${content} can't produce data for type ${type}")
var invalidOperatorInType = errors.InternalServerError.Build("Nirvana:Service:InvalidOperatorInType",
	"the type ${type} is not compatible to the in type of the ${index} operator")
var invalidOperatorOutType = errors.InternalServerError.Build("Nirvana:Service:InvalidOperatorOutType",
	"the out type of the ${index} operator is not compatible to the type ${type}")
