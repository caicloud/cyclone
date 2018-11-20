package storageclass

import (
	"testing"

	tntv1a1 "github.com/caicloud/clientset/pkg/apis/tenant/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/kubernetes/fake"
	"github.com/caicloud/storage-admin/pkg/util"
	tt "github.com/caicloud/storage-admin/pkg/util/testtools"
)

func testClassVarInit() (kc *fake.Clientset, c *Controller) {
	kc = fake.NewSimpleClientset()
	c, _ = NewController(kc)
	return
}

func TestController_deletePvcInAllNamespace(t *testing.T) {
	kc, c := testClassVarInit()
	ns := tt.DefaultNamespaceBuilder().Get()
	pvc := tt.DefaultPvcBuilder().Get()
	apvc := tt.DefaultPvcBuilder().ChangeName().ChangeClass().Get()

	kc.Clientset = fake.NewSimpleFakeKubeClientset(ns, pvc, apvc)

	// execute
	_, e := c.deletePvcInAllNamespace(tt.ClassNameDefault)
	if e != nil {
		t.Fatalf("deletePvcInAllNamespace %s failed, %v", tt.ClassNameDefault, e)
	}
	// check pvc should be deleted
	_, e = kc.CoreV1().PersistentVolumeClaims(pvc.Namespace).Get(pvc.Name, metav1.GetOptions{})
	if !kubernetes.IsNotFound(e) {
		t.Fatalf("get pvc deleted %s/%s got unexpected error, %v", pvc.Namespace, pvc.Name, e)
	}
	t.Logf("get pvc deleted %s/%s check done", pvc.Namespace, pvc.Name)
	// check pvc in other class
	_, e = kc.CoreV1().PersistentVolumeClaims(apvc.Namespace).Get(apvc.Name, metav1.GetOptions{})
	if e != nil {
		t.Fatalf("get pvc other class %s/%s got unexpected error, %v", apvc.Namespace, apvc.Name, e)
	}
	t.Logf("get pvc other class %s/%s check done", apvc.Namespace, apvc.Name)
}

func TestController_processSync(t *testing.T) {
	kc, c := testClassVarInit()

	type testCase struct {
		describe string
		saObjs   []runtime.Object
		obj      *storagev1.StorageClass
		isErr    bool
	}
	tp := tt.DefaultStorageTypeBuilder().Get()
	ss := tt.DefaultStorageServiceBuilder().Get()
	sc := tt.DefaultStorageClassBuilder().Get()
	testCaseList := []testCase{
		{
			describe: "simple sync",
			saObjs:   []runtime.Object{tp, ss},
			obj:      sc.DeepCopy(),
			isErr:    false,
		},
	}
	testFunc := func(tc *testCase) {
		kc.Clientset = fake.NewSimpleFakeKubeClientset(append(tc.saObjs, tc.obj)...)
		e := c.processSync(tc.obj.DeepCopy())
		if (tc.isErr && e == nil) || (!tc.isErr && e != nil) {
			t.Fatalf("test case <%s> unexpected error %v", tc.describe, e)
		}
		if !tc.isErr {
			sc, e := kc.StorageV1().StorageClasses().Get(tc.obj.Name, metav1.GetOptions{})
			if e != nil {
				t.Fatalf("test case <%s> check class failed, %v", tc.describe, e)
			}
			if !util.HasClassFinalizer(sc.Finalizers) {
				t.Fatalf("test case <%s> check class failed, finalizers missing, %v", tc.describe, sc.Finalizers)
			}
		}
		t.Logf("test case <%s> done", tc.describe)
	}

	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}

