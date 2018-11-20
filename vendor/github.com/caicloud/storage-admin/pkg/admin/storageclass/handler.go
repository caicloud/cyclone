package storageclass

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/golang/glog"

	"github.com/caicloud/storage-admin/pkg/admin/content"
	"github.com/caicloud/storage-admin/pkg/admin/helper"
	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

func HandleStorageClassList(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var (
			fErr errors.FormatError
			resp apiv1a1.ListStorageClassResponse
		)
		// pre work
		cluster, kc, typeName, name, start, limit, fe := handleStorageClassListPreWork(request, c)
		if fe != nil {
			glog.Errorf("handleStorageClassListPreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageServiceList[cluster=%s][type=%s][name=%s][start=%d][limit=%d]",
			cluster, typeName, name, start, limit)
		glog.Infof("%s start", logPath)

		// list
		scs, e := helper.ListStorageClass(kc, typeName, name)
		if e != nil {
			glog.Errorf("%s ListStorageClass failed, %v", logPath, e)
			fErr.SetErrorInternalServerError(e)
			response.WriteHeaderAndEntity(fErr.Code, fErr.Api())
			return
		}

		// pack response
		items := util.DefaultTranslatorMetaToApi().StorageClassListActiveOnly(scs)
		resp.MetaData.Total = len(items)
		if start >= resp.MetaData.Total {
			glog.Warningf("%s requested start=%d out of total=%d", logPath, start, resp.MetaData.Total)
			response.WriteHeaderAndEntity(http.StatusOK, resp)
			return
		}
		end := util.GetStartLimitEnd(start, limit, resp.MetaData.Total)
		resp.Items = items[start:end]

		// return
		glog.Infof("%s done [total=%d]", logPath, resp.MetaData.Total)
		response.WriteHeaderAndEntity(http.StatusOK, resp)
		return
	}
}

func HandleStorageClassCreate(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var resp apiv1a1.CreateStorageClassResponse
		mc := c.GetClient()
		// pre work
		cluster, kc, req, fe := handleStorageClassCreatePreWork(request, c)
		if fe != nil {
			glog.Errorf("handleStorageClassCreatePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageClassCreate[cluster=%s][name=%s][service=%s][alias=%s][desc=%d]",
			cluster, req.Name, req.Service, req.Alias, len(req.Description))
		glog.Infof("%s start", logPath)

		// get service, type and check
		service, tp, fe := helper.GetStorageServiceAndTypeWithOptParamMapCheck(req.Service, req.Parameters, mc)
		if fe != nil {
			glog.Errorf("%s GetStorageServiceAndTypeWithCheckStatusAndOptParamMap failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// parse
		class := util.DefaultTranslatorApiToMeta().StorageClassNewFromCreate(req, service, tp)
		// check and prework
		switch class.Provisioner {
		case kubernetes.StorageClassProvisionerGlusterfs:
			class, fe = helper.CreateGlusterFsPrework(class, service, mc, kc)
		}
		if fe != nil {
			glog.Errorf("%s prework for create %s failed, %v", logPath, class.Provisioner, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		// create
		re, fe := helper.CreateStorageClass(class, kc)
		if fe != nil {
			glog.Errorf("%s CreateStorageClass failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		util.DefaultTranslatorMetaToApi().StorageClassSet(&resp, re)
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusCreated, resp)
		return
	}
}

func HandleStorageClassGet(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var resp apiv1a1.ReadStorageClassResponse
		// pre work
		cluster, kc, name, fe := handleStorageClassGetPreWork(request, c)
		if fe != nil {
			glog.Errorf("handleStorageClassGetPreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageClassGet[cluster=%s][name=%s]", cluster, name)
		glog.Infof("%s start", logPath)

		// get
		re, fe := helper.GetStorageClass(name, kc)
		if fe != nil {
			glog.Errorf("%s GetStorageClass failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		util.DefaultTranslatorMetaToApi().StorageClassSet(&resp, re)
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusOK, resp)
		return
	}
}

func HandleStorageClassUpdate(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var resp apiv1a1.ReadStorageClassResponse
		// pre work
		cluster, kc, name, req, fe := handleStorageClassUpdatePreWork(request, c)
		if fe != nil {
			glog.Errorf("handleStorageClassUpdatePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageClassUpdate[cluster=%s][name=%s][alias=%s][description=%s]",
			cluster, name, req.Alias, req.Description)
		glog.Infof("%s start", logPath)

		// get pre version
		class, fe := helper.GetStorageClassWithStatusCheck(name, kc)
		if fe != nil {
			glog.Errorf("%s GetStorageClass failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// pack and update
		util.DefaultTranslatorApiToMeta().StorageClassSetFromUpdate(class, req)
		re, fe := helper.UpdateStorageClass(class, kc)
		if fe != nil {
			glog.Errorf("%s UpdateStorageClass failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		util.DefaultTranslatorMetaToApi().StorageClassSet(&resp, re)
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusOK, resp)
		return
	}
}

func HandleStorageClassDelete(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		// pre work
		cluster, kc, name, fe := handleStorageClassDeletePreWork(request, c)
		if fe != nil {
			glog.Errorf("handleStorageClassDeletePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageClassDelete[cluster=%s][name=%s]", cluster, name)
		glog.Infof("%s start", logPath)

		// check
		if name == constants.SystemStorageClass {
			fe = errors.NewError().SetErrorSystemObject(name)
			glog.Errorf("%s check failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		// terminate
		if fe = helper.TerminateStorageClass(name, kc); fe != nil {
			glog.Errorf("%s TerminateStorageClass failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusNoContent, nil)
		return
	}
}
