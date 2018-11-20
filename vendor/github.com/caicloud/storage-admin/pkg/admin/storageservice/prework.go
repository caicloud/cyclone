package storageservice

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"

	"github.com/caicloud/storage-admin/pkg/admin/content"
	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/util"
)

var (
	servicePath       = fmt.Sprintf("/services")
	singleServicePath = fmt.Sprintf("/services/{%s}", constants.ParameterStorageService)

	SubPathListService   = servicePath
	SubPathCreateService = servicePath
	SubPathGetService    = singleServicePath
	SubPathUpdateService = singleServicePath
	SubPathDeleteService = singleServicePath
)

func AddEndpoints(ws *restful.WebService, c content.Interface) {
	ws.Route(ws.GET(SubPathListService).To(HandleStorageServiceList(c)).Doc("list storage service").
		Param(util.QueryParamStart(ws)).
		Param(util.QueryParamLimit(ws)).
		Param(util.QueryParamType(ws)).
		Param(util.QueryParamName(ws)).
		Returns(http.StatusOK, "ok", apiv1a1.ListStorageServiceResponse{}))
	ws.Route(ws.POST(SubPathCreateService).To(HandleStorageServiceCreate(c)).Doc("create storage service").
		Reads(apiv1a1.CreateStorageServiceRequest{}).
		Returns(http.StatusCreated, "ok", apiv1a1.CreateStorageServiceResponse{}))
	ws.Route(ws.GET(SubPathGetService).To(HandleStorageServiceGet(c)).Doc("get storage service").
		Param(util.PathParamStorageService(ws)).
		Returns(http.StatusOK, "ok", apiv1a1.ReadStorageServiceResponse{}))
	ws.Route(ws.PUT(SubPathUpdateService).To(HandleStorageServiceUpdate(c)).Doc("update storage service").
		Param(util.PathParamStorageService(ws)).
		Reads(apiv1a1.UpdateStorageServiceRequest{}).
		Returns(http.StatusOK, "ok", apiv1a1.UpdateStorageServiceResponse{}))
	ws.Route(ws.DELETE(SubPathDeleteService).To(HandleStorageServiceDelete(c)).Doc("delete storage service").
		Param(util.PathParamStorageService(ws)).
		Returns(http.StatusNoContent, "ok", nil))
}

func handleStorageServiceListPreWork(request *restful.Request) (typeName, name string, start, limit int, fe *errors.FormatError) {
	typeName, name = util.GetRequestTypeAndName(request)
	start, limit, fe = util.HandleSimpleListPreWork(request)
	return
}

func handleStorageServiceCreatePreWork(request *restful.Request) (req *apiv1a1.CreateStorageServiceRequest, fe *errors.FormatError) {
	// parse body
	req = new(apiv1a1.CreateStorageServiceRequest)
	e := util.ReadBodyJson(req, request)
	if e != nil {
		fe = errors.NewError().SetErrorBadRequestBody(e)
		return
	}
	// check req
	if len(req.Name) == 0 {
		e = fmt.Errorf("empty service name")
	} else if len(req.Type) == 0 {
		e = fmt.Errorf("empty service type")
	}
	if e != nil {
		fe = errors.NewError().SetErrorBadRequestBody(e)
		return
	}
	return
}

func handleStorageServiceGetPreWork(request *restful.Request) (name string, fe *errors.FormatError) {
	return util.GetStorageService(request)
}

func handleStorageServiceUpdatePreWork(request *restful.Request) (name string, req *apiv1a1.UpdateStorageServiceRequest,
	fe *errors.FormatError) {
	name, fe = handleStorageServiceGetPreWork(request)
	if fe != nil {
		return
	}
	// parse body
	req = new(apiv1a1.UpdateStorageServiceRequest)
	e := util.ReadBodyJson(req, request)
	if e != nil {
		fe = errors.NewError().SetErrorBadRequestBody(e)
	}
	return
}

func handleStorageServiceDeletePreWork(request *restful.Request) (name string, fe *errors.FormatError) {
	return handleStorageServiceGetPreWork(request)
}