func TestStorageClassController_updateQuotaInAllNamespace(t *testing.T) {
	kc, c := testClassVarInit()

	type testCase struct {
		describe string
		ns       *corev1.Namespace
		rq       *corev1.ResourceQuota
	}

	ns := tt.DefaultNamespaceBuilder().Get()
	rq := tt.ObjectDefaultQuota()
	arq := tt.ObjectDefaultQuota() // rq.DeepCopy()
	tt.ResListSetClassQuota(tt.ClassNameDefault, arq.Spec.Hard)

	testCaseList := []testCase{
		{
			describe: "delete in exist",
			ns:       ns,
			rq:       rq,
		}, {
			describe: "delete in not exist",
			ns:       ns,
			rq:       arq,
		},
	}
	testFunc := func(tc *testCase) {
		kc.Clientset = fake.NewSimpleFakeKubeClientset(tc.ns, tc.rq)
		if _, e := c.updateQuotaInAllNamespace(tt.ClassNameDefault); e != nil {
			t.Fatalf("test case <%s> unexpected error %v", tc.describe, e)
		}
		rq, e := kc.CoreV1().ResourceQuotas(tc.rq.Namespace).Get(tc.rq.Name, metav1.GetOptions{})
		if e != nil {
			t.Fatalf("test case <%s> check quota failed, %v", tc.describe, e)
		}
		ok := tt.CheckClassInResList(tt.ClassNameDefault, rq.Spec.Hard)
		if ok {
			t.Fatalf("test case <%s> check quota failed, quota not deleted, %v", tc.describe, rq.Spec.Hard)
		}
		t.Logf("test case <%s> done", tc.describe)
	}

	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}

func TestStorageClassController_updateTenants(t *testing.T) {
	kc, c := testClassVarInit()

	type testCase struct {
		describe string
		tnt      *tntv1a1.Tenant
	}

	tnt := tt.ObjectDefaultTenant()
	tntSet := tt.ObjectDefaultTenant()
	tt.TenantSetClassQuota(tntSet, tt.ClassNameDefault)

	testCaseList := []testCase{
		{
			describe: "delete in exist",
			tnt:      tntSet,
		}, {
			describe: "delete in not exist",
			tnt:      tnt,
		},
	}
	testFunc := func(tc *testCase) {
		kc.Clientset = fake.NewSimpleFakeKubeClientset(tc.tnt)
		if _, e := c.updateTenants(tt.ClassNameDefault); e != nil {
			t.Fatalf("test case <%s> unexpected error %v", tc.describe, e)
		}
		tnt, e := kc.TenantV1alpha1().Tenants().Get(tc.tnt.Name, metav1.GetOptions{})
		if e != nil {
			t.Fatalf("test case <%s> check quota failed, %v", tc.describe, e)
		}
		ok := tt.CheckClassInResList(tt.ClassNameDefault,
			tnt.Spec.Quota, tnt.Status.Hard, tnt.Status.Used, tnt.Status.ActualUsed)
		if ok {
			t.Fatalf("test case <%s> check quota failed, quota not deleted, %v | %v | %v | %v", tc.describe,
				tnt.Spec.Quota, tnt.Status.Hard, tnt.Status.Used, tnt.Status.ActualUsed)
		}
		t.Logf("test case <%s> done", tc.describe)
	}

	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}

func TestStorageClassController_updatePartitions(t *testing.T) {
	kc, c := testClassVarInit()

	type testCase struct {
		describe string
		pt       *tntv1a1.Partition
	}

	pt := tt.ObjectDefaultPartition()
	ptSet := tt.ObjectDefaultPartition()
	tt.PartitionSetClassQuota(ptSet, tt.ClassNameDefault)

	testCaseList := []testCase{
		{
			describe: "delete in exist",
			pt:       ptSet,
		}, {
			describe: "delete in not exist",
			pt:       pt,
		},
	}
	testFunc := func(tc *testCase) {
		kc.Clientset = fake.NewSimpleFakeKubeClientset(tc.pt)
		if _, e := c.updateTenants(tt.ClassNameDefault); e != nil {
			t.Fatalf("test case <%s> unexpected error %v", tc.describe, e)
		}
		pt, e := kc.TenantV1alpha1().Partitions().Get(tc.pt.Name, metav1.GetOptions{})
		if e != nil {
			t.Fatalf("test case <%s> check quota failed, %v", tc.describe, e)
		}
		ok := tt.CheckClassInResList(tt.ClassNameDefault, pt.Spec.Quota, pt.Status.Hard, pt.Status.Used)
		if ok {
			t.Fatalf("test case <%s> check quota failed, quota not deleted, %v | %v | %v", tc.describe,
				pt.Spec.Quota, pt.Status.Hard, pt.Status.Used)
		}
		t.Logf("test case <%s> done", tc.describe)
	}

	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}

