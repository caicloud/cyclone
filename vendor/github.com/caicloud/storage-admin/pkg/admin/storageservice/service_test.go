package storageservice

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

func TestHandleStorageServiceList(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId      string
		runtimeObjs []runtime.Object
		tp          string
		alias       string
		isErr       bool
		num         int
	}
	testCaseTest := func(c *testCase) {
		var (
			ssList apiv1a1.ListStorageServiceResponse
		)
		describe := fmt.Sprintf("[%s][num=%d][err=%v]", c.caseId, c.num, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(SubPathListService)
		if len(c.alias) > 0 {
			u.Param(constants.ParameterName, c.alias)
		}
		if len(c.tp) > 0 {
			u.Param(constants.ParameterStorageType, c.tp)
		}
		// run
		code, b, e := tt.Get(u, http.StatusOK, &ssList)
		switch {
		case c.isErr && e != nil:
			t.Logf("test case <%s> done", describe)
			return
		case (c.isErr && e == nil) || (!c.isErr && e != nil):
			t.Fatalf("test case <%s> Get unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(ssList), e)
		}

		if len(ssList.Items) != c.num {
			t.Fatalf("test case <%s> servList:[%d!=%d] %v", describe, len(ssList.Items), c.num, string(b))
		}
		t.Logf("test case <%s> done", describe)
	}
	tp := tt.DefaultStorageTypeBuilder().Get()                                             // type
	otp := tt.DefaultStorageTypeBuilder().ChangeName().Get()                               // other type
	ss := tt.DefaultStorageServiceBuilder().Get()                                          // service
	oss := tt.DefaultStorageServiceBuilder().ChangeName().ChangeAlias().ChangeType().Get() // other service (set name and type)
	ossTypeDiff := tt.DefaultStorageServiceBuilder().ChangeType().Get()                    // service diff only in type
	ossNameDiff := tt.DefaultStorageServiceBuilder().ChangeName().Get()                    // other service diff only in name
	testCaseList := []testCase{
		{
			caseId:      "type(01)+serv(0,2)",
			runtimeObjs: []runtime.Object{tp, ossNameDiff, ss},
			isErr:       false, num: 2,
		}, {
			caseId:      "type(01)+serv(1,1)",
			runtimeObjs: []runtime.Object{tp, oss, ss},
			isErr:       false, num: 1,
		}, {
			caseId:      "type(01)+serv(2,0)",
			runtimeObjs: []runtime.Object{tp, oss, ossTypeDiff},
			isErr:       false, num: 0,
		}, {
			caseId:      "type(11)+serv(1,1)",
			runtimeObjs: []runtime.Object{tp, otp, oss, ss},
			isErr:       false, num: 2,
		}, {
			caseId:      "filter type",
			runtimeObjs: []runtime.Object{tp, otp, oss, ss},
			isErr:       false,
			tp:          tt.TypeNameDefault,
			num:         1,
		}, {
			caseId:      "filter alias",
			runtimeObjs: []runtime.Object{tp, otp, oss, ss},
			isErr:       false,
			alias:       tt.ServiceAliasOther,
			num:         1,
		}, {
			caseId:      "filter alias contains",
			runtimeObjs: []runtime.Object{tp, otp, oss, ss},
			isErr:       false,
			alias:       tt.ServiceAliasDefault,
			num:         2,
		}, {
			caseId:      "filter both",
			runtimeObjs: []runtime.Object{tp, otp, oss, ss},
			isErr:       false,
			tp:          tt.TypeNameOther,
			alias:       tt.ServiceAliasOther,
			num:         1,
		}, {
			caseId:      "filter both no ret",
			runtimeObjs: []runtime.Object{tp, otp, oss, ss},
			isErr:       false,
			tp:          tt.TypeNameDefault,
			alias:       tt.ServiceAliasOther,
			num:         0,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleStorageServiceCreate(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId      string
		runtimeObjs []runtime.Object
		req         *apiv1a1.CreateStorageServiceRequest
		isErr       bool
		hasSecret   bool
	}
	testCaseTest := func(c *testCase) {
		var (
			resp apiv1a1.CreateStorageServiceResponse
		)
		describe := fmt.Sprintf("[%s]", c.caseId)
		if c.isErr {
			describe += "[err]"
		}
		if c.hasSecret {
			describe += "[secret]"
		}

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(SubPathCreateService)
		// run
		code, b, e := tt.Post(u, http.StatusCreated, c.req, &resp)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("%s Post unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(resp), e)
		}

		// secret
		if !c.isErr &&
			len(c.req.Parameters[kubernetes.StorageClassParamNameRestUser]) > 0 &&
			len(c.req.Parameters[kubernetes.StorageClassParamNameRestUserKey]) > 0 {
			e := tt.CheckServiceSecret(kc, c.req.Name,
				c.req.Parameters[kubernetes.StorageClassParamNameRestUser],
				c.req.Parameters[kubernetes.StorageClassParamNameRestUserKey])
			if c.hasSecret && e != nil {
				t.Fatalf("%s CheckServiceSecret failed, %v", describe, e)
			}
			if !c.hasSecret && e == nil {
				t.Fatalf("%s CheckServiceSecret failed, should not have a secret", describe)
			}
		}

		if !c.isErr && e == nil {
			if resp.Metadata.Description != c.req.Description {
				t.Fatalf("test case <%s> check update alias result failed, want '%v' got '%v'",
					describe, c.req.Alias, resp.Metadata.Alias)
			}
			if resp.Metadata.Description != c.req.Description {
				t.Fatalf("test case <%s> check update describe result failed, want '%v' got '%v'",
					describe, c.req.Description, resp.Metadata.Description)
			}
		}

		// done
		t.Logf("test case <%s> done", describe)
	}
	tp := tt.DefaultStorageTypeBuilder().Get()                   // type
	tpG := tt.DefaultStorageTypeBuilder().AppendGfsParam().Get() // type
	otp := tt.DefaultStorageTypeBuilder().ChangeName().Get()     // other type
	ss := tt.DefaultStorageServiceBuilder().Get()                // service
	secret := tt.DefaultSecretBuilder().Get()
	secretBadType := tt.DefaultSecretBuilder().SetType("AAA").Get()
	secretEmptyKey := tt.DefaultSecretBuilder().SetKey("").Get()
	secretBadKey := tt.DefaultSecretBuilder().SetKey("BBB").Get()
	testCaseList := []testCase{
		{
			caseId:      "normal",
			runtimeObjs: []runtime.Object{tp},
			req:         tt.NewDefaultServiceCRBuilder().Get(),
			isErr:       false,
		}, {
			caseId:      "type miss",
			runtimeObjs: []runtime.Object{otp},
			req:         tt.NewDefaultServiceCRBuilder().Get(),
			isErr:       true,
		}, {
			caseId:      "param miss",
			runtimeObjs: []runtime.Object{tp},
			req:         tt.NewDefaultServiceCRBuilder().SetDelParam().Get(),
			isErr:       false,
		}, {
			caseId:      "param extra",
			runtimeObjs: []runtime.Object{tp},
			req:         tt.NewDefaultServiceCRBuilder().SetAddParam().Get(),
			isErr:       true,
		}, {
			caseId:      "param wrong",
			runtimeObjs: []runtime.Object{tp},
			req:         tt.NewDefaultServiceCRBuilder().SetReplaceParam().Get(),
			isErr:       true,
		}, {
			caseId:      "exist",
			runtimeObjs: []runtime.Object{tp, ss},
			req:         tt.NewDefaultServiceCRBuilder().Get(),
			isErr:       true,
		}, {
			caseId:      "empty name",
			runtimeObjs: []runtime.Object{otp},
			req:         tt.NewDefaultServiceCRBuilder().SetName("").Get(),
			isErr:       true,
		}, {
			caseId:      "empty type",
			runtimeObjs: []runtime.Object{otp},
			req:         tt.NewDefaultServiceCRBuilder().SetType("").Get(),
			isErr:       true,
		}, {
			caseId:      "secret normal user+key",
			runtimeObjs: []runtime.Object{tpG},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameRestUser, tt.DefaultGfsRestUser).
				SetKeyValue(kubernetes.StorageClassParamNameRestUserKey, tt.DefaultGfsRestUserKey).
				Get(),
			isErr:     false,
			hasSecret: true,
		}, {
			caseId:      "secret normal secret",
			runtimeObjs: []runtime.Object{tpG, secret},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameSecretName, tt.DefaultGfsSecretName).
				SetKeyValue(kubernetes.StorageClassParamNameSecretNamespace, tt.DefaultGfsSecretNamespace).
				Get(),
			isErr:     false,
			hasSecret: true,
		}, {
			caseId:      "secret missing user",
			runtimeObjs: []runtime.Object{tpG},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameRestUser, "").
				SetKeyValue(kubernetes.StorageClassParamNameRestUserKey, tt.DefaultGfsRestUserKey).
				Get(),
			isErr:     true,
			hasSecret: false,
		}, {
			caseId:      "secret missing key",
			runtimeObjs: []runtime.Object{tpG},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameRestUser, tt.DefaultGfsRestUser).
				SetKeyValue(kubernetes.StorageClassParamNameRestUserKey, "").
				Get(),
			isErr:     true,
			hasSecret: false,
		}, {
			caseId:      "secret has secret namespace",
			runtimeObjs: []runtime.Object{tpG},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameRestUser, tt.DefaultGfsRestUser).
				SetKeyValue(kubernetes.StorageClassParamNameRestUserKey, tt.DefaultGfsRestUserKey).
				SetKeyValue(kubernetes.StorageClassParamNameSecretNamespace, tt.DefaultGfsSecretNamespace).
				Get(),
			isErr:     true,
			hasSecret: false,
		}, {
			caseId:      "secret has secret name",
			runtimeObjs: []runtime.Object{tpG},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameRestUser, tt.DefaultGfsRestUser).
				SetKeyValue(kubernetes.StorageClassParamNameRestUserKey, tt.DefaultGfsRestUserKey).
				SetKeyValue(kubernetes.StorageClassParamNameSecretName, tt.DefaultGfsSecretName).
				Get(),
			isErr:     true,
			hasSecret: false,
		}, {
			caseId:      "secret not exist",
			runtimeObjs: []runtime.Object{tpG},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameSecretName, tt.DefaultGfsSecretName).
				SetKeyValue(kubernetes.StorageClassParamNameSecretNamespace, tt.DefaultGfsSecretNamespace).
				Get(),
			isErr:     true,
			hasSecret: false,
		}, {
			caseId:      "secret user not match",
			runtimeObjs: []runtime.Object{tpG, secret},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameRestUser, tt.DefaultGfsRestUser+"-").
				SetKeyValue(kubernetes.StorageClassParamNameSecretName, tt.DefaultGfsSecretName).
				SetKeyValue(kubernetes.StorageClassParamNameSecretNamespace, tt.DefaultGfsSecretNamespace).
				Get(),
			isErr:     true,
			hasSecret: true,
		}, {
			caseId:      "secret key not match",
			runtimeObjs: []runtime.Object{tpG, secret},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameRestUserKey, tt.DefaultGfsRestUserKey+"-").
				SetKeyValue(kubernetes.StorageClassParamNameSecretName, tt.DefaultGfsSecretName).
				SetKeyValue(kubernetes.StorageClassParamNameSecretNamespace, tt.DefaultGfsSecretNamespace).
				Get(),
			isErr:     true,
			hasSecret: true,
		}, {
			caseId:      "secret empty key",
			runtimeObjs: []runtime.Object{tpG, secretEmptyKey},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameSecretName, tt.DefaultGfsSecretName).
				SetKeyValue(kubernetes.StorageClassParamNameSecretNamespace, tt.DefaultGfsSecretNamespace).
				Get(),
			isErr:     true,
			hasSecret: true,
		}, {
			caseId:      "secret bad key",
			runtimeObjs: []runtime.Object{tpG, secretBadKey},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameSecretName, tt.DefaultGfsSecretName).
				SetKeyValue(kubernetes.StorageClassParamNameSecretNamespace, tt.DefaultGfsSecretNamespace).
				Get(),
			isErr:     true,
			hasSecret: true,
		}, {
			caseId:      "secret type not match",
			runtimeObjs: []runtime.Object{tpG, secretBadType},
			req: tt.NewDefaultServiceCRBuilder().
				SetKeyValue(kubernetes.StorageClassParamNameSecretName, tt.DefaultGfsSecretName).
				SetKeyValue(kubernetes.StorageClassParamNameSecretNamespace, tt.DefaultGfsSecretNamespace).
				Get(),
			isErr:     true,
			hasSecret: true,
		}, {
			caseId:      "test alias",
			runtimeObjs: []runtime.Object{tp},
			req:         tt.NewDefaultServiceCRBuilder().SetAlias(tt.ServiceAliasOther).Get(),
			isErr:       false,
		}, {
			caseId:      "test describe",
			runtimeObjs: []runtime.Object{tp},
			req:         tt.NewDefaultServiceCRBuilder().SetDescription(tt.ServiceDescriptionOther).Get(),
			isErr:       false,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleStorageServiceGet(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId      string
		runtimeObjs []runtime.Object
		name        string
		isErr       bool
	}
	testCaseTest := func(c *testCase) {
		var resp apiv1a1.ReadStorageServiceResponse
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		subPath := tt.SetPathParam(SubPathGetService, constants.ParameterStorageService, c.name)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Get(u, http.StatusOK, &resp)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Get unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(resp), e)
		}
		t.Logf("test case <%s> done", describe)
	}
	tp := tt.DefaultStorageTypeBuilder().Get()    // type
	ss := tt.DefaultStorageServiceBuilder().Get() // service
	testCaseList := []testCase{
		{
			caseId:      "normal",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        tt.ServiceNameDefault,
			isErr:       false,
		}, {
			caseId:      "miss name",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        "",
			isErr:       true,
		}, {
			caseId:      "miss object",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        tt.ServiceNameOther,
			isErr:       true,
		}, {
			caseId:      "miss type",
			runtimeObjs: []runtime.Object{ss},
			name:        tt.ServiceNameDefault,
			isErr:       true,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleStorageServiceUpdate(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId      string
		runtimeObjs []runtime.Object
		name        string
		req         *apiv1a1.UpdateStorageServiceRequest
		isErr       bool
	}
	testCaseTest := func(c *testCase) {
		var resp apiv1a1.UpdateStorageServiceResponse
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		subPath := tt.SetPathParam(SubPathUpdateService, constants.ParameterStorageService, c.name)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Put(u, http.StatusOK, c.req, &resp)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Put unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(resp), e)
		}
		if !c.isErr && e == nil {
			if resp.Metadata.Description != c.req.Description {
				t.Fatalf("test case <%s> check update alias result failed, want '%v' got '%v'",
					describe, c.req.Alias, resp.Metadata.Alias)
			}
			if resp.Metadata.Description != c.req.Description {
				t.Fatalf("test case <%s> check update describe result failed, want '%v' got '%v'",
					describe, c.req.Description, resp.Metadata.Description)
			}
		}
		t.Logf("test case <%s> done", describe)
	}
	tp := tt.DefaultStorageTypeBuilder().Get()    // type
	ss := tt.DefaultStorageServiceBuilder().Get() // service
	reqDefault := tt.ObjectDefaultStorageServiceUR()
	testCaseList := []testCase{
		{
			caseId:      "normal",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        tt.ServiceNameDefault,
			req:         reqDefault,
			isErr:       false,
		}, {
			caseId:      "miss name",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        "",
			req:         reqDefault,
			isErr:       true,
		}, {
			caseId:      "miss object",
			runtimeObjs: []runtime.Object{tp},
			name:        tt.ServiceNameDefault,
			req:         reqDefault,
			isErr:       true,
		}, {
			caseId:      "miss type",
			runtimeObjs: []runtime.Object{ss},
			name:        tt.ServiceNameDefault,
			req:         reqDefault,
			isErr:       true,
		}, {
			caseId:      "empty",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        tt.ServiceNameDefault,
			req:         tt.ObjectMakeStorageServiceUR("", ""),
			isErr:       false,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleStorageServiceDelete(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId      string
		runtimeObjs []runtime.Object
		name        string
		isErr       bool
	}
	testCaseTest := func(c *testCase) {
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		subPath := tt.SetPathParam(SubPathDeleteService, constants.ParameterStorageService, c.name)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Delete(u, http.StatusNoContent)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Delete unexpected error, resp[%d]: %v -> err: %v",
				describe, code, string(b), e)
		}
		t.Logf("test case <%s> done", describe)
	}
	tp := tt.DefaultStorageTypeBuilder().Get()    // type
	ss := tt.DefaultStorageServiceBuilder().Get() // service
	testCaseList := []testCase{
		{
			caseId:      "normal",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        tt.ServiceNameDefault,
			isErr:       false,
		}, {
			caseId:      "miss name",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        "",
			isErr:       true,
		}, {
			caseId:      "miss object",
			runtimeObjs: []runtime.Object{tp, ss},
			name:        tt.ServiceNameOther,
			isErr:       true,
		}, {
			caseId:      "miss type",
			runtimeObjs: []runtime.Object{ss},
			name:        tt.ServiceNameDefault,
			isErr:       false,
		}, {
			caseId:      "terminated",
			runtimeObjs: []runtime.Object{},
			name:        tt.ServiceNameDefault,
			isErr:       true,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}
