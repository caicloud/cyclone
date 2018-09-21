/*
Copyright 2018 caicloud authors. All rights reserved.
*/

package context

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/nirvana/service"
	log "github.com/golang/glog"

	. "github.com/caicloud/cyclone/pkg/util/http/errors"
)

// GetHttpRequest gets request from context.
func GetHttpRequest(ctx context.Context) *http.Request {
	return service.HTTPContextFrom(ctx).Request()
}

// GetHttpRequest gets request from context.
func GetHttpResponseWriter(ctx context.Context) http.ResponseWriter {
	return service.HTTPContextFrom(ctx).ResponseWriter()
}

// GetJsonPayload reads json payload from request and unmarshal it into entity.
func GetJsonPayload(ctx context.Context, entity interface{}) error {
	request := GetHttpRequest(ctx)

	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}

	err = json.Unmarshal(content, entity)
	if err != nil {
		log.Errorf("Failed to unmarshal request body: %v\n %s", err, string(content))
		return ErrorUnknownInternal.Error(err)
	}
	return nil
}
