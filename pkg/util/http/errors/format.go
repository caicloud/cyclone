/*
Copyright 2017 caicloud authors. All rights reserved.

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

package errors

import (
	nerror "github.com/caicloud/nirvana/errors"
)

// defines reason types
const (
	// ReasonInternal is a type about internal errors
	ReasonInternal = "ReasonInternal"
	// ReasonRequest is a type about request errors
	ReasonRequest = "ReasonRequest"
)

var (
	// ErrorParamTypeError defines param type error
	ErrorParamTypeError = nerror.BadRequest.Build(ReasonRequest, "parameter ${parameter} should be ${expect}, but got ${real}")
	// ErrorParamNotFound defines request param error
	ErrorParamNotFound = nerror.BadRequest.Build(ReasonRequest, "can't find parameter ${parameter} in request")
	// ErrorUrlParamNotFound defines request url query param error
	ErrorUrlParamNotFound = nerror.BadRequest.Build(ReasonRequest, "can't find parameter ${parameter} in request url query")
	// ErrorHeaderNotFound defines request header error
	ErrorHeaderParamNotFound = nerror.BadRequest.Build(ReasonRequest, "can't find parameter ${parameter} in request header")

	// ErrorValidationFailed defines validation failed error
	ErrorValidationFailed = nerror.BadRequest.Build(ReasonRequest, "failed to validate ${field}: ${error}")
	// ErrorContentNotFound defines not found error
	ErrorContentNotFound = nerror.NotFound.Build(ReasonRequest, "content ${content} not found")
	// ErrorQuotaExceeded defines quota exceeded error, creating or updating was not allowed
	ErrorQuotaExceeded = nerror.Forbidden.Build(ReasonRequest, "${resource} quota exceeded")
	// ErrorAlreadyExist defines conflict error.
	ErrorAlreadyExist = nerror.Conflict.Build(ReasonRequest, "conflict: ${resource} already exist")

	// ErrorAuthenticationRequired defines error that authentication not provided.
	ErrorAuthenticationRequired = nerror.Unauthorized.Build(ReasonRequest, "authentication required")

	// ErrorInternalTypeError defines internal type error
	//ErrorInternalTypeError = nerror.InternalServerError.Build(ReasonInternal, "type of ${resource} should be ${expect}, but got ${real}")

	// ErrorUnknownNotFoundError defines not found error that we can't find a reason
	ErrorUnknownNotFoundError = nerror.InternalServerError.Build(ReasonInternal, "content ${content} not found, may be it's a serious error")
	// ErrorUnknownInternal defines any internal error and not one of above errors
	ErrorUnknownInternal = nerror.InternalServerError.Build(ReasonInternal, "unknown error: ${error}")

	// ErrorCreateFailed defines error that failed creating of something.
	ErrorCreateFailed = nerror.InternalServerError.Build(ReasonInternal, "failed to create ${name}: ${error}")
	// ErrorUpdateFailed defines error that failed updating of something.
	ErrorUpdateFailed = nerror.InternalServerError.Build(ReasonInternal, "failed to update ${name}: ${error}")
	// ErrorDeleteFailed defines error that failed deleting of something.
	ErrorDeleteFailed = nerror.InternalServerError.Build(ReasonInternal, "failed to delete ${name}: ${error}")
	// ErrorGetFailed defines error that failed geting of something.
	ErrorGetFailed = nerror.InternalServerError.Build(ReasonInternal, "failed to get ${name}: ${error}")
	// ErrorListFailed defines error that failed listing of something.
	ErrorListFailed = nerror.InternalServerError.Build(ReasonInternal, "failed to list ${name}: ${error}")

	// ErrorCreateWebhookPermissionDenied defines error that failed creating webhook as permission denied.
	ErrorCreateWebhookPermissionDenied = nerror.InternalServerError.Build("ReasonCreateWebhookPermissionDenied",
		"failed to create webhook of pipeline ${pipeline}, please check your account permissions.")

	// ErrorPRNotFound defines error that failed getting GitHub/GitLab pr as permission denied or not exist.
	ErrorPRNotFound = nerror.InternalServerError.Build("ReasonPRNotFound",
		"failed to get pull request ${id} of project ${project}, please check if it exists.")

	// ErrorUnsupported defines some feature/field not supported yet.
	ErrorUnsupported = nerror.BadRequest.Build("ReasonUnsupported", "unsupported ${resource}: ${type}")
	// ErrorNotImplemented defines some feature not implemented yet.
	ErrorNotImplemented = nerror.InternalServerError.Build("ReasonNotImplemented", "not implement: ${feature}")
)

func NewCreateError(name, errorMessage string) error {
	return ErrorCreateFailed.Error(name, errorMessage)
}

func NewUpdateError(name, errorMessage string) error {
	return ErrorUpdateFailed.Error(name, errorMessage)
}

func NewDeleteError(name, errorMessage string) error {
	return ErrorDeleteFailed.Error(name, errorMessage)
}

func NewListError(name, errorMessage string) error {
	return ErrorListFailed.Error(name, errorMessage)
}

func NewGetError(name, errorMessage string) error {
	return ErrorGetFailed.Error(name, errorMessage)
}

// NewValidateError creates new ErrorValidationFailed
func NewValidateError(name, reason string) error {
	return ErrorValidationFailed.Error(name, reason)
}

func NewNotFoundError(name string) error {
	return ErrorContentNotFound.Error(name)
}
