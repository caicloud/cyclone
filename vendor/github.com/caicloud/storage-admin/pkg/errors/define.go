package errors

import (
	"fmt"
	"net/http"
)

var (
	ErrVarKubeClientNil = fmt.Errorf("kubernetes client is nil")
	ErrVarCdsHostEmpty  = fmt.Errorf("cds host is empty")
)

const (
	ReasonGroupBase    = "resource:"
	ReasonGroupStorage = "storage." + ReasonGroupBase

	// inner use
	ErrorReasonUnknownStorageClassStatus   = ReasonGroupStorage + "UnknownStorageClassStatus"
	ErrorReasonUnknownStorageServiceStatus = ReasonGroupStorage + "UnknownStorageServiceStatus"
	// bad request TODO merge to one?
	ErrorReasonBadPageStartOrLimit    = ReasonGroupStorage + "BadPageStartOrLimit"
	ErrorReasonBadRequestBody         = ReasonGroupStorage + "BadRequestBody"
	ErrorReasonTypeNotFound           = ReasonGroupStorage + "TypeNotFound"
	ErrorReasonServiceStatusNotActive = ReasonGroupStorage + "ServiceStatusNotActive"
	ErrorReasonObjectAlreadyExist     = ReasonGroupStorage + "ObjectAlreadyExist"
	ErrorReasonObjectNotFound         = ReasonGroupStorage + "ObjectNotFound"
	ErrorReasonObjectBadName          = ReasonGroupStorage + "ObjectBadName"
	ErrorReasonStorageSecretNotFound  = ReasonGroupStorage + "StorageSecretNotFound"
	ErrorReasonMissParameter          = ReasonGroupStorage + "MissParameter"
	ErrorReasonClusterNotFound        = ReasonGroupStorage + "ClusterNotFound"
	ErrorReasonClassStatusNotActive   = ReasonGroupStorage + "ClassStatusNotActive"
	ErrorReasonMapParameterNotMatch   = ReasonGroupStorage + "MapParameterNotMatch"
	ErrorReasonClassNameTooLong       = ReasonGroupStorage + "ClassNameTooLong"
	ErrorReasonSystemObject           = ReasonGroupStorage + "SystemObject"
	ErrorReasonObjInSystemNamespace   = ReasonGroupStorage + "ObjectInSystemNamespace"
	// other error
	ErrorReasonAuthFailed          = ReasonGroupStorage + "AuthFailed"
	ErrorReasonInternalServerError = ReasonGroupStorage + "InternalServerError"
	ErrorReasonClusterBadConfig    = ReasonGroupStorage + "ClusterBadConfig"
	// admission
	ErrorReasonQuotaNotComplete = ReasonGroupStorage + "QuotaNotComplete"
	ErrorReasonQuotaExceeded    = ReasonGroupStorage + "QuotaExceeded"
)

// inner use, no http code
func (fe *FormatError) SetErrorUnknownStorageClassStatus(status interface{}) *FormatError {
	fe.Message = fmt.Sprintf("Unknown StorageClass Status %v", status)
	fe.Reason = ErrorReasonUnknownStorageClassStatus
	fe.Code = http.StatusBadRequest
	return fe
}
func (fe *FormatError) SetErrorUnknownStorageServiceStatus(status interface{}) *FormatError {
	fe.Message = fmt.Sprintf("Unknown StorageService Status %v", status)
	fe.Reason = ErrorReasonUnknownStorageServiceStatus
	fe.Code = http.StatusBadRequest
	return fe
}

// bad request
func (fe *FormatError) SetErrorBadPageStartOrLimit(start, limit string) *FormatError {
	fe.Message = fmt.Sprintf("bad start or limit in query parameters start=%s, limit=%s", start, limit)
	fe.Reason = ErrorReasonBadPageStartOrLimit
	fe.Code = http.StatusBadRequest
	return fe
}

func (fe *FormatError) SetErrorBadRequest(e error) *FormatError {
	fe.Message = fmt.Sprintf("bad request: %v", e)
	fe.Reason = ErrorReasonBadRequestBody
	fe.Code = http.StatusBadRequest
	fe.SetRawError(e)
	return fe
}

func (fe *FormatError) SetErrorBadRequestBody(e error) *FormatError {
	fe.Message = fmt.Sprintf("parse request body failed")
	fe.Reason = ErrorReasonBadRequestBody
	fe.Code = http.StatusBadRequest
	fe.SetRawError(e)
	return fe
}

func (fe *FormatError) SetErrorTypeNotFound(typeName string) *FormatError {
	fe.Message = fmt.Sprintf("storage type %s not found", typeName)
	fe.Reason = ErrorReasonTypeNotFound
	fe.Code = http.StatusNotFound
	return fe
}

func (fe *FormatError) SetErrorServiceStatusNotActive(status string) *FormatError {
	fe.Message = fmt.Sprintf("service status %s not active", status)
	fe.Reason = ErrorReasonServiceStatusNotActive
	fe.Code = http.StatusBadRequest
	return fe
}

func (fe *FormatError) SetErrorObjectAlreadyExist(name string, e error) *FormatError {
	fe.Message = fmt.Sprintf("object %s already exist", name)
	fe.Reason = ErrorReasonObjectAlreadyExist
	fe.Code = http.StatusBadRequest
	fe.SetRawError(e)
	return fe
}

