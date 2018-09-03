/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package definition

import (
	"context"
	"net/http"
	"reflect"

	"github.com/caicloud/devops-admin/pkg/api/models"
	"github.com/caicloud/devops-admin/pkg/errors"
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"
)

// Verb used by HandlerDecoration
type Verb string

const (
	// VerbList lists an array
	VerbList Verb = "list"
	// VerbGet gets an object
	VerbGet Verb = "get"
	// VerbCreate creates an object
	VerbCreate Verb = "create"
	// VerbUpdate updates an object
	VerbUpdate Verb = "update"
	// VerbDelete deletes an object
	VerbDelete Verb = "delete"
)

// Key is the type of context.Context
type Key string

const (
	// KeyRequest is the key of request
	KeyRequest Key = "Context.Request"
)

// HandlerDecoration defines a decoration of handler
// A handler is a function. The declaration of handler
// should be compatible with the definition of specified Verb.
// VerbDelete definition (return 1 value):
// func(ctx context.Context) error -> response with 204 or error
// e.g.
// func DeleteApplication(ctx context.Context) error
//
// VerbGet, VerbCreate, VerbUpdate definition (return 2 values):
// The first return value (type interface{}) can be any type which you like.
// func(ctx context.Context) (interface{},error) -> response with 200/201 or error
// e.g.
// func GetApplication(ctx context.Context) (*Application,error)
//
// VerbList definition (return 3 values):
// The first return value is the total number of requested resources.
// The second return value is an array of any type.
// func(ctx context.Context) (int,interface{},error) -> response with 200 or error
// e.g.
// func ListApplication(ctx context.Context) (int,[]Application,error)
type HandlerDecoration struct {
	Verb    Verb
	Handler interface{}
	Value   reflect.Value
}

// A mapping of verb and number of return values
var verbMapping = map[Verb]int{
	VerbDelete: 1,
	VerbGet:    2,
	VerbCreate: 2,
	VerbUpdate: 2,
	VerbList:   3,
}

// NewHandlerDecoration creates a HandlerDecoration. handler should meet the requirement
func NewHandlerDecoration(verb Verb, handler interface{}) *HandlerDecoration {
	// if hadnler does not meet the requirements, panic()
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if handlerType.NumIn() != 1 {
		glog.Fatalf("handler must have only 1 parameter")
	}
	if handlerType.In(0) != contextType {
		glog.Fatalf("the parameter of handler must be context.Context")
	}
	returnValueCount := handlerType.NumOut()
	// check return values count
	if count, ok := verbMapping[verb]; ok {
		if count != returnValueCount {
			glog.Fatalf("%s handler must have %d return value", verb, count)
		}
	} else {
		glog.Fatalf("unknown verb: %s", verb)
	}
	// check the last return value
	if !handlerType.Out(returnValueCount - 1).AssignableTo(errorType) {
		glog.Fatalf("the last return value of %s handler must be compatible with error: %s",
			verb, handlerType.String())
	}
	return &HandlerDecoration{
		verb,
		handler,
		handlerValue,
	}
}

// Handle handles a request
func (hd *HandlerDecoration) Handle(request *restful.Request, resp *restful.Response) {
	ctx := context.WithValue(context.Background(), KeyRequest, request)
	result := hd.Value.Call([]reflect.Value{reflect.ValueOf(ctx)})
	errValue := result[verbMapping[hd.Verb]-1]
	if errValue.IsNil() {
		switch hd.Verb {
		case VerbDelete:
			resp.WriteHeader(http.StatusNoContent)
			return
		case VerbCreate, VerbGet, VerbUpdate:
			statusCode := http.StatusOK
			if hd.Verb == VerbCreate {
				statusCode = http.StatusCreated
			}
			// check obj type
			obj := result[0]
			// if obj is []byte, writes by resp.Write()
			// otherwise resp.WriteHeaderAndEntity()
			objType := obj.Type()
			if (objType.Kind() == reflect.Array || objType.Kind() == reflect.Slice) &&
				objType.Elem().AssignableTo(reflect.TypeOf(byte(0))) {
				resp.WriteHeader(statusCode)
				data := obj.Interface().([]byte)
				resp.Write(data)
			} else {
				resp.WriteHeaderAndEntity(statusCode, obj.Interface())
			}
			return
		case VerbList:
			total := int(result[0].Int())
			resp.WriteHeaderAndEntity(http.StatusOK, models.NewListResponse(total, result[1].Interface()))
			return
		default:
			// should not come here
			glog.Fatalf("app enters unknown area. handler should not have %d results", len(result))
		}
	}
	// handle error
	glog.Errorf("error information %#v", errValue.Interface())
	switch err := errValue.Interface().(type) {
	case *errors.Error:
		resp.WriteHeaderAndEntity(err.Code, err)
	case error:
		glog.Infof("%s handler returns an error but the type is not custom error type", hd.Verb)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Error{
			Message: err.Error(),
			Reason:  errors.ReasonInternal,
			Code:    http.StatusInternalServerError,
		})
	default:
		// should not come here
		glog.Fatalf("%s handler returns an unknown error type, check the function: %s",
			hd.Verb, hd.Value.Type().String())
	}
}
