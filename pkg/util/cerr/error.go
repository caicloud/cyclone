package cerr

import (
	"fmt"
	nerror "github.com/caicloud/nirvana/errors"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// ErrorURLParamNotFound defines request url query param error
	ErrorURLParamNotFound = nerror.BadRequest.Build(ReasonRequest, "can't find parameter ${parameter} in request url query")
	// ErrorHeaderParamNotFound defines request header error
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

	// ErrorUnsupported defines some feature/field not supported yet.
	ErrorUnsupported = nerror.BadRequest.Build("ReasonUnsupported", "unsupported ${resource}: ${type}")
	// ErrorNotImplemented defines some feature not implemented yet.
	ErrorNotImplemented = nerror.InternalServerError.Build("ReasonNotImplemented", "not implement: ${feature}")

	// ErrorCreateIntegration defines error that failed to create integration,
	// this error is used to indicate create control cluster integration failed while creating tenant.
	ErrorCreateIntegration = nerror.InternalServerError.Build("ReasonCreateIntegration",
		"tenant created, but the related control cluster integration created failed: ${error}")
)

// ConvertK8sError converts k8s error to Cyclone errors.
func ConvertK8sError(err error) error {
	if err == nil {
		return nil
	}

	switch t := err.(type) {
	case k8serr.APIStatus:
		details := t.Status().Details
		switch t.Status().Reason {
		case metav1.StatusReasonNotFound:
			return ErrorContentNotFound.Error(fmt.Sprintf("%s %s", details.Kind, details.Name))
		case metav1.StatusReasonConflict, metav1.StatusReasonAlreadyExists:
			return ErrorAlreadyExist.Error(fmt.Sprintf("%s %s", details.Kind, details.Name))
		}
	}

	return ErrorUnknownInternal.Error(err)
}