func (fe *FormatError) SetErrorObjectNotFound(name string, e error) *FormatError {
	fe.Message = fmt.Sprintf("object %s not found", name)
	fe.Reason = ErrorReasonObjectNotFound
	fe.Code = http.StatusNotFound
	fe.SetRawError(e)
	return fe
}

func (fe *FormatError) SetErrorObjectBadName(name string, e error) *FormatError {
	fe.Message = fmt.Sprintf("object name %s is in bad format, %v", name, e)
	fe.Reason = ErrorReasonObjectBadName
	fe.Code = http.StatusBadRequest
	fe.SetRawError(e)
	return fe
}

func (fe *FormatError) SetErrorStorageSecretNotFound(secretNamespace, secretName string, e error) *FormatError {
	fe.Message = fmt.Sprintf("storage related secret %s/%s not found", secretNamespace, secretName)
	fe.Reason = ErrorReasonStorageSecretNotFound
	fe.Code = http.StatusFailedDependency
	fe.SetRawError(e)
	return fe
}

func (fe *FormatError) SetErrorMissParameter(name string) *FormatError {
	fe.Message = fmt.Sprintf("parameter %s is empty", name)
	fe.Reason = ErrorReasonMissParameter
	fe.Code = http.StatusBadRequest
	return fe
}

func (fe *FormatError) SetErrorClusterNotFound(cluster string) *FormatError {
	fe.Message = fmt.Sprintf("cluster %s not found", cluster)
	fe.Reason = ErrorReasonClusterNotFound
	fe.Code = http.StatusNotFound
	return fe
}

func (fe *FormatError) SetErrorClassStatusNotActive(status string) *FormatError {
	fe.Message = fmt.Sprintf("class status %s not active", status)
	fe.Reason = ErrorReasonClassStatusNotActive
	fe.Code = http.StatusBadRequest
	return fe
}

func (fe *FormatError) SetErrorMapParameterMissing(name string) *FormatError {
	fe.Message = fmt.Sprintf("missing parameter %v", name)
	fe.Reason = ErrorReasonMapParameterNotMatch
	fe.Code = http.StatusBadRequest
	return fe
}
func (fe *FormatError) SetErrorMapParameterNumNotMatch(input, max int) *FormatError {
	fe.Message = fmt.Sprintf("get Parameters number %d not match the num %d in type",
		input, max)
	fe.Reason = ErrorReasonMapParameterNotMatch
	fe.Code = http.StatusBadRequest
	return fe
}

func (fe *FormatError) SetErrorClassNameTooLong(className string) *FormatError {
	fe.Message = fmt.Sprintf("storage class name '%v' too long", className)
	fe.Reason = ErrorReasonClassNameTooLong
	fe.Code = http.StatusBadRequest
	return fe
}

func (fe *FormatError) SetErrorSystemObject(objName string) *FormatError {
	fe.Message = fmt.Sprintf("'%v' is system object", objName)
	fe.Reason = ErrorReasonSystemObject
	fe.Code = http.StatusForbidden
	return fe
}

func (fe *FormatError) SetErrorObjInSystemNamespace(objName, nsName string) *FormatError {
	fe.Message = fmt.Sprintf("'%v' is system namespace '%v'", objName, nsName)
	fe.Reason = ErrorReasonObjInSystemNamespace
	fe.Code = http.StatusForbidden
	return fe
}

// other error
func (fe *FormatError) SetErrorAuthFailed(e error) *FormatError {
	fe.Message = fmt.Sprintf("parse auth info failed")
	fe.Reason = ErrorReasonAuthFailed
	fe.Code = http.StatusForbidden
	fe.SetRawError(e)
	return fe
}

func (fe *FormatError) SetErrorInternalServerError(e error) *FormatError {
	fe.Message = fmt.Sprintf("internal server error")
	fe.Reason = ErrorReasonInternalServerError
	fe.Code = http.StatusInternalServerError
	fe.SetRawError(e)
	return fe
}

func (fe *FormatError) SetErrorBadClusterConfig(cluster string, e error) *FormatError {
	fe.Message = fmt.Sprintf("cluster %s got a bad config", cluster)
	fe.Reason = ErrorReasonClusterBadConfig
	fe.Code = http.StatusBadRequest
	fe.SetRawError(e)
	return fe
}

// admission
func (fe *FormatError) SetErrorQuotaNotComplete(storageClass, partition string) *FormatError {
	fe.Message = fmt.Sprintf("quota for storageClass %s in partition %s not complete", storageClass, partition)
	fe.Reason = ErrorReasonQuotaNotComplete
	fe.Code = http.StatusBadRequest
	return fe
}
func (fe *FormatError) SetErrorQuotaNotCompleteByError(e error) *FormatError {
	fe.Message = fmt.Sprintf("check quota failed, %v", e)
	fe.Reason = ErrorReasonQuotaNotComplete
	fe.Code = http.StatusBadRequest
	return fe
}
func (fe *FormatError) SetErrorQuotaNotCompleteFromApi(ae *ApiError) *FormatError {
	fe.Message = ae.Message
	fe.Data = ae.Data
	fe.Reason = ErrorReasonQuotaNotComplete
	fe.Code = http.StatusBadRequest
	return fe
}

func (fe *FormatError) SetErrorQuotaExceeded(e error) *FormatError {
	fe.Message = e.Error()
	fe.Reason = ErrorReasonQuotaExceeded
	fe.Code = http.StatusBadRequest
	fe.SetRawError(e)
	return fe
}
