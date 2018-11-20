package storageservice

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/golang/glog"

	"github.com/caicloud/storage-admin/pkg/admin/content"
	"github.com/caicloud/storage-admin/pkg/admin/helper"
	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

func HandleStorageServiceList(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var (
			fErr errors.FormatError
			resp apiv1a1.ListStorageServiceResponse
		)
		kc := c.GetClient()
		// pre work
		typeName, name, start, limit, fe := handleStorageServiceListPreWork(request)
		if fe != nil {
			glog.Errorf("handleStorageServiceListPreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageServiceList[type=%s][name=%s][start=%d][limit=%d]",
			typeName, name, start, limit)
		glog.Infof("%s start", logPath)

		// list
		aServiceList, missTypeList, e := helper.ListStorageService(kc, typeName, name)
		if e != nil {
			glog.Errorf("%s ListStorageService failed, %v", logPath, e)
			fErr.SetErrorInternalServerError(e)
			response.WriteHeaderAndEntity(fErr.Code, fErr.Api())
			return
		}
		if len(missTypeList) > 0 { // just alert, the service that miss type won't appeared in resp items
			glog.Errorf("%s got types not exist: %v", logPath, missTypeList)
		}

		// pack response
		resp.MetaData.Total = len(aServiceList)
		if start >= len(aServiceList) {
			glog.Warningf("%s requested start=%d out of total=%d", logPath, start, len(aServiceList))
			response.WriteHeaderAndEntity(http.StatusOK, resp)
			return
		}
		end := util.GetStartLimitEnd(start, limit, len(aServiceList))
		resp.Items = aServiceList[start:end]

		// return
		glog.Infof("%s done [total=%d]", logPath, len(aServiceList))
		response.WriteHeaderAndEntity(http.StatusOK, resp)
		return
	}
}

func HandleStorageServiceCreate(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var resp apiv1a1.CreateStorageServiceResponse
		kc := c.GetClient()
		// pre work
		req, fe := handleStorageServiceCreatePreWork(request)
		if fe != nil {
			glog.Errorf("handleStorageServiceCreatePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageServiceCreate[name=%s][type=%s][alias=%s][desc=%d]",
			req.Name, req.Type, req.Alias, len(req.Description))
		glog.Infof("%s start", logPath)

		// get type
		tp, fe := helper.GetStorageTypeInner(req.Type, kc)
		if fe != nil {
			glog.Errorf("%s GetStorageTypeInner failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		// pre work & check
		switch tp.Provisioner {
		case kubernetes.StorageClassProvisionerGlusterfs:
			fe = helper.CheckAndPreworkForGlusterFS(req.Name, req.Parameters, tp, kc)
		default:
			fe = util.CheckStorageServiceParameters(req.Parameters, tp.RequiredParameters)
		}
		// check and translate
		if fe != nil {
			glog.Errorf("%s pre work and check failed, %s", logPath, fe.Message)
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		service := util.DefaultTranslatorApiToMeta().StorageServiceNewFromCreate(req)

		// create
		re, fe := helper.CreateStorageService(service, kc)
		if fe != nil {
			glog.Errorf("%s CreateStorageService failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		util.DefaultTranslatorMetaToApi().StorageServiceSet(&resp, re, tp)
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusCreated, resp)
	}
}

func HandleStorageServiceGet(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var (
			resp apiv1a1.ReadStorageServiceResponse
		)
		kc := c.GetClient()
		// pre work
		name, fe := handleStorageServiceGetPreWork(request)
		if fe != nil {
			glog.Errorf("HandleStorageServiceCreatePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("handleStorageServiceCreate[name=%s]", name)
		glog.Infof("%s start", logPath)

		// get service
		service, tp, fe := helper.GetStorageServiceAndType(name, kc)
		if fe != nil {
			glog.Errorf("%s GetStorageServiceAndType failed, %v", logPath, fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		util.DefaultTranslatorMetaToApi().StorageServiceSet(&resp, service, tp)
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusOK, resp)
	}
}

func HandleStorageServiceUpdate(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var resp apiv1a1.UpdateStorageServiceResponse
		kc := c.GetClient()
		// pre work
		name, req, fe := handleStorageServiceUpdatePreWork(request)
		if fe != nil {
			glog.Errorf("handleStorageServiceUpdatePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageServiceUpdate[name=%s]", name)
		glog.Infof("%s start", logPath)

		// get pre version and check status
		service, tp, fe := helper.GetStorageServiceAndType(name, kc)
		if fe != nil {
			glog.Errorf("%s GetStorageServiceAndType failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		util.SetObjectAlias(service, req.Alias)
		util.SetObjectDescription(service, req.Description)

		// update
		re, fe := helper.UpdateStorageService(service, kc)
		if fe != nil {
			glog.Errorf("%s UpdateStorageService failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		util.DefaultTranslatorMetaToApi().StorageServiceSet(&resp, re, tp)
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusOK, resp)
	}
}

func HandleStorageServiceDelete(c content.Interface) func(*restful.Request, *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		kc := c.GetClient()
		// pre work
		name, fe := handleStorageServiceDeletePreWork(request)
		if fe != nil {
			glog.Errorf("handleStorageServiceDeletePreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("HandleStorageServiceDelete[name=%s]", name)
		glog.Infof("%s start", logPath)

		// delete service
		if fe = helper.TerminateStorageService(name, kc); fe != nil {
			glog.Errorf("%s TerminateStorageService failed, %v", logPath, fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// return
		glog.Infof("%s done", logPath)
		response.WriteHeaderAndEntity(http.StatusNoContent, nil)
	}
}
