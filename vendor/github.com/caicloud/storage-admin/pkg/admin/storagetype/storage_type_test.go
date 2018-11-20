package storagetype

import (
	"fmt"
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/kubernetes/fake"
	tt "github.com/caicloud/storage-admin/pkg/util/testtools"
)

func TestHandleStorageTypeList(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId       string
		runtimeObjs  []runtime.Object
		start, limit string
		isErr        bool
		num          int
	}
	testCaseTest := func(c *testCase) {
		var (
			tpList apiv1a1.ListStorageTypeResponse
		)
		describe := fmt.Sprintf("%s [%s:%s] get %d/%d err=%v", c.caseId, c.start, c.limit, c.num,
			len(c.runtimeObjs), c.isErr)
		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(SubPathListType).
			Param(constants.ParameterStart, c.start).Param(constants.ParameterLimit, c.limit)
		// run
		code, b, e := tt.Get(u, http.StatusOK, &tpList)
		switch {
		case c.isErr && e != nil:
			t.Logf("test case <%s> done", describe)
			return
		case (c.isErr && e == nil) || (!c.isErr && e != nil):
			t.Fatalf("test case <%s> Get unexpected error, resp[%d]: %v -> err: %v",
				describe, code, string(b), e)
		}

		if len(tpList.Items) != c.num {
			t.Fatalf("test case <%s> servList:[%d!=%d] %v", describe, len(tpList.Items), c.num, string(b))
		}
		t.Logf("test case <%s> done", describe)
	}
	tp := tt.DefaultStorageTypeBuilder().Get()
	otp := tt.DefaultStorageTypeBuilder().ChangeName().Get()
	testCaseList := []testCase{
		// normal
		{
			caseId: "case 0", runtimeObjs: []runtime.Object{tp, otp},
			start: "", limit: "",
			isErr: false, num: 2,
		},
		{
			caseId: "case 1-0", runtimeObjs: []runtime.Object{tp, otp},
			start: "", limit: "1",
			isErr: false, num: 1,
		},
		{
			caseId: "case 2-0", runtimeObjs: []runtime.Object{tp, otp},
			start: "1", limit: "",
			isErr: false, num: 1,
		},
		{
			caseId: "case 3-0", runtimeObjs: []runtime.Object{tp, otp},
			start: "1", limit: "1",
			isErr: false, num: 1,
		},
		// normal-edge
		{
			caseId: "case 1-0", runtimeObjs: []runtime.Object{tp, otp},
			start: "", limit: "2",
			isErr: false, num: 2,
		},
		{
			caseId: "case 1-0", runtimeObjs: []runtime.Object{tp},
			start: "", limit: "2",
			isErr: false, num: 1,
		},
		{
			caseId: "case 2-0", runtimeObjs: []runtime.Object{tp, otp},
			start: "0", limit: "",
			isErr: false, num: 2,
		},
		{
			caseId: "case 2-0", runtimeObjs: []runtime.Object{tp, otp},
			start: "2", limit: "",
			isErr: false, num: 0,
		},
		{
			caseId: "case 3-0", runtimeObjs: []runtime.Object{tp, otp},
			start: "2", limit: "1",
			isErr: false, num: 0,
		},
		{
			caseId: "case 3-0", runtimeObjs: []runtime.Object{tp},
			start: "2", limit: "1",
			isErr: false, num: 0,
		},
		// error
		{
			caseId: "case 1-1", runtimeObjs: []runtime.Object{tp, otp},
			start: "", limit: "a",
			isErr: true, num: 0,
		},
		{
			caseId: "case 1-1", runtimeObjs: []runtime.Object{tp, otp},
			start: "", limit: "0",
			isErr: true, num: 0,
		},
		{
			caseId: "case 2-1", runtimeObjs: []runtime.Object{tp, otp},
			start: "a", limit: "",
			isErr: true, num: 0,
		},
		{
			caseId: "case 2-2", runtimeObjs: []runtime.Object{tp, otp},
			start: "-1", limit: "",
			isErr: true, num: 0,
		},
		{
			caseId: "case 3-1", runtimeObjs: []runtime.Object{tp, otp},
			start: "a", limit: "0",
			isErr: true, num: 0,
		},
		{
			caseId: "case 3-2", runtimeObjs: []runtime.Object{tp, otp},
			start: "0", limit: "a",
			isErr: true, num: 0,
		},
		{
			caseId: "case 3-3", runtimeObjs: []runtime.Object{tp, otp},
			start: "-1", limit: "0",
			isErr: true, num: 0,
		},
		{
			caseId: "case 3-3", runtimeObjs: []runtime.Object{tp, otp},
			start: "0", limit: "0",
			isErr: true, num: 0,
		},
		{
			caseId: "case 3-3", runtimeObjs: []runtime.Object{tp, otp},
			start: "-1", limit: "0",
			isErr: true, num: 0,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}
