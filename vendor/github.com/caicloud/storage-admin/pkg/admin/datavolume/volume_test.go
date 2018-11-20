package datavolume

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

// TODO add some response check, fix empty path parameter set

func TestHandleDataVolumeList(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId      string
		runtimeObjs []runtime.Object
		namespace   string
		class, name *string
		isErr       bool
		num         int
	}
	testCaseTest := func(c *testCase) {
		var (
			pvcList apiv1a1.ListDataVolumeResponse
		)
		describe := fmt.Sprintf("[%s][num=%d][err=%v]", c.caseId, c.num, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		subPath := tt.SetPathParam(SubPathListClass, constants.ParameterCluster, tt.ClusterNameDefault)
		subPath = tt.SetPathParam(subPath, constants.ParameterPartition, c.namespace)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		if c.class != nil {
			u.Param(constants.ParameterStorageClass, *c.class)
		}
		if c.name != nil {
			u.Param(constants.ParameterName, *c.name)
		}
		// run
		code, b, e := tt.Get(u, http.StatusOK, &pvcList)
		switch {
		case c.isErr && e != nil:
			t.Logf("test case <%s> done", describe)
			return
		case (c.isErr && e == nil) || (!c.isErr && e != nil):
			t.Fatalf("test case <%s> Get unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(pvcList), e)
		}

		if len(pvcList.Items) != c.num {
			t.Fatalf("test case <%s> servList:[%d!=%d] %v", describe, len(pvcList.Items), c.num, string(b))
		}
		t.Logf("test case <%s> done", describe)
	}
	stringPoint := func(s string) *string { return &s }
	pvc := tt.DefaultPvcBuilder().Get()                                          // pvc
	pvcDiffName := tt.DefaultPvcBuilder().ChangeName().Get()                     // pvc diff in name
	pvcDiffNameNs := tt.DefaultPvcBuilder().ChangeName().ChangeNamespace().Get() // pvc diff in name and namespace
	pvcDiffNameClass := tt.DefaultPvcBuilder().ChangeName().ChangeClass().Get()  // pvc diff in name and class
	testCaseList := []testCase{
		{
			caseId:      "simple",
			runtimeObjs: []runtime.Object{pvc, pvcDiffName},
			namespace:   tt.NamespaceNameDefault,
			class:       nil, name: nil,
			isErr: false, num: 2,
		}, {
			caseId:      "namespace",
			runtimeObjs: []runtime.Object{pvc, pvcDiffNameNs},
			namespace:   tt.NamespaceNameDefault,
			class:       nil, name: nil,
			isErr: false, num: 1,
		}, {
			caseId:      "class",
			runtimeObjs: []runtime.Object{pvc, pvcDiffNameClass},
			namespace:   tt.NamespaceNameDefault,
			class:       stringPoint(tt.ClassNameDefault), name: nil,
			isErr: false, num: 1,
		}, {
			caseId:      "name",
			runtimeObjs: []runtime.Object{pvc, pvcDiffName},
			namespace:   tt.NamespaceNameDefault,
			class:       nil, name: stringPoint(tt.VolumeNameOther),
			isErr: false, num: 1,
		}, {
			caseId:      "class & name",
			runtimeObjs: []runtime.Object{pvc, pvcDiffNameClass},
			namespace:   tt.NamespaceNameDefault,
			class:       stringPoint(tt.ClassNameOther), name: stringPoint(tt.VolumeNameOther),
			isErr: false, num: 1,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleDataVolumeCreate(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId     string
		saRtObjs   []runtime.Object
		kubeRtObjs []runtime.Object
		namespace  string
		req        *apiv1a1.CreateDataVolumeRequest
		isErr      bool
	}
	testCaseTest := func(c *testCase) {
		var (
			resp apiv1a1.CreateDataVolumeResponse
		)
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(append(c.kubeRtObjs, c.saRtObjs...)...)
		subPath := tt.SetPathParam(SubPathCreateClass, constants.ParameterCluster, tt.ClusterNameDefault)
		subPath = tt.SetPathParam(subPath, constants.ParameterPartition, c.namespace)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Post(u, http.StatusCreated, c.req, &resp)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Post unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(resp), e)
		}
		t.Logf("test case <%s> done", describe)
	}
	sc := tt.DefaultStorageClassBuilder().Get() // class
	pvc := tt.DefaultPvcBuilder().Get()         // pvc
	scTerminated := tt.DefaultStorageClassBuilder().SetTerminated().Get()
	reqDefault := tt.ObjectDefaultVolumeCR()
	testCaseList := []testCase{
		{
			caseId:     "normal",
			saRtObjs:   []runtime.Object{},
			kubeRtObjs: []runtime.Object{sc},
			namespace:  tt.NamespaceNameDefault,
			req:        reqDefault,
			isErr:      false,
		}, {
			caseId:     "class miss",
			saRtObjs:   []runtime.Object{},
			kubeRtObjs: []runtime.Object{},
			namespace:  tt.NamespaceNameDefault,
			req:        reqDefault,
			isErr:      true,
		}, {
			caseId:     "class terminated",
			saRtObjs:   []runtime.Object{},
			kubeRtObjs: []runtime.Object{scTerminated},
			namespace:  tt.NamespaceNameDefault,
			req:        reqDefault,
			isErr:      true,
		}, {
			caseId:     "name empty",
			saRtObjs:   []runtime.Object{},
			kubeRtObjs: []runtime.Object{sc},
			req:        tt.VolumeCRSetName(tt.ObjectDefaultVolumeCR(), ""),
			namespace:  tt.NamespaceNameDefault,
			isErr:      true,
		}, {
			caseId:     "class empty",
			saRtObjs:   []runtime.Object{},
			kubeRtObjs: []runtime.Object{sc},
			req:        tt.VolumeCRSetClass(tt.ObjectDefaultVolumeCR(), ""),
			namespace:  tt.NamespaceNameDefault,
			isErr:      true,
		}, {
			caseId:     "access modes miss",
			saRtObjs:   []runtime.Object{},
			kubeRtObjs: []runtime.Object{sc},
			req:        tt.VolumeCRSetAccessModes(tt.ObjectDefaultVolumeCR(), nil),
			namespace:  tt.NamespaceNameDefault,
			isErr:      true,
		}, {
			caseId:     "size miss",
			saRtObjs:   []runtime.Object{},
			kubeRtObjs: []runtime.Object{sc},
			req:        tt.VolumeCRSetSize(tt.ObjectDefaultVolumeCR(), 0),
			namespace:  tt.NamespaceNameDefault,
			isErr:      true,
		}, {
			caseId:     "exist",
			saRtObjs:   []runtime.Object{},
			kubeRtObjs: []runtime.Object{sc, pvc},
			req:        reqDefault,
			namespace:  tt.NamespaceNameDefault,
			isErr:      true,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleDataVolumeGet(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId      string
		runtimeObjs []runtime.Object
		namespace   string
		name        string
		isErr       bool
	}
	testCaseTest := func(c *testCase) {
		var resp apiv1a1.ReadDataVolumeResponse
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		subPath := tt.SetPathParam(SubPathGetClass, constants.ParameterCluster, tt.ClusterNameDefault)
		subPath = tt.SetPathParam(subPath, constants.ParameterPartition, c.namespace)
		subPath = tt.SetPathParam(subPath, constants.ParameterVolume, c.name)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Get(u, http.StatusOK, &resp)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Get unexpected error, resp[%d]: %v -> %v -> err: %v",
				describe, code, string(b), tt.ToJson(resp), e)
		}
		t.Logf("test case <%s> done", describe)
	}
	pvc := tt.DefaultPvcBuilder().Get() // pvc
	testCaseList := []testCase{
		{
			caseId:      "normal",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   tt.NamespaceNameDefault,
			name:        tt.VolumeNameDefault,
			isErr:       false,
		}, {
			caseId:      "empty name",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   tt.NamespaceNameDefault,
			name:        "",
			isErr:       true,
		}, {
			caseId:      "empty namespace",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   "",
			name:        tt.VolumeNameDefault,
			isErr:       true,
		}, {
			caseId:      "miss object",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   tt.NamespaceNameDefault,
			name:        tt.VolumeNameOther,
			isErr:       true,
		}, // service and type actually not required
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestHandleDataVolumeDelete(t *testing.T) {
	port := tt.GetNewPort()
	kc := fake.NewClientset(nil)
	c := tt.NewFakeContent(map[string]kubernetes.Interface{tt.ClusterNameDefault: kc}, tt.ClusterNameDefault)

	addEndpointsFunc := tt.AddEndpointsFunc(c, AddEndpoints)
	s := tt.RunHttpServer(port, addEndpointsFunc, t)
	defer s.Close()

	type testCase struct {
		caseId      string
		runtimeObjs []runtime.Object
		namespace   string
		name        string
		isErr       bool
	}
	testCaseTest := func(c *testCase) {
		describe := fmt.Sprintf("[%s][err=%v]", c.caseId, c.isErr)

		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		subPath := tt.SetPathParam(SubPathDeleteClass, constants.ParameterCluster, tt.ClusterNameDefault)
		subPath = tt.SetPathParam(subPath, constants.ParameterPartition, c.namespace)
		subPath = tt.SetPathParam(subPath, constants.ParameterVolume, c.name)
		u := tt.NewUrl().Host("127.0.0.1").Port(port).RootPathDefault().SubPath(subPath)
		// run
		code, b, e := tt.Delete(u, http.StatusNoContent)
		if (c.isErr && e == nil) || (!c.isErr && e != nil) {
			t.Fatalf("test case <%s> Delete unexpected error, resp[%d]: %v -> err: %v",
				describe, code, string(b), e)
		}
		t.Logf("test case <%s> done", describe)
	}
	// nsSysName := constants.SystemNamespaceKubeSystem
	pvc := tt.DefaultPvcBuilder().Get() // pvc
	// pvcSysNs := tt.VolumeSetNamespace(pvc.DeepCopy(), nsSysName) // pvc
	testCaseList := []testCase{
		{
			caseId:      "normal",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   tt.NamespaceNameDefault,
			name:        tt.VolumeNameDefault,
			isErr:       false,
		}, {
			caseId:      "empty namespace",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   "",
			name:        tt.VolumeNameDefault,
			isErr:       true,
		}, {
			caseId:      "wrong namespace",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   tt.NamespaceNameOther,
			name:        tt.VolumeNameDefault,
			isErr:       true,
		}, {
			caseId:      "empty name",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   tt.NamespaceNameDefault,
			name:        "",
			isErr:       true,
		}, {
			caseId:      "wrong name",
			runtimeObjs: []runtime.Object{pvc},
			namespace:   tt.NamespaceNameDefault,
			name:        tt.VolumeNameOther,
			isErr:       true,
		}, {
			caseId:      "miss object",
			runtimeObjs: []runtime.Object{},
			namespace:   tt.NamespaceNameDefault,
			name:        tt.VolumeNameDefault,
			isErr:       true,
			// }, {
			// 	caseId:      "system namespace",
			// 	runtimeObjs: []runtime.Object{pvcSysNs},
			// 	namespace:   nsSysName,
			// 	name:        tt.VolumeNameDefault,
			// 	isErr:       true,
		}, // service and type actually not required
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}
