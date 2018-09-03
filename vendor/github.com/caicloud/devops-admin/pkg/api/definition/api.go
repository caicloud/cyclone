/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package definition

import (
	"fmt"

	"github.com/emicklei/go-restful"
)

// GenerateRoutes generates routes from descriptors and registers them into ws
func GenerateRoutes(ws *restful.WebService, descriptors []Descriptor) *restful.WebService {
	// generate parameters from params and add them in builder
	var generateParameters = func(method func(string, string) *restful.Parameter,
		builder *restful.RouteBuilder,
		params []Param) {
		for _, param := range params {
			parameter := method(param.Name, param.Doc)
			parameter.DataType(param.Type)
			parameter.Required(param.Required)
			if param.Default != nil {
				parameter.DefaultValue(fmt.Sprint(param.Default))
			}
			builder.Param(parameter)
		}
	}

	for _, descriptor := range descriptors {
		for _, handler := range descriptor.Handlers {
			builder := ws.Method(handler.HTTPMethod).
				Path(descriptor.Path).
				To(handler.Handler).
				Doc(handler.Doc).
				Notes(handler.Note)

			generateParameters(ws.PathParameter, builder, handler.PathParams)
			generateParameters(ws.QueryParameter, builder, handler.QueryParams)
			generateParameters(ws.HeaderParameter, builder, handler.HeaderParams)

			for _, code := range handler.StatusCode {
				builder.Returns(code.Code, code.Message, code.Sample)
			}
			for _, filter := range handler.Filters {
				builder.Filter(filter)
			}
			ws.Route(builder)
		}
	}
	return ws
}
