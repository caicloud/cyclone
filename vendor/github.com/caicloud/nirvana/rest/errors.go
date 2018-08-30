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
	"github.com/caicloud/nirvana/errors"
)

var (
	unrecognizedHTTPScheme  = errors.InternalServerError.Build("Nirvana:REST:UnrecognizedHTTPScheme", "unrecognized http scheme ${scheme}")
	noHTTPHost              = errors.InternalServerError.Build("Nirvana:REST:NoHTTPHost", "no http host")
	invalidPath             = errors.InternalServerError.Build("Nirvana:REST:InvalidPath", "invalid path ${path}: ${reason}")
	duplicatedPathParameter = errors.InternalServerError.Build("Nirvana:REST:DuplicatedPathParameter", "duplicated path patameter ${param} in path ${path}")
	noPathParameter         = errors.InternalServerError.Build("Nirvana:REST:NoPathParameter", "no path parameter ${parameter} in path ${path}")
	conflictBodyParameter   = errors.InternalServerError.Build("Nirvana:REST:ConflictBodyParameter", "conflict body parameter in path ${path}")
	unconvertibleObject     = errors.InternalServerError.Build("Nirvana:REST:UnconvertibleObject", "can't write object ${object} for path ${path}: ${reason}")
	unwritableForm          = errors.InternalServerError.Build("Nirvana:REST:UnwritableForm", "can't write form ${form} of path ${path}: ${reason}")
	unwritableFile          = errors.InternalServerError.Build("Nirvana:REST:UnwritableFile", "can't write file ${file} of path ${path}: ${reason}")
	unwritableForms         = errors.InternalServerError.Build("Nirvana:REST:UnwritableForms", "can't write forms of path ${path}: ${reason}")
	unreadableBody          = errors.InternalServerError.Build("Nirvana:REST:UnreadableBody", "can't read body of path ${path}: ${reason}")
	unrecognizedBody        = errors.InternalServerError.Build("Nirvana:REST:UnrecognizedBody", "can't recognize body parameter of path ${path}: ${reason}")
	invalidRequest          = errors.InternalServerError.Build("Nirvana:REST:InvalidRequest", "can't create request for path ${path}: ${reason}")
	invalidContentType      = errors.InternalServerError.Build("Nirvana:REST:InvalidContentType", "can't parse content type ${type} for path ${path}: ${reason}")
	unmatchedStatusCode     = errors.InternalServerError.Build("Nirvana:REST:UnmatchedStatusCode", "desired code of path ${path} is ${desired} but got ${current}")
)

// IsRESTError checks if an error is generated from this package.
func IsRESTError(err error) bool {
	switch {
	case unrecognizedHTTPScheme.Derived(err):
	case noHTTPHost.Derived(err):
	case invalidPath.Derived(err):
	case duplicatedPathParameter.Derived(err):
	case noPathParameter.Derived(err):
	case conflictBodyParameter.Derived(err):
	case unconvertibleObject.Derived(err):
	case unwritableForm.Derived(err):
	case unwritableFile.Derived(err):
	case unwritableForms.Derived(err):
	case unreadableBody.Derived(err):
	case unrecognizedBody.Derived(err):
	case invalidRequest.Derived(err):
	case invalidContentType.Derived(err):
	case unmatchedStatusCode.Derived(err):
	default:
		return false
	}
	return true
}
