package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	restful "github.com/emicklei/go-restful"

	. "github.com/caicloud/storage-admin/pkg/constants"
	errors "github.com/caicloud/storage-admin/pkg/errors"
)

func QueryParamStart(ws *restful.WebService) *restful.Parameter {
	return ws.QueryParameter(ParameterStart, "page split start").DataType("number").Required(false)
}
func QueryParamLimit(ws *restful.WebService) *restful.Parameter {
	return ws.QueryParameter(ParameterLimit, "page split limit").DataType("number").Required(false)
}
func QueryParamType(ws *restful.WebService) *restful.Parameter {
	return ws.QueryParameter(ParameterStorageType, "storage type name").DataType("string").Required(false)
}
func QueryParamName(ws *restful.WebService) *restful.Parameter {
	return ws.QueryParameter(ParameterName, "object name").DataType("string").Required(false)
}
func PathParamStorageService(ws *restful.WebService) *restful.Parameter {
	return ws.PathParameter(ParameterStorageService, "storage service name").DataType("string").Required(true)
}
func PathParamStorageClass(ws *restful.WebService) *restful.Parameter {
	return ws.PathParameter(ParameterStorageClass, "storage solution name").DataType("string").Required(true)
}
func PathParamCluster(ws *restful.WebService) *restful.Parameter {
	return ws.PathParameter(ParameterCluster, "cluster id").DataType("string").Required(true)
}
func PathParamPartition(ws *restful.WebService) *restful.Parameter {
	return ws.PathParameter(ParameterPartition, "partition name").DataType("string").Required(true)
}
func PathParamVolume(ws *restful.WebService) *restful.Parameter {
	return ws.PathParameter(ParameterVolume, "volume name").DataType("string").Required(true)
}

func GetRequestTypeAndName(request *restful.Request) (typeName, name string) {
	typeName = request.QueryParameter(ParameterStorageType)
	name = request.QueryParameter(ParameterName)
	return
}
func HandleSimpleListPreWork(request *restful.Request) (start, limit int, fe *errors.FormatError) {
	return GetRequestPageStartAndLimit(request)
}

func ReadBodyJson(req interface{}, request *restful.Request) error {
	// body read
	defer request.Request.Body.Close()
	b, e := ioutil.ReadAll(request.Request.Body)
	if e != nil {
		return e
	}
	// body json
	e = json.Unmarshal(b, req)
	if e != nil {
		return e
	}
	return nil
}

func GetRequestPageStartAndLimit(request *restful.Request) (start, limit int, fe *errors.FormatError) {
	startStr := request.QueryParameter(ParameterStart)
	limitStr := request.QueryParameter(ParameterLimit)
	return getRequestPageStartAndLimit(startStr, limitStr)
}

func getRequestPageStartAndLimit(startStr, limitStr string) (start, limit int, fe *errors.FormatError) {
	var e error
	switch {
	case len(startStr) == 0 && len(limitStr) == 0:
		return 0, 0, nil
	case len(startStr) == 0 && len(limitStr) > 0:
		if limit, e = strconv.Atoi(limitStr); e != nil || limit < 1 {
			return 0, 0, errors.NewError().SetErrorBadPageStartOrLimit(startStr, limitStr)
		}
		return 0, limit, nil
	case len(startStr) > 0 && len(limitStr) == 0:
		if start, e = strconv.Atoi(startStr); e != nil || start < 0 {
			return 0, 0, errors.NewError().SetErrorBadPageStartOrLimit(startStr, limitStr)
		}
		return start, 0, nil
	case len(startStr) > 0 && len(limitStr) > 0:
		if start, e = strconv.Atoi(startStr); e == nil {
			limit, e = strconv.Atoi(limitStr)
		}
		if e != nil || start < 0 || limit < 1 {
			return 0, 0, errors.NewError().SetErrorBadPageStartOrLimit(startStr, limitStr)
		}
		return start, limit, nil
	}
	return start, limit, nil
}

func GetStartLimitEnd(start, limit, arrayLen int) (end int) {
	if limit == 0 { // no limit
		end = arrayLen
	} else {
		end = start + limit
		if end > arrayLen {
			end = arrayLen
		}
	}
	return end
}

func GetStorageService(request *restful.Request) (string, *errors.FormatError) {
	name := request.PathParameter(ParameterStorageService)
	if len(name) == 0 {
		return "", errors.NewError().SetErrorMissParameter(ParameterStorageService)
	}
	return name, nil
}

func GetStorageClass(request *restful.Request) (string, *errors.FormatError) {
	name := request.PathParameter(ParameterStorageClass)
	if len(name) == 0 {
		return "", errors.NewError().SetErrorMissParameter(ParameterStorageClass)
	}
	return name, nil
}

func GetCluster(request *restful.Request) (string, *errors.FormatError) {
	name := request.PathParameter(ParameterCluster)
	if len(name) == 0 {
		return "", errors.NewError().SetErrorMissParameter(ParameterCluster)
	}
	return name, nil
}

func GetPartition(request *restful.Request) (string, *errors.FormatError) {
	name := request.PathParameter(ParameterPartition)
	if len(name) == 0 {
		return "", errors.NewError().SetErrorMissParameter(ParameterPartition)
	}
	return name, nil
}

func GetVolume(request *restful.Request) (string, *errors.FormatError) {
	name := request.PathParameter(ParameterVolume)
	if len(name) == 0 {
		return "", errors.NewError().SetErrorMissParameter(ParameterVolume)
	}
	return name, nil
}

// map about

func CheckStorageServiceParameters(input, commons map[string]string) *errors.FormatError {
	return checkOptionalMapParameters(input, commons)
}

func CheckStorageClassParameters(input, optional map[string]string) *errors.FormatError {
	return checkOptionalMapParameters(input, optional)
}

func checkOptionalMapParameters(input, optional map[string]string) *errors.FormatError {
	// all input keys should exist in optional
	if len(input) > len(optional) {
		return errors.NewError().SetErrorMapParameterNumNotMatch(len(input), len(optional))
	}
	for key := range input {
		_, ok := optional[key]
		if !ok {
			return errors.NewError().SetErrorMapParameterMissing(key)
		}
	}
	return nil
}

func SyncStorageClassWithTypeAndService(scParameters, ssParameters,
	tpRequired, tpOptional map[string]string) (newClassParameters map[string]string) {
	needUpdate := false
	// delete extra parameters
	for k := range scParameters {
		_, ok := tpRequired[k]
		if !ok {
			_, ok = tpOptional[k]
		}
		if !ok {
			needUpdate = true
			break
		}
	}
	// sync with service
	for k, v := range ssParameters {
		if scParameters[k] != v {
			needUpdate = true
			break
		}
	}
	if !needUpdate {
		return nil
	}

	// make a new map to reset parameters
	newClassParameters = make(map[string]string, len(scParameters))
	for k := range tpOptional {
		if v, ok := scParameters[k]; ok {
			newClassParameters[k] = v
		}
	}
	for k, v := range ssParameters {
		newClassParameters[k] = v
	}
	return newClassParameters
}

// namespaceQuotaName get namespace quota name, quota name should under some rules about ns name
func NamespaceQuotaName(nsName string) string {
	return nsName
}

// num quota label by storage class
func LabelKeyStorageQuotaNum(className string) string {
	return fmt.Sprintf(FormatLabelKeyStorageQuotaNum, className)
}

// size quota label by storage class
func LabelKeyStorageQuotaSize(className string) string {
	return fmt.Sprintf(FormatLabelKeyStorageQuotaSize, className)
}
