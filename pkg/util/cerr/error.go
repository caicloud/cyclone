package cerr

import (
	"fmt"
	"strings"

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
	// ReasonCreateWebhookPermissionDenied is a type of errors about creating webhook
	ReasonCreateWebhookPermissionDenied = "ReasonCreateWebhookPermissionDenied"
	// ReasonExternalSystemError is a type of errors that occurred in external system
	ReasonExternalSystemError = "ReasonExternalSystemError"
	// ReasonPRNotFound is a type of errors about pull request not found
	ReasonPRNotFound = "ReasonPRNotFound"
	// ReasonAuthorizationFailed represents authorization error
	ReasonAuthorizationFailed = "ReasonAuthorizationFailed"
	// ReasonAuthenticationFailed represents authentication error
	ReasonAuthenticationFailed = "ReasonAuthenticationFailed"
	// ReasonNotFound represents not found error
	ReasonNotFound = "ReasonNotFound"
	// ReasonConnectionRefused represents connection refused error
	ReasonConnectionRefused = "ReasonConnectionRefused"
	// ReasonNoSuchHost represents no such host error
	ReasonNoSuchHost = "ReasonNoSuchHost"
	// ReasonIOTimeout represents io timeout error
	ReasonIOTimeout = "ReasonIOTimeout"
	// ReasonExistRunningWorkflows represents error that update persisten volume while there are workflows running.
	ReasonExistRunningWorkflows = "ReasonExistRunningWorkflows"
	// ReasonCreateIntegrationFailed represents creating integration failed error
	ReasonCreateIntegrationFailed = "ReasonCreateIntegrationFailed"
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
	// ErrorQueryParamNotCorrect defines request query param error
	ErrorQueryParamNotCorrect = nerror.BadRequest.Build(ReasonRequest, "bad request as query ${parameter} is not correct")

	// ErrorIntegrationTypeNotCorrect defines wrong integration type error
	ErrorIntegrationTypeNotCorrect = nerror.BadRequest.Build(ReasonRequest, "type of ${integration} should be ${expect}, but got ${real}")

	// ErrorValidationFailed defines validation failed error
	ErrorValidationFailed = nerror.BadRequest.Build(ReasonRequest, "failed to validate ${field}: ${error}")
	// ErrorContentNotFound defines not found error
	ErrorContentNotFound = nerror.NotFound.Build(ReasonRequest, "content ${content} not found")
	// ErrorQuotaExceeded defines quota exceeded error, creating or updating was not allowed
	ErrorQuotaExceeded = nerror.Forbidden.Build(ReasonRequest, "${resource} quota exceeded")
	// ErrorClusterNotClosed defines error that represents some operations are forbidden while cluster is not closed
	ErrorClusterNotClosed = nerror.Forbidden.Build(ReasonRequest, "should close cluster integration ${integration} firstly")
	// ErrorAlreadyExist defines conflict error.
	ErrorAlreadyExist = nerror.Conflict.Build(ReasonRequest, "conflict: ${resource} already exist")

	// ErrorAuthorizationRequired defines error that authorization not provided.
	ErrorAuthorizationRequired = nerror.Unauthorized.Build(ReasonRequest, "authorization required")

	// ErrorAuthorizationFailed defines error that authorization failed.
	ErrorAuthorizationFailed = nerror.Unauthorized.Build(ReasonRequest, "authorization failed")

	// ErrorAuthenticationFailed defines error that authentication failed.
	ErrorAuthenticationFailed = nerror.Forbidden.Build(ReasonRequest, "authentication failed")

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

	// ErrorUnsupported defines some feature/field not supported yet.
	ErrorUnsupported = nerror.BadRequest.Build("ReasonUnsupported", "unsupported ${resource}: ${type}")
	// ErrorNotImplemented defines some feature not implemented yet.
	ErrorNotImplemented = nerror.InternalServerError.Build("ReasonNotImplemented", "not implement: ${feature}")

	// ErrorCreateIntegrationFailed defines error that failed to create integration.
	ErrorCreateIntegrationFailed = nerror.InternalServerError.Build(ReasonCreateIntegrationFailed,
		"create integration ${name} failed: ${error}, please check your auth infomation")

	// ErrorCreateWebhookPermissionDenied defines error that failed creating webhook as permission denied.
	ErrorCreateWebhookPermissionDenied = nerror.InternalServerError.Build(ReasonCreateWebhookPermissionDenied,
		"failed to create webhook, please check your account permissions.")

	// ErrorExternalSystemError defines error that occurred in external system (GitHub GitLab SVN SonarQube) server side.
	ErrorExternalSystemError = nerror.InternalServerError.Build(ReasonExternalSystemError,
		"External system ${system-type} internal error: ${error}, Maybe the external server is not running well, Please contact the administrator of the external system if this problem persists.")

	// ErrorPRNotFound defines error that failed creating webhook as permission denied.
	ErrorPRNotFound = nerror.InternalServerError.Build(ReasonPRNotFound,
		"failed to find the PR ${id} in your SCM server ${server}, please check if it exists.")

	// ErrorExternalAuthorizationFailed defines error that authorization failed for external system, Unauthorized 401.
	ErrorExternalAuthorizationFailed = nerror.InternalServerError.Build(ReasonAuthorizationFailed,
		"authentication failed: ${error}")

	// ErrorExternalAuthenticationFailed defines error that authentication failed for external system, Forbidden 403.
	ErrorExternalAuthenticationFailed = nerror.InternalServerError.Build(ReasonAuthenticationFailed,
		"authentication failed: ${error}")

	// ErrorExternalNotFound defines error that not found for external system, NotFound 404.
	ErrorExternalNotFound = nerror.InternalServerError.Build(ReasonNotFound, "not found: ${error}")

	// ErrorExternalConnectionRefused defines error that connection refused while sending requests to external system.
	ErrorExternalConnectionRefused = nerror.InternalServerError.Build(ReasonConnectionRefused, "connection refused: ${error}")

	// ErrorExternalNoSuchHost defines error that no such host while sending requests to external system.
	ErrorExternalNoSuchHost = nerror.InternalServerError.Build(ReasonNoSuchHost, "no such host: ${error}")

	// ErrorExternalIOTimeout defines error that io timeout while sending requests to external system.
	ErrorExternalIOTimeout = nerror.InternalServerError.Build(ReasonIOTimeout, "i/o timeout: ${error}")

	// ErrorExistRunningWorkflows defines error that can not update persisten volume while there are workflows running.
	ErrorExistRunningWorkflows = nerror.InternalServerError.Build(ReasonExistRunningWorkflows,
		"can not update persistent volume, since there are workflows running, need to stop following workflows firstly: ${workflows}")
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

// AutoAnalyse analyses if an error belongs to a concrete type error,
// Yes, will translate it to the type;
// No, return it originally.
func AutoAnalyse(err error) error {
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "dial tcp") && strings.Contains(err.Error(), "connect: connection refused") {
		return ErrorExternalConnectionRefused.Error(err)
	}

	if strings.Contains(err.Error(), "dial tcp") && strings.Contains(err.Error(), "no such host") {
		return ErrorExternalNoSuchHost.Error(err)
	}

	if strings.Contains(err.Error(), "dial tcp") && strings.Contains(err.Error(), "i/o timeout") {
		return ErrorExternalIOTimeout.Error(err)
	}

	return err
}
