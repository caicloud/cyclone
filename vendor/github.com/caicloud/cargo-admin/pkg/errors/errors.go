package errors

import (
	"net/http"

	"github.com/caicloud/nirvana/errors"
)

// defines reason types
const (
	// ReasonInternal is a type about internal errors
	ReasonInternal = "ReasonInternal"
	// ReasonRequest is a type about request errors
	ReasonRequest = "ReasonRequest"
)

var (
	ErrorUnknownRequest      = errors.BadRequest.Build(ReasonRequest, "${msg}")
	ErrorContentNotFound     = errors.NotFound.Build(ReasonInternal, "content ${content} not found")
	ErrorSystemTenantAllowed = errors.Forbidden.Build(ReasonRequest, "only system tenant users allowed ${action}")
	ErrorNotAllowed          = errors.Forbidden.Build(ReasonRequest, "${handle}")
	ErrorAlreadyExist        = errors.Conflict.Build(ReasonRequest, "conflict: ${resource} already exist")
	ErrorProjectAlreadyExist = errors.Conflict.Build("cargo:ProjectAlreadyExist", "${project} project already exist")

	ErrorUnauthentication = errors.Unauthorized.Build(ReasonRequest, http.StatusText(http.StatusUnauthorized))

	ErrorUnknownInternal = errors.BadRequest.Build(ReasonInternal, "unknown error: ${msg}")

	// ErrorCreateFailed defines error that failed creating of something.
	ErrorCreateFailed = errors.BadRequest.Build(ReasonInternal, "failed to create ${value1}: ${value2}")
	// ErrorUpdateFailed defines error that failed updating of something.
	ErrorUpdateFailed = errors.BadRequest.Build(ReasonInternal, "failed to update ${value1}: ${value2}")
	// ErrorDeleteFailed defines error that failed deletinwg of something.
	ErrorDeleteFailed = errors.BadRequest.Build(ReasonInternal, "failed to delete ${value1}: ${value2}")
	// ErrorGetFailed defines error that failed geting of something.
	ErrorGetFailed = errors.BadRequest.Build(ReasonInternal, "failed to get ${value1}: ${value2}")
	// ErrorListFailed defines error that failed listing of something.
	ErrorListFailed = errors.BadRequest.Build(ReasonInternal, "failed to list ${value1}: ${value2}")
)
