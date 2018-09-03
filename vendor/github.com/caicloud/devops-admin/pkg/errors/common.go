/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package errors

import (
	"net/http"
)

// defines reason types
const (
	// ReasonInternal is a type about internal errors
	ReasonInternal = "ReasonInternal"
	// ReasonRequest is a type about request errors
	ReasonRequest = "ReasonRequest"
)

var (
	// ErrorUnknownRequest defines unknown error when failed to handle a request
	ErrorUnknownRequest = NewFormatError(http.StatusBadRequest, ReasonRequest, "%s")
	// ErrorParamTypeError defines param type error
	ErrorParamTypeError = NewFormatError(http.StatusBadRequest, ReasonRequest, "param %s should be %s, but got %s")
	// ErrorParamNotFound defines request param error
	ErrorParamNotFound = NewFormatError(http.StatusBadRequest, ReasonRequest, "can't find param %s in request")
	// ErrorHeaderParamNotFound defines request header error
	ErrorHeaderParamNotFound = NewFormatError(http.StatusForbidden, ReasonRequest, "can't find param %s in request header")
	// ErrorValidationFailed defines validation failed error
	ErrorValidationFailed = NewFormatError(http.StatusBadRequest, ReasonRequest, "failed to validate %s: %s")
	// ErrorContentNotFound defines not found error
	ErrorContentNotFound = NewFormatError(http.StatusNotFound, ReasonRequest, "content %s not found")
	// ErrorQuotaExceeded defines quota exceeded error, creating or updating was not allowed
	ErrorQuotaExceeded          = NewFormatError(http.StatusForbidden, ReasonRequest, "%s quota exceeded")
	ErrorAlreadyExist           = NewFormatError(http.StatusConflict, ReasonRequest, "conflict: %s already exist")
	ErrorAuthenticationRequired = NewFormatError(http.StatusProxyAuthRequired, ReasonRequest, "authentication required")

	// ErrorInternalTypeError defines internal type error
	ErrorInternalTypeError = NewFormatError(http.StatusInternalServerError, ReasonInternal, "type of %s should be %s, but got %s")
	// ErrorUnknownNotFoundError defines not found error that we can't find a reason
	ErrorUnknownNotFoundError = NewFormatError(http.StatusInternalServerError, ReasonInternal, "content %s not found, may be it's a serious error")
	// ErrorUnknownInternal defines any internal error and not one of above errors
	ErrorUnknownInternal = NewFormatError(http.StatusBadRequest, ReasonInternal, "unknown error: %s")

	// ErrorCreateFailed defines error that failed creating of something.
	ErrorCreateFailed = NewFormatError(http.StatusBadRequest, ReasonInternal, "failed to create %s: %s")
	// ErrorUpdateFailed defines error that failed updating of something.
	ErrorUpdateFailed = NewFormatError(http.StatusBadRequest, ReasonInternal, "failed to update %s: %s")
	// ErrorDeleteFailed defines error that failed deleting of something.
	ErrorDeleteFailed = NewFormatError(http.StatusBadRequest, ReasonInternal, "failed to delete %s: %s")
	// ErrorGetFailed defines error that failed geting of something.
	ErrorGetFailed = NewFormatError(http.StatusBadRequest, ReasonInternal, "failed to get %s: %s")
	// ErrorListFailed defines error that failed listing of something.
	ErrorListFailed = NewFormatError(http.StatusBadRequest, ReasonInternal, "failed to list %s: %s")
)

func NewCreateError(name, errorMessage string) error {
	return ErrorCreateFailed.Format(name, errorMessage)
}

func NewUpdateError(name, errorMessage string) error {
	return ErrorUpdateFailed.Format(name, errorMessage)
}

func NewDeleteError(name, errorMessage string) error {
	return ErrorDeleteFailed.Format(name, errorMessage)
}

func NewListError(name, errorMessage string) error {
	return ErrorListFailed.Format(name, errorMessage)
}

func NewGetError(name, errorMessage string) error {
	return ErrorGetFailed.Format(name, errorMessage)
}

// NewValidateError creates new ErrorValidationFailed
func NewValidateError(name, reason string) error {
	return ErrorValidationFailed.Format(name, reason)
}

func NewNotFoundError(name string) error {
	return ErrorContentNotFound.Format(name)
}
