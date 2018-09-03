/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"

	log "github.com/golang/glog"

	"github.com/caicloud/devops-admin/pkg/api/v1"
	. "github.com/caicloud/devops-admin/pkg/errors"
)

const (
	// limitQueryName represents the name of the query parameter for pagination limit, if not set, return all.
	limitQueryName string = "limit"

	// startQueryName represents the name of the query parameter for pagination start.
	startQueryName string = "start"

	// defaultStart represents the default value of pagination start.
	defaultStart = "0"

	// HeaderTenant represents the the key of tenant in request header.
	HeaderTenant = "X-Tenant"
)

// GetPagination gets pagination from query parameters of request.
func GetPagination(request *http.Request) (int, int, error) {
	start := request.FormValue(startQueryName)
	if len(start) <= 0 {
		start = defaultStart
	}

	s, err := strconv.Atoi(start)
	if err != nil {
		return 0, 0, ErrorParamTypeError.Format(startQueryName, "number", "string")
	}

	l := 0
	limit := request.FormValue(limitQueryName)
	if len(limit) > 0 {
		l, err = strconv.Atoi(limit)
		if err != nil {
			return 0, 0, ErrorParamTypeError.Format(limitQueryName, "number", "string")
		}
	}

	return s, l, nil
}

// GetHeaderParameter gets parameter from request path.
func GetHeaderParameter(request *http.Request, name string) (string, error) {
	tenant := request.Header.Get(name)
	if len(tenant) <= 0 {
		return "", ErrorHeaderParamNotFound.Format(name)
	}

	return tenant, nil
}

// GetTenant get the tenant from request header.
func GetTenant(req *http.Request) (string, error) {
	return GetHeaderParameter(req, HeaderTenant)
}

// GetJsonPayload reads json payload from request and unmarshal it into entity.
func GetJsonPayload(request *http.Request, entity interface{}) error {
	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return ErrorUnknownInternal.Format(err)
	}

	err = json.Unmarshal(content, entity)
	if err != nil {
		log.Errorf("Failed to unmarshal request body: %v\n %s", err, string(content))
		return ErrorUnknownInternal.Format(err)
	}

	return nil
}

// GetJsonPayloadAndKeepState reads json payload from request and unmarshal it into entity, and keep the request body.
func GetJsonPayloadAndKeepState(request *http.Request, entity interface{}) error {
	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return ErrorUnknownInternal.Format(err)
	}

	// Restore the io.ReadCloser to its original state.
	request.Body = ioutil.NopCloser(bytes.NewBuffer(content))

	err = json.Unmarshal(content, entity)
	if err != nil {
		log.Errorf("Failed to unmarshal request body: %v\n %s", err, string(content))
		return ErrorUnknownInternal.Format(err)
	}

	return nil
}

// SetJsonPayload marshals entity into json payload, and writes it into request body.
func SetJsonPayload(request *http.Request, entity interface{}) error {
	content, err := json.Marshal(entity)
	if err != nil {
		log.Errorf("Failed to unmarshal request body: %v\n %s", err, string(content))
		return ErrorUnknownInternal.Format(err)
	}

	request.Body = ioutil.NopCloser(bytes.NewBuffer(content))
	contentLength := len(content)
	request.Header.Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(contentLength))
	request.ContentLength = int64(contentLength)

	return nil
}

// ResponseWithError responses the request with error.
func ResponseWithError(response http.ResponseWriter, err error) {
	recorder := httptest.NewRecorder()
	switch err := err.(type) {
	case *Error:
		respJSON, tErr := json.Marshal(v1.ErrorResponse{
			Message: err.Error(),
			Reason:  err.Reason,
		})
		if tErr != nil {
			log.Errorf("encode response error: %s", err)
			break
		}

		if _, tErr = recorder.Body.Write(respJSON); tErr != nil {
			log.Errorf("write response error: %s", err)
			break
		}

		recorder.Code = err.Code
	case error:
		respJSON, tErr := json.Marshal(v1.ErrorResponse{
			Message: err.Error(),
			Reason:  ReasonInternal,
		})
		if tErr != nil {
			log.Errorf("encode response error: %s", err)
			break
		}

		if _, tErr = recorder.Body.Write(respJSON); tErr != nil {
			log.Errorf("write response error: %s", err)
			break
		}

		recorder.Code = http.StatusInternalServerError
	default:
		// should not come here
		log.Errorf("%s is an unknown error type", err)
	}

	copyResp(recorder, response)
}

// ResponseWithList responses list with metadata.
func ResponseWithList(response http.ResponseWriter, list interface{}, total int) {
	recorder := httptest.NewRecorder()
	respJSON, err := json.Marshal(v1.ListResponse{
		Meta: v1.ListMeta{
			Total: total,
		},
		Items: list,
	})
	if err != nil {
		log.Errorf("encode response error: %s", err)
		copyResp(recorder, response)
		return
	}

	if _, err = recorder.Body.Write(respJSON); err != nil {
		log.Errorf("write response error: %s", err)
	}

	copyResp(recorder, response)
}

// ResponseWithHeaderAndEntity responses list with metadata.
func ResponseWithHeaderAndEntity(response http.ResponseWriter, status int, value interface{}) {
	recorder := httptest.NewRecorder()
	if value != nil {
		respJSON, err := json.Marshal(value)
		if err != nil {
			log.Errorf("encode response error: %s", err)
			copyResp(recorder, response)
			return
		}

		if _, err = recorder.Body.Write(respJSON); err != nil {
			log.Errorf("write response error: %s", err)
		}
	}

	recorder.Code = status

	copyResp(recorder, response)
}

func copyResp(rec *httptest.ResponseRecorder, rw http.ResponseWriter) {
	for k, v := range rec.Header() {
		rw.Header()[k] = v
	}

	// Must set the content type of header before write header.
	rw.Header().Set(http.CanonicalHeaderKey("Content-Type"), "application/json")
	rw.WriteHeader(rec.Result().StatusCode)
	rw.Write(rec.Body.Bytes())
}
