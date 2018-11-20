package storageclass

import (
	"fmt"
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/kubernetes/fake"
	"github.com/caicloud/storage-admin/pkg/util"
	tt "github.com/caicloud/storage-admin/pkg/util/testtools"
)

// TODO add some response check

func TestHandleStorageClassList(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId   string
		kubeObjs []runtime.Object
		tp       string
		alias    string
		isErr    bool
		num      int
	}
	testCaseTest := func(c *testCase) {
		var (
			scList apiv1a1.ListStorageClassResponse
		)
		describe := fmt.Sprintf("[%s][num=%d][err=%v]", c.caseId, c.num, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.kubeObjs...)
		subPath := tt.SetPathParam(SubPathListClass, constants.ParameterCluster, tt.ClusterNameDefault)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		if len(c.alias) > 0 {
			u.Param(constants.ParameterName, c.alias)
		}
		if len(c.tp) > 0 {
			u.Param(constants.ParameterStorageType, c.tp)
		}
		// run
		code, b, e := tt.Get(u, http.StatusOK, &scList)
		switch {
		case c.isErr && e != nil:
			t.Logf("test case <%s> done", describe)
			return
		case (c.isErr && e == nil) || (!c.isErr && e != nil):
			t.Fatalf("test case <%s> Get unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(scList), e)
		}

		if len(scList.Items) != c.num {
			t.Fatalf("test case <%s> servList:[%d!=%d] %v", describe, len(scList.Items), c.num, string(b))
		}
		t.Logf("test case <%s> done", describe)
	}
	sc := tt.DefaultStorageClassBuilder().Get()                               // class
	osc := tt.DefaultStorageClassBuilder().ChangeName().ChangeService().Get() // other class name and service diff
	oscNameDiff := tt.DefaultStorageClassBuilder().ChangeName().Get()         // other class name diff only
	oscOther := tt.DefaultStorageClassBuilder().ChangeName().ChangeType().
		ChangeService().ChangeAlias().Get() // other class name diff only

	testCaseList := []testCase{
		{
			caseId:   "type(11)+serv(11)+class(1,1)",
			kubeObjs: []runtime.Object{sc, osc},
			isErr:    false, num: 2,
		}, {
			caseId:   "type(01)+serv(01)+class(0,2)",
			kubeObjs: []runtime.Object{sc, oscNameDiff},
			isErr:    false, num: 2,
		}, {
			caseId:   "filter type",
			kubeObjs: []runtime.Object{sc, oscOther},
			isErr:    false,
			tp:       tt.TypeNameDefault,
			num:      1,
		}, {
			caseId:   "filter alias",
			kubeObjs: []runtime.Object{sc, oscOther},
			isErr:    false,
			alias:    tt.ClassAliasOther,
			num:      1,
		}, {
			caseId:   "filter alias more",
			kubeObjs: []runtime.Object{sc, oscOther},
			isErr:    false,
			alias:    tt.ClassAliasDefault,
			num:      2,
		}, {
			caseId:   "filter both",
			kubeObjs: []runtime.Object{sc, oscOther},
			isErr:    false,
			tp:       tt.TypeNameOther,
			alias:    tt.ClassAliasOther,
			num:      1,
		}, {
			caseId:   "filter both no ret",
			kubeObjs: []runtime.Object{sc, oscOther},
			isErr:    false,
			tp:       tt.TypeNameDefault,
			alias:    tt.ClassAliasOther,
			num:      0,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleStorageClassCreate(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId   string
		saObjs   []runtime.Object
		kubeObjs []runtime.Object
		secret   *corev1.Secret
		req      *apiv1a1.CreateStorageClassRequest
		isErr    bool
	}
	testCaseTest := func(c *testCase) {
		var (
			resp apiv1a1.CreateStorageClassResponse
		)
		describe := fmt.Sprintf("[%s]", c.caseId)
		if c.isErr {
			describe += "[err]"
		}
		if c.secret != nil {
			describe += "[secret]"
		}

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(append(c.saObjs, c.kubeObjs...)...)
		subPath := tt.SetPathParam(SubPathCreateClass, constants.ParameterCluster, tt.ClusterNameDefault)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Post(u, http.StatusCreated, c.req, &resp)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Post unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(resp), e)
		}

		// secret
		if !c.isErr {
			e := tt.CheckClassSecret(kc, c.req.Name, c.secret)
			if c.secret != nil && e != nil {
				t.Fatalf("%s CheckServiceSecret failed, %v", describe, e)
			}
			if c.secret == nil && e == nil {
				t.Fatalf("%s CheckServiceSecret failed, should not have a secret", describe)
			}
		}

		if !c.isErr && e == nil {
			objDes := util.GetObjectDescription(&resp)
			if objDes != c.req.Description {
				t.Fatalf("test case <%s> check update describe result failed, want '%v' got '%v'",
					describe, c.req.Description, objDes)
			}
			objAls := util.GetObjectAlias(&resp)
			if objAls != c.req.Alias {
				t.Fatalf("test case <%s> check update alias result failed, want '%v' got '%v'",
					describe, c.req.Alias, objAls)
			}
		}

		// done
		t.Logf("test case <%s> done", describe)
	}
	tp := tt.DefaultStorageTypeBuilder().Get()               // type
	otp := tt.DefaultStorageTypeBuilder().ChangeName().Get() // other type
	ss := tt.DefaultStorageServiceBuilder().Get()            // service
	sc := tt.DefaultStorageClassBuilder().Get()              // class
	ssG := tt.DefaultStorageServiceBuilder().AppendGfsParam().Get()
	secret := tt.DefaultSecretBuilder().Get()
	testCaseList := []testCase{
		{
			caseId: "normal",
			saObjs: []runtime.Object{tp, ss},
			req:    tt.NewDefaultClassCRBuilder().Get(),
			isErr:  false,
		}, {
			caseId: "service miss",
			saObjs: []runtime.Object{tp},
			req:    tt.NewDefaultClassCRBuilder().Get(),
			isErr:  true,
		}, {
			caseId: "type miss",
			saObjs: []runtime.Object{ss},
			req:    tt.NewDefaultClassCRBuilder().Get(),
			isErr:  true,
		}, {
			caseId: "param miss",
			saObjs: []runtime.Object{tp, ss},
			req:    tt.NewDefaultClassCRBuilder().SetDelParam().Get(),
			isErr:  false,
		}, {
			caseId: "param extra",
			saObjs: []runtime.Object{tp, ss},
			req:    tt.NewDefaultClassCRBuilder().SetAddParam().Get(),
			isErr:  true,
		}, {
			caseId: "param wrong",
			saObjs: []runtime.Object{tp, ss},
			req:    tt.NewDefaultClassCRBuilder().SetReplaceParam().Get(),
			isErr:  true,
		}, {
			caseId:   "exist",
			saObjs:   []runtime.Object{tp, ss},
			kubeObjs: []runtime.Object{sc},
			req:      tt.NewDefaultClassCRBuilder().Get(),
			isErr:    true,
		}, {
			caseId: "empty name",
			saObjs: []runtime.Object{otp},
			req:    tt.NewDefaultClassCRBuilder().SetName("").Get(),
			isErr:  true,
		}, {
			caseId: "empty service",
			saObjs: []runtime.Object{otp},
			req:    tt.NewDefaultClassCRBuilder().SetService("").Get(),
			isErr:  true,
		}, {
			caseId: "wrong service",
			saObjs: []runtime.Object{otp, ss},
			req:    tt.NewDefaultClassCRBuilder().SetService(tt.ServiceNameOther).Get(),
			isErr:  true,
		}, {
			caseId:   "secret normal",
			saObjs:   []runtime.Object{tp, ssG},
			kubeObjs: []runtime.Object{secret},
			secret:   secret,
			req:      tt.NewDefaultClassCRBuilder().Get(),
			isErr:    false,
		}, {
			caseId: "secret missing",
			saObjs: []runtime.Object{tp, ssG},
			secret: secret,
			req:    tt.NewDefaultClassCRBuilder().Get(),
			isErr:  true,
		}, {
			caseId: "test alias",
			saObjs: []runtime.Object{tp, ss},
			req:    tt.ClassCRSetAlias(tt.ObjectDefaultStorageClassCR(), tt.ClassAliasOther),
			isErr:  false,
		}, {
			caseId: "test describe",
			saObjs: []runtime.Object{tp, ss},
			req:    tt.ClassCRSetDescription(tt.ObjectDefaultStorageClassCR(), tt.ClassDescriptionOther),
			isErr:  false,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleStorageClassGet(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId   string
		kubeObjs []runtime.Object
		name     string
		isErr    bool
	}
	testCaseTest := func(c *testCase) {
		var resp apiv1a1.ReadStorageClassResponse
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.kubeObjs...)
		subPath := tt.SetPathParam(SubPathGetClass, constants.ParameterCluster, tt.ClusterNameDefault)
		subPath = tt.SetPathParam(subPath, constants.ParameterStorageClass, c.name)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Get(u, http.StatusOK, &resp)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Get unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(resp), e)
		}
		t.Logf("test case <%s> done", describe)
	}
	sc := tt.DefaultStorageClassBuilder().Get() // class
	testCaseList := []testCase{
		{
			caseId:   "normal",
			kubeObjs: []runtime.Object{sc},
			name:     tt.ClassNameDefault,
			isErr:    false,
		}, {
			caseId:   "empty name",
			kubeObjs: []runtime.Object{sc},
			name:     "",
			isErr:    true,
		}, {
			caseId:   "miss object",
			kubeObjs: []runtime.Object{},
			name:     tt.ClassNameDefault,
			isErr:    true,
		}, // service and type actually not required
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleStorageClassUpdate(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId   string
		kubeObjs []runtime.Object
		name     string
		req      *apiv1a1.UpdateStorageClassRequest
		isErr    bool
	}
	testCaseTest := func(c *testCase) {
		var resp apiv1a1.UpdateStorageClassResponse
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.kubeObjs...)
		subPath := tt.SetPathParam(SubPathUpdateClass, constants.ParameterCluster, tt.ClusterNameDefault)
		subPath = tt.SetPathParam(subPath, constants.ParameterStorageClass, c.name)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Put(u, http.StatusOK, c.req, &resp)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Put unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(resp), e)
		}
		if !c.isErr && e == nil {
			objDes := util.GetObjectDescription(&resp)
			if objDes != c.req.Description {
				t.Fatalf("test case <%s> check update describe result failed, want '%v' got '%v'",
					describe, c.req.Description, objDes)
			}
			objAls := util.GetObjectAlias(&resp)
			if objAls != c.req.Alias {
				t.Fatalf("test case <%s> check update alias result failed, want '%v' got '%v'",
					describe, c.req.Alias, objAls)
			}
		}
		t.Logf("test case <%s> done", describe)
	}
	sc := tt.DefaultStorageClassBuilder().Get() // class
	reqDefault := tt.ObjectDefaultStorageClassUR()
	testCaseList := []testCase{
		{
			caseId:   "normal",
			kubeObjs: []runtime.Object{sc.DeepCopy()},
			name:     tt.ClassNameDefault,
			req:      reqDefault,
			isErr:    false,
		}, {
			caseId:   "miss name",
			kubeObjs: []runtime.Object{sc.DeepCopy()},
			name:     "",
			req:      reqDefault,
			isErr:    true,
		}, {
			caseId: "miss object",
			name:   tt.ClassNameDefault,
			req:    reqDefault,
			isErr:  true,
		}, {
			caseId:   "other",
			kubeObjs: []runtime.Object{sc.DeepCopy()},
			name:     tt.ClassNameOther,
			req:      tt.ObjectDefaultStorageClassUR(),
			isErr:    true,
		}, {
			caseId:   "empty",
			kubeObjs: []runtime.Object{sc.DeepCopy()},
			name:     tt.ClassNameDefault,
			req:      tt.ObjectMakeStorageClassUR("", ""),
			isErr:    false,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleStorageClassDelete(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(fake.NewSimpleFakeKubeClientset())
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId   string
		saObjs   []runtime.Object
		kubeObjs []runtime.Object
		name     string
		isErr    bool
	}
	testCaseTest := func(c *testCase) {
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(append(c.saObjs, c.kubeObjs...)...)
		subPath := tt.SetPathParam(SubPathDeleteClass, constants.ParameterCluster, tt.ClusterNameDefault)
		subPath = tt.SetPathParam(subPath, constants.ParameterStorageClass, c.name)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Delete(u, http.StatusNoContent)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Delete unexpected error, resp[%d]: %v -> err: %v",
				describe, code, string(b), e)
		}
		t.Logf("test case <%s> done", describe)
	}
	tp := tt.DefaultStorageTypeBuilder().Get()                                           // type
	ss := tt.DefaultStorageServiceBuilder().Get()                                        // service
	sc := tt.DefaultStorageClassBuilder().Get()                                          // class
	scTerminated := tt.DefaultStorageClassBuilder().SetTerminated().Get()                // class
	scSys := tt.DefaultStorageClassBuilder().SetName(constants.SystemStorageClass).Get() // class
	testCaseList := []testCase{
		{
			caseId:   "normal",
			saObjs:   []runtime.Object{tp, ss},
			kubeObjs: []runtime.Object{sc},
			name:     tt.ClassNameDefault,
			isErr:    false,
		}, {
			caseId:   "miss name",
			saObjs:   []runtime.Object{tp, ss},
			kubeObjs: []runtime.Object{sc},
			name:     "",
			isErr:    true,
		}, {
			caseId: "miss object",
			saObjs: []runtime.Object{tp, ss},
			name:   tt.ClassNameDefault,
			isErr:  true,
		}, {
			caseId:   "miss type",
			saObjs:   []runtime.Object{ss},
			kubeObjs: []runtime.Object{sc},
			name:     tt.ClassNameDefault,
			isErr:    false,
		}, {
			caseId:   "miss service",
			saObjs:   []runtime.Object{tp},
			kubeObjs: []runtime.Object{sc},
			name:     tt.ClassNameDefault,
			isErr:    false,
		}, {
			caseId:   "terminated",
			saObjs:   []runtime.Object{tp, ss},
			kubeObjs: []runtime.Object{scTerminated},
			name:     tt.ClassNameDefault,
			isErr:    false,
		}, {
			caseId:   "system object",
			saObjs:   []runtime.Object{tp, ss},
			kubeObjs: []runtime.Object{scSys},
			name:     constants.SystemStorageClass,
			isErr:    true,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}
