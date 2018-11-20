package storagetype

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

var (
	SubPathListType = "/types"
)

func AddEndpoints(ws *restful.WebService, c content.Interface) {
	ws.Route(ws.GET(SubPathListType).To(HandleStorageTypeList(c)).Doc("list storage type").
		Param(util.QueryParamStart(ws)).
		Param(util.QueryParamLimit(ws)).
		Returns(http.StatusOK, "ok", apiv1a1.ListStorageTypeResponse{}))
}

func handleStorageTypeListPreWork(request *restful.Request) (start, limit int, fe *errors.FormatError) {
	return util.HandleSimpleListPreWork(request)
}

func HandleStorageTypeList(c content.Interface) func(request *restful.Request, response *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		var resp apiv1a1.ListStorageTypeResponse
		kc := c.GetClient()

		// pre work
		start, limit, fe := handleStorageTypeListPreWork(request)
		if fe != nil {
			glog.Errorf("handleStorageTypeListPreWork failed, %v", fe.String())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}
		logPath := fmt.Sprintf("handleStorageTypeList[start=%d][limit=%d]", start, limit)
		glog.Infof("%s start", logPath)

		// list
		typeList, fe := helper.ListStorageType(kc)
		if fe != nil {
			glog.Errorf("%s ListStorageType failed, %v", logPath, fe.Error())
			response.WriteHeaderAndEntity(fe.Code, fe.Api())
			return
		}

		// pack response
		resp.MetaData.Total = len(typeList.Items)
		if start >= len(typeList.Items) {
			glog.Warningf("%s requested start=%d out of total=%d", logPath, start, len(typeList.Items))
			response.WriteHeaderAndEntity(http.StatusOK, resp)
			return
		}
		end := util.GetStartLimitEnd(start, limit, len(typeList.Items))
		resp.Items = util.DefaultTranslatorMetaToApi().StorageTypeList(typeList.Items[start:end])

		// return
		glog.Infof("%s done [total=%d]", logPath, len(typeList.Items))
		response.WriteHeaderAndEntity(http.StatusOK, resp)
		return
	}
}
