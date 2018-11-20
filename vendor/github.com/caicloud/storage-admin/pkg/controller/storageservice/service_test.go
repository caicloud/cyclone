package storageservice

import (
	"fmt"
	"testing"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/kubernetes/fake"
	"github.com/caicloud/storage-admin/pkg/util"
	tt "github.com/caicloud/storage-admin/pkg/util/testtools"
)

func testServiceVarInit() (cluster string, kc *fake.Clientset, ccg kubernetes.ClusterClientsetGetter, c *Controller) {
	cluster = tt.ClusterNameDefault
	kc = fake.NewSimpleClientset()
	ccg = kubernetes.NewClusterClientsetGetter(map[string]kubernetes.Interface{})
	c, _ = NewController(kc, ccg)
	return
}

func TestStorageServiceController_processSync(t *testing.T) {
	_, kc, _, c := testServiceVarInit()
	tp := tt.DefaultStorageTypeBuilder().Get()
	ss := tt.DefaultStorageServiceBuilder().Get()
	sc := tt.DefaultStorageClassBuilder().Get()
	// osc := tt.ClassSetChangeName(tt.ClassSetChangeService(tt.ObjectDefaultStorageClass())) // other storage class

	// [discard] test list/update class failed -> hard to control error return
	// [discard] parameters map work test in util

	type testCase struct {
		describe      string
		resObj        []runtime.Object
		kubeObj       []runtime.Object
		remainClasses []string
	}
	testCaseList := []testCase{
		{
			describe:      "normal",
			resObj:        []runtime.Object{tp, ss},
			kubeObj:       []runtime.Object{sc},
			remainClasses: []string{sc.Name},
		}, {
			describe:      "service deleted",
			resObj:        []runtime.Object{tp},
			kubeObj:       []runtime.Object{sc},
			remainClasses: []string{},
		}, {
			describe:      "class not marked",
			resObj:        []runtime.Object{tp},
			kubeObj:       []runtime.Object{tt.ClassRemoveMark(sc.DeepCopy())},
			remainClasses: []string{sc.Name},
		},
	}
	testFunc := func(tc *testCase) {
		kc.Clientset = fake.NewSimpleFakeKubeClientset(append(tc.kubeObj, tc.resObj...)...)
		// process
		c.processSync("test trigger")
		scList, e := kc.StorageV1().StorageClasses().List(metav1.ListOptions{})
		if e != nil {
			t.Fatalf("test case <%s> check classes remain failed, %v", tc.describe, e)
		}
		remainList := tt.GetClassNameList(scList.Items)
		if !tt.IsStringsSame(remainList, tc.remainClasses) {
			t.Fatalf("test case <%s> check classes remain not same, want %v, got %v", tc.describe, tc.remainClasses, remainList)
		}
		t.Logf("test case <%s> done", tc.describe)
	}

	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}

func TestStorageServiceController_cleanDeletedService(t *testing.T) {
	_, kc, _, c := testServiceVarInit()

	type testCase struct {
		describe    string
		ss          *resv1b1.StorageService
		secret      *corev1.Secret
		serviceName string
		isErr       bool
	}
	testFunc := func(tc *testCase) {
		// init
		logPrefix := fmt.Sprintf("[%s]", tc.describe)
		if tc.isErr {
			logPrefix += "[err]"
		}
		var objs []runtime.Object
		if tc.ss != nil {
			objs = append(objs, tc.ss)
		}
		if tc.secret != nil {
			objs = append(objs, tc.secret)
		}
		kc.Clientset = fake.NewSimpleFakeKubeClientset(objs...)
		// run
		e := c.cleanDeletedService(tc.serviceName, kc)
		if (tc.isErr && e == nil) || (!tc.isErr && e != nil) {
			t.Fatalf("%s cleanDeletedService unexpected error, %v", logPrefix, e)
		}
		// check
		if !tc.isErr && tc.ss != nil {
			e := tt.CheckServiceSecret(kc, tc.ss.Name,
				tc.ss.Parameters[kubernetes.StorageClassParamNameRestUser],
				tc.ss.Parameters[kubernetes.StorageClassParamNameRestUserKey])
			if util.IsObjStorageAdminMarked(tc.secret) && e == nil {
				t.Fatalf("%s CheckServiceSecret not exist failed, should not have a secret", logPrefix)
			} else if !util.IsObjStorageAdminMarked(tc.secret) && e != nil {
				t.Fatalf("%s CheckServiceSecret exist failed, %v", logPrefix, e)
			}
		}
		t.Logf("%s done", logPrefix)
	}

	ssG := tt.DefaultStorageServiceBuilder().AppendGfsParam().Get()
	secret := tt.DefaultSecretBuilder().Get()
	secretUser := tt.DefaultSecretBuilder().DelLabel(constants.LabelKeyStorageAdminMarkKey).Get()

	testCaseList := []testCase{
		{
			describe:    "simple",
			ss:          ssG,
			secret:      secret,
			serviceName: tt.ServiceNameDefault,
		}, {
			describe:    "user created secret",
			ss:          ssG,
			secret:      secretUser,
			serviceName: tt.ServiceNameDefault,
		}, {
			describe:    "trigger",
			ss:          nil,
			secret:      nil,
			serviceName: "<trigger>",
		},
	}

	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}
