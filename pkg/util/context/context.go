package context

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"

	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// GetHTTPRequest gets request from context.
func GetHTTPRequest(ctx context.Context) *http.Request {
	return service.HTTPContextFrom(ctx).Request()
}

// GetHTTPResponseWriter gets response writer from context.
func GetHTTPResponseWriter(ctx context.Context) http.ResponseWriter {
	return service.HTTPContextFrom(ctx).ResponseWriter()
}

// GetHeaderParameter gets value from request.HeaderParameter.
func GetHeaderParameter(ctx context.Context, name string) (string, error) {
	request := GetHTTPRequest(ctx)

	value := request.Header.Get(name)
	if len(value) <= 0 {
		return "", cerr.ErrorHeaderParamNotFound.Error(name)
	}
	return value, nil
}

// GetJSONPayload reads json payload from request and unmarshal it into entity.
func GetJSONPayload(ctx context.Context, entity interface{}) error {
	request := GetHTTPRequest(ctx)

	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return cerr.ErrorUnknownInternal.Error(err)
	}

	err = json.Unmarshal(content, entity)
	if err != nil {
		log.Errorf("Failed to unmarshal request body: %v\n %s", err, string(content))
		return cerr.ErrorUnknownInternal.Error(err)
	}
	return nil
}

// GetQueryParameters gets values from request.QueryParameter.
func GetQueryParameters(ctx context.Context) (string, error) {
	request := GetHTTPRequest(ctx)
	return request.Form.Encode(), nil
}
