package storageclass

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"

	"github.com/caicloud/storage-admin/pkg/admin/content"
	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

var (
	storageclassPath       = fmt.Sprintf("/clusters/{%s}/storageclasses", constants.ParameterCluster)
	singleStorageclassPath = fmt.Sprintf("/clusters/{%s}/storageclasses/{%s}", constants.ParameterCluster, constants.ParameterStorageClass)

	SubPathListClass   = storageclassPath
	SubPathCreateClass = storageclassPath
	SubPathGetClass    = singleStorageclassPath
	SubPathUpdateClass = singleStorageclassPath
	SubPathDeleteClass = singleStorageclassPath
)

func AddEndpoints(ws *restful.WebService, c content.Interface) {
	ws.Route(ws.GET(SubPathListClass).To(HandleStorageClassList(c)).Doc("list storage class").
		Param(util.PathParamCluster(ws)).
		Param(util.QueryParamStart(ws)).
		Param(util.QueryParamLimit(ws)).
		Param(util.QueryParamType(ws)).
		Param(util.QueryParamName(ws)).
		Returns(http.StatusOK, "ok", apiv1a1.ListStorageClassResponse{}))
	ws.Route(ws.POST(SubPathCreateClass).To(HandleStorageClassCreate(c)).Doc("create storage class").
		Param(util.PathParamCluster(ws)).
		Reads(apiv1a1.CreateStorageClassRequest{}).
		Returns(http.StatusCreated, "ok", apiv1a1.CreateStorageClassResponse{}))

	ws.Route(ws.GET(SubPathGetClass).To(HandleStorageClassGet(c)).Doc("get storage class").
		Param(util.PathParamCluster(ws)).
		Param(util.PathParamStorageClass(ws)).
		Returns(http.StatusOK, "ok", apiv1a1.ReadStorageClassResponse{}))
	ws.Route(ws.PUT(SubPathUpdateClass).To(HandleStorageClassUpdate(c)).Doc("update storage class").
		Param(util.PathParamCluster(ws)).
		Param(util.PathParamStorageClass(ws)).
		Reads(apiv1a1.UpdateStorageClassRequest{}).
		Returns(http.StatusOK, "ok", apiv1a1.UpdateStorageClassResponse{}))
	ws.Route(ws.DELETE(SubPathDeleteClass).To(HandleStorageClassDelete(c)).Doc("delete storage class").
		Param(util.PathParamCluster(ws)).
		Param(util.PathParamStorageClass(ws)).
		Returns(http.StatusNoContent, "ok", nil))
}

func handleStorageClassListPreWork(request *restful.Request, c content.Interface) (
	cluster string, kc kubernetes.Interface, typeName, name string, start, limit int, fe *errors.FormatError) {
	typeName, name = util.GetRequestTypeAndName(request)
	start, limit, fe = util.HandleSimpleListPreWork(request)
	if fe != nil {
		return
	}
	// cluster
	cluster, fe = util.GetCluster(request)
	if fe != nil {
		return
	}
	// client
	kc, fe = c.GetSubClient(cluster)
	if fe != nil {
		return
	}
	return
}

func handleStorageClassCreatePreWork(request *restful.Request, c content.Interface) (
	cluster string, kc kubernetes.Interface, req *apiv1a1.CreateStorageClassRequest, fe *errors.FormatError) {
	// cluster
	cluster, fe = util.GetCluster(request)
	if fe != nil {
		return
	}
	// client
	kc, fe = c.GetSubClient(cluster)
	if fe != nil {
		return
	}
	// parse body
	req = new(apiv1a1.CreateStorageClassRequest)
	e := util.ReadBodyJson(req, request)
	if e != nil {
		fe = errors.NewError().SetErrorBadRequestBody(e)
		return
	}
	// check req
	if len(req.Name) == 0 {
		e = fmt.Errorf("empty class name")
	} else if len(req.Name) > constants.ClassNameMaxLength {
		fe = errors.NewError().SetErrorClassNameTooLong(req.Name)
		return
	} else if len(req.Service) == 0 {
		e = fmt.Errorf("empty class service")
	}
	if e != nil {
		fe = errors.NewError().SetErrorBadRequestBody(e)
		return
	}
	return
}

func handleStorageClassGetPreWork(request *restful.Request, c content.Interface) (
	cluster string, kc kubernetes.Interface, name string, fe *errors.FormatError) {
	// cluster
	cluster, fe = util.GetCluster(request)
	if fe != nil {
		return
	}
	// client
	kc, fe = c.GetSubClient(cluster)
	if fe != nil {
		return
	}
	// class name
	name, fe = util.GetStorageClass(request)
	return
}

func handleStorageClassUpdatePreWork(request *restful.Request, c content.Interface) (
	cluster string, kc kubernetes.Interface, name string,
	req *apiv1a1.UpdateStorageClassRequest, fe *errors.FormatError) {
	cluster, kc, name, fe = handleStorageClassGetPreWork(request, c)
	if fe != nil {
		return
	}
	// parse body
	req = new(apiv1a1.UpdateStorageClassRequest)
	e := util.ReadBodyJson(req, request)
	if e != nil {
		fe = errors.NewError().SetErrorBadRequestBody(e)
	}
	return
}

func handleStorageClassDeletePreWork(request *restful.Request, c content.Interface) (
	cluster string, kc kubernetes.Interface, name string, fe *errors.FormatError) {
	return handleStorageClassGetPreWork(request, c)
}