func TestStorageClassController_updateClusterQuotas(t *testing.T) {
	kc, c := testClassVarInit()

	type testCase struct {
		describe string
		cq       *tntv1a1.ClusterQuota
	}

	cq := tt.ObjectDefaultClusterQuota()
	cqQuota := tt.ObjectDefaultClusterQuota()
	tt.ClusterQuotaSetClassQuota(cqQuota, tt.ClassNameDefault)
	cqRatio := tt.ObjectDefaultClusterQuota()
	tt.ClusterQuotaSetClassRatio(cqRatio, tt.ClassNameDefault)
	cqAll := tt.ObjectDefaultClusterQuota()
	tt.ClusterQuotaSetClassRatio(cqAll, tt.ClassNameDefault)
	tt.ClusterQuotaSetClassRatio(cqAll, tt.ClassNameDefault)

	testCaseList := []testCase{
		{
			describe: "delete in quota exist",
			cq:       cqQuota,
		}, {
			describe: "delete in ratio exist",
			cq:       cqRatio,
		}, {
			describe: "delete in all exist",
			cq:       cqAll,
		}, {
			describe: "delete in not exist",
			cq:       cq,
		},
	}
	testFunc := func(tc *testCase) {
		kc.Clientset = fake.NewSimpleFakeKubeClientset(tc.cq)
		if _, e := c.updateTenants(tt.ClassNameDefault); e != nil {
			t.Fatalf("test case <%s> unexpected error %v", tc.describe, e)
		}
		cq, e := kc.TenantV1alpha1().ClusterQuotas().Get(tc.cq.Name, metav1.GetOptions{})
		if e != nil {
			t.Fatalf("test case <%s> check quota failed, %v", tc.describe, e)
		}
		ok := tt.CheckClassInResList(tt.ClassNameDefault,
			cq.Status.Total, cq.Status.Allocated, cq.Status.SystemUsed, cq.Status.Used,
			cq.Status.Capacity, cq.Status.Allocatable, cq.Status.Unavailable)
		if ok {
			t.Fatalf("test case <%s> check quota failed, quota not deleted, %v | %v | %v | %v | %v | %v | %v",
				tc.describe,
				cq.Status.Total, cq.Status.Allocated, cq.Status.SystemUsed, cq.Status.Used,
				cq.Status.Capacity, cq.Status.Allocatable, cq.Status.Unavailable)
		}
		ok = tt.CheckClassInRatio(tt.ClassNameDefault, cq.Spec.Ratio)
		if ok {
			t.Fatalf("test case <%s> check ratio failed, ratio not deleted, %v", tc.describe, cq.Spec.Ratio)
		}
		t.Logf("test case <%s> done", tc.describe)
	}

	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}

// tenant client has no fake client, so cancel related tests
func TestController_processTerminate(t *testing.T) {
	kc, c := testClassVarInit()

	// delete quota and pvc has independent test

	// terminate normally
	// init
	sc := tt.DefaultStorageClassBuilder().AppendGfsParam().SetTerminated().Get()
	secret := tt.DefaultSecretBuilder().Get()
	kc.Clientset = fake.NewSimpleFakeKubeClientset(sc, secret)
	// do
	if e := c.processTerminate(sc); e != nil {
		t.Fatalf("processTerminate %s normally failed, %v", sc.Name, e)
	}
	// check sc
	nsc, e := kc.StorageV1().StorageClasses().Get(sc.Name, metav1.GetOptions{})
	if e != nil {
		t.Fatalf("processTerminate %s normally get for check failed, %v", sc.Name, e)
	}
	if util.HasClassFinalizer(nsc.Finalizers) {
		t.Fatalf("processTerminate %s normally check failed, finalizers %v", sc.Name, nsc.Finalizers)
	}
	// check secret
	_, e = kc.CoreV1().Secrets(tt.DefaultGfsSecretNamespace).Get(tt.DefaultGfsSecretName, metav1.GetOptions{})
	if !kubernetes.IsNotFound(e) {
		t.Fatalf("processTerminate %s normally get secret for check failed, %v", sc.Name, e)
	}
	// done
	t.Logf("processTerminate %s normally done", sc.Name)
}
