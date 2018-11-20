package datavolume

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/golang/glog"

	"github.com/caicloud/storage-admin/pkg/admin/content"
	"github.com/caicloud/storage-admin/pkg/admin/helper"
	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/util"
)

func HandleDataVolumeList(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var (
			fErr errors.FormatError
			resp apiv1a1.ListDataVolumeResponse
		)
		// pre work
		cluster, kc, namespace, storageClass, name, start, limit, fe := handleDataVolumeListPreWork(request, c)
		if fe != nil {
			glog.Errorf("handleDataVolumeListPreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleDataVolumeList[cluster=%s][namespace=%s][class=%s][name=%s][start=%d][limit=%d]",
			cluster, namespace, storageClass, name, start, limit)
		glog.Infof("%s start", logPath)

		// list
		vols, e := helper.ListDataVolume(namespace, storageClass, name, kc)
		if e != nil {
			glog.Errorf("%s ListDataVolume failed, %v", logPath, e)
			fErr.SetErrorInternalServerError(e)
			response.WriteHeaderAndEntity(fErr.Code, fErr.Api())
			return
		}

		// pack response
		resp.MetaData.Total = len(vols)
		if start >= len(vols) {
			glog.Warningf("%s requested start=%d out of total=%d", logPath, start, len(vols))
			response.WriteHeaderAndEntity(http.StatusOK, resp)
			return
		}
		end := util.GetStartLimitEnd(start, limit, len(vols))
		resp.Items = util.DefaultTranslatorMetaToApi().DataVolumeList(vols[start:end])

		// return
		glog.Infof("%s done [total=%d]", logPath, len(vols))
		response.WriteHeaderAndEntity(http.StatusOK, resp)
		return
	}
}
func HandleDataVolumeCreate(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var resp *apiv1a1.CreateDataVolumeResponse
		// pre work
		cluster, kc, namespace, req, fe := handleDataVolumeCreatePreWork(request, c)
		if fe != nil {
			glog.Errorf("handleDataVolumeCreatePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleDataVolumeCreate[cluster=%s][namespace=%s][class=%s][name=%s]",
			cluster, namespace, req.StorageClass, req.Name)
		glog.Infof("%s start", logPath)

		// get class and check
		sc, fe := helper.GetStorageClassWithStatusCheck(req.StorageClass, kc)
		if fe != nil {
			glog.Errorf("%s GetStorageClassWithStatusCheck %s failed, %v", logPath, req.StorageClass, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		// translate and create
		volume := util.DefaultTranslatorApiToMeta().DataVolumeNewFromCreate(req, sc, namespace)
		resp, fe = helper.CreateDataVolume(volume, kc)
		if fe != nil {
			glog.Errorf("%s CreateDataVolume failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusCreated, resp)
		return
	}
}
func HandleDataVolumeGet(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var resp *apiv1a1.ReadDataVolumeResponse
		// pre work
		cluster, kc, namespace, name, fe := handleDataVolumeGetPreWork(request, c)
		if fe != nil {
			glog.Errorf("handleDataVolumeGetPreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleDataVolumeGet[cluster=%s][namespace=%s][name=%s]", cluster, namespace, name)
		glog.Infof("%s start", logPath)

		// get
		resp, fe = helper.GetDataVolume(namespace, name, kc)
		if fe != nil {
			glog.Errorf("%s GetDataVolume failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusOK, resp)
		return
	}
}
func HandleDataVolumeDelete(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		// pre work
		cluster, kc, namespace, name, fe := handleDataVolumeDeletePreWork(request, c)
		if fe != nil {
			glog.Errorf("handleDataVolumeDeletePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleDataVolumeGet[cluster=%s][namespace=%s][name=%s]",
			cluster, namespace, name)
		glog.Infof("%s start", logPath)

		// check
		// if _, ok := constants.SystemNamespaces[namespace]; ok {
		// 	fe = errors.NewError().SetErrorObjInSystemNamespace(name, namespace)
		// 	glog.Errorf("%s check namespace failed, %v", logPath, fe.Error())
		// 	response.WriteHeaderAndEntity(fe.Code, fe.Api())
		// 	return
		// }

		// get
		fe = helper.DeleteDataVolume(namespace, name, kc)
		if fe != nil {
			glog.Errorf("%s DeleteDataVolume failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusNoContent, nil)
		return
	}
}
