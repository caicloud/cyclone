package datavolume

import (
	"fmt"

	restful "github.com/emicklei/go-restful"

	"github.com/caicloud/storage-admin/pkg/admin/content"
	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

var (
	volumePath       = fmt.Sprintf("/clusters/{%s}/partitions/{%s}/volumes", constants.ParameterCluster, constants.ParameterPartition)
	singleVolumePath = fmt.Sprintf("/clusters/{%s}/partitions/{%s}/volumes/{%s}", constants.ParameterCluster, constants.ParameterPartition, constants.ParameterVolume)

	SubPathListClass   = volumePath
	SubPathCreateClass = volumePath
	SubPathGetClass    = singleVolumePath
	SubPathDeleteClass = singleVolumePath
)

func AddEndpoints(ws *restful.WebService, c content.Interface) {
	ws.Route(ws.GET(SubPathListClass).To(HandleDataVolumeList(c)).Doc("list volume").
		Param(util.PathParamCluster(ws)).
		Param(util.PathParamPartition(ws)).
		Param(ws.QueryParameter(constants.ParameterStorageClass, "storage class name").DataType("string").Required(false)).
		Param(ws.QueryParameter(constants.ParameterName, "volume class name").DataType("string").Required(false)).
		Param(util.QueryParamStart(ws)).
		Param(util.QueryParamLimit(ws)).
		Returns(200, "ok", apiv1a1.ListDataVolumeResponse{}))
	ws.Route(ws.POST(SubPathCreateClass).To(HandleDataVolumeCreate(c)).Doc("create volume").
		Param(util.PathParamCluster(ws)).
		Param(util.PathParamPartition(ws)).
		Reads(apiv1a1.CreateDataVolumeRequest{}).
		Returns(201, "ok", apiv1a1.CreateDataVolumeResponse{}))
	ws.Route(ws.GET(SubPathGetClass).To(HandleDataVolumeGet(c)).Doc("get volume").
		Param(util.PathParamCluster(ws)).
		Param(util.PathParamPartition(ws)).
		Param(util.PathParamVolume(ws)).
		Returns(200, "ok", apiv1a1.ReadDataVolumeResponse{}))
	ws.Route(ws.DELETE(SubPathDeleteClass).To(HandleDataVolumeDelete(c)).Doc("delete volume").
		Param(util.PathParamCluster(ws)).
		Param(util.PathParamPartition(ws)).
		Param(util.PathParamVolume(ws)).
		Returns(204, "ok", nil))
}

func handleDataVolumeListPreWork(request *restful.Request, c content.Interface) (cluster string,
	kc kubernetes.Interface, namespace, storageClass, name string, start, limit int, fe *errors.FormatError) {
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
	// ns
	namespace, fe = util.GetPartition(request)
	if fe != nil {
		return
	}
	storageClass = request.QueryParameter(constants.ParameterStorageClass)
	name = request.QueryParameter(constants.ParameterName)
	return
}

func handleDataVolumeCreatePreWork(request *restful.Request, c content.Interface) (cluster string,
	kc kubernetes.Interface, namespace string, req *apiv1a1.CreateDataVolumeRequest, fe *errors.FormatError) {
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
	// ns
	namespace, fe = util.GetPartition(request)
	if fe != nil {
		return
	}
	// parse body
	req = new(apiv1a1.CreateDataVolumeRequest)
	e := util.ReadBodyJson(req, request)
	if e != nil {
		fe = errors.NewError().SetErrorBadRequestBody(e)
		return
	}
	// check req
	if len(req.Name) == 0 {
		e = fmt.Errorf("empty volume name")
	} else if len(req.StorageClass) == 0 {
		e = fmt.Errorf("empty storage class")
	} else if len(req.AccessModes) == 0 {
		e = fmt.Errorf("empty access modes")
	} else if req.Size < 1 {
		e = fmt.Errorf("bad volume size: %dG", req.Size)
	}
	if e != nil {
		fe = errors.NewError().SetErrorBadRequestBody(e)
		return
	}
	return
}

func handleDataVolumeGetPreWork(request *restful.Request, c content.Interface) (cluster string,
	kc kubernetes.Interface, namespace, name string, fe *errors.FormatError) {
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
	// ns
	namespace, fe = util.GetPartition(request)
	if fe != nil {
		return
	}
	name, fe = util.GetVolume(request)
	return
}

func handleDataVolumeDeletePreWork(request *restful.Request, c content.Interface) (cluster string,
	kc kubernetes.Interface, namespace, name string, fe *errors.FormatError) {
	return handleDataVolumeGetPreWork(request, c)
}
