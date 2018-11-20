package helper

import (
	"strconv"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/kubernetes/fake"
	tt "github.com/caicloud/storage-admin/pkg/util/testtools"
)

func TestListStorageService(t *testing.T) {
	type testCase struct {
		describe    string
		runtimeObjs []runtime.Object
		servListLen int
		missTypeNum int
	}
	kc := fake.NewClientset(nil)
	testCaseTest := func(c *testCase) {
		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		t.Logf("%v", c.runtimeObjs)
		// run
		servList, missType, fe := ListStorageService(kc, "", "")
		if fe != nil {
			t.Fatalf("test case <%s> unexpected error: %v", c.describe, fe.Error())
		}
		switch {
		case len(servList) != c.servListLen && len(missType) == c.missTypeNum:
			t.Fatalf("test case <%s> servList:[%d!=%d]", c.describe, len(servList), c.servListLen)
		case len(servList) == c.servListLen && len(missType) != c.missTypeNum:
			t.Fatalf("test case <%s> missType:[%d!=%d]", c.describe, len(missType), c.missTypeNum)
		case len(servList) != c.servListLen && len(missType) != c.missTypeNum:
			t.Fatalf("test case <%s> servList:[%d!=%d], missType:[%d!=%d]", c.describe,
				len(servList), c.servListLen, len(missType), c.missTypeNum)
		default:
			t.Logf("test case <%s> done", c.describe)
		}
	}
	tp := tt.DefaultStorageTypeBuilder().Get()
	ss := tt.DefaultStorageServiceBuilder().Get()
	otp := tt.DefaultStorageTypeBuilder().ChangeName().Get()
	oss := tt.DefaultStorageServiceBuilder().ChangeName().ChangeType().Get()
	testCaseList := []testCase{
		{
			describe:    "miss 0/1",
			runtimeObjs: []runtime.Object{tp, ss},
			servListLen: 1,
			missTypeNum: 0,
		},
		{
			describe:    "miss 1/1",
			runtimeObjs: []runtime.Object{ss},
			servListLen: 0,
			missTypeNum: 1,
		},
		{
			describe:    "miss 0/2",
			runtimeObjs: []runtime.Object{tp, ss, otp, oss},
			servListLen: 2,
			missTypeNum: 0,
		},
		{
			describe:    "miss 1/2",
			runtimeObjs: []runtime.Object{tp, ss, oss},
			servListLen: 1,
			missTypeNum: 1,
		},
		{
			describe:    "miss 2/2",
			runtimeObjs: []runtime.Object{ss, oss},
			servListLen: 0,
			missTypeNum: 2,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func TestListDataVolume(t *testing.T) {
	type testCase struct {
		describe string

		runtimeObjs  []runtime.Object
		namespace    string
		storageClass string
		name         string

		num int
	}
	kc := fake.NewClientset(nil)
	testCaseTest := func(c *testCase) {
		// init
		kc.Clientset = fake.NewSimpleFakeKubeClientset(c.runtimeObjs...)
		// run
		re, e := ListDataVolume(c.namespace, c.storageClass, c.name, kc)
		if e != nil {
			t.Fatalf("test case <%s> unexpected error: %v", c.describe, e)
		}
		if len(re) != c.num {
			t.Fatalf("test case <%s> [%d!=%d]", c.describe, len(re), c.num)
		}
		t.Logf("test case <%s> done", c.describe)
	}

	pvc := tt.DefaultPvcBuilder().Get()               // pvc: namespace default, name default, class default
	opvc := tt.DefaultPvcBuilder().ChangeName().Get() // pvc other
	testCaseList := []testCase{
		{
			describe:     "empty-all",
			runtimeObjs:  []runtime.Object{},
			namespace:    tt.NamespaceNameDefault,
			storageClass: "",
			name:         "",
			num:          0,
		},
		{
			describe:     "all-one",
			runtimeObjs:  []runtime.Object{pvc},
			namespace:    tt.NamespaceNameDefault,
			storageClass: "",
			name:         "",
			num:          1,
		},
		{
			describe:     "all-two",
			runtimeObjs:  []runtime.Object{pvc, opvc},
			namespace:    tt.NamespaceNameDefault,
			storageClass: "",
			name:         "",
			num:          2,
		},
		{
			describe:     "wrong namespace",
			runtimeObjs:  []runtime.Object{pvc},
			namespace:    tt.NamespaceNameOther,
			storageClass: "",
			name:         "",
			num:          0,
		},
		{
			describe:     "wrong class",
			runtimeObjs:  []runtime.Object{pvc},
			namespace:    tt.NamespaceNameDefault,
			storageClass: tt.ClassNameOther,
			name:         "",
			num:          0,
		},
		{
			describe:     "wrong name",
			runtimeObjs:  []runtime.Object{pvc},
			namespace:    tt.NamespaceNameDefault,
			storageClass: "",
			name:         tt.VolumeNameOther,
			num:          0,
		},
		{
			describe:     "upper name",
			runtimeObjs:  []runtime.Object{pvc},
			namespace:    tt.NamespaceNameDefault,
			storageClass: "",
			name:         strings.ToUpper(tt.VolumeNameDefault),
			num:          1,
		},
		{
			describe:     "wrong name and class",
			runtimeObjs:  []runtime.Object{pvc},
			namespace:    tt.NamespaceNameDefault,
			storageClass: tt.ClassNameOther,
			name:         tt.VolumeNameOther,
			num:          0,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}

func Test_checkGlusterfsGidRange(t *testing.T) {
	type testCase struct {
		describe string

		pm map[string]string

		isErr bool
	}
	testCaseTest := func(tc *testCase) {
		e := checkGlusterFsGidRange(tc.pm)
		if (!tc.isErr && e != nil) || (tc.isErr && e == nil) {
			t.Fatalf("test case <%s> unexpected error: %v", tc.describe, e)
		}
		t.Logf("test case <%s> done", tc.describe)
	}

	min := kubernetes.StorageClassParamGidRangeMin
	max := kubernetes.StorageClassParamGidRangeMax
	mid := (min + max) / 2
	testCaseList := []testCase{
		{
			describe: "empty",
			pm:       map[string]string{},
			isErr:    false,
		}, {
			describe: "min only in edge",
			pm: map[string]string{
				kubernetes.StorageClassParamNameGidMin: strconv.Itoa(min),
			},
			isErr: false,
		}, {
			describe: "max only in edge",
			pm: map[string]string{
				kubernetes.StorageClassParamNameGidMax: strconv.Itoa(max),
			},
			isErr: false,
		}, {
			describe: "both in edge",
			pm: map[string]string{
				kubernetes.StorageClassParamNameGidMin: strconv.Itoa(min),
				kubernetes.StorageClassParamNameGidMax: strconv.Itoa(max),
			},
			isErr: false,
		}, {
			describe: "both in same",
			pm: map[string]string{
				kubernetes.StorageClassParamNameGidMin: strconv.Itoa(mid),
				kubernetes.StorageClassParamNameGidMax: strconv.Itoa(mid),
			},
			isErr: false,
		}, {
			describe: "min only out of edge min",
			pm: map[string]string{
				kubernetes.StorageClassParamNameGidMin: strconv.Itoa(min - 1),
			},
			isErr: true,
		}, {
			describe: "max only out of edge max",
			pm: map[string]string{
				kubernetes.StorageClassParamNameGidMax: strconv.Itoa(max + 1),
			},
			isErr: true,
		}, {
			describe: "both out of edge",
			pm: map[string]string{
				kubernetes.StorageClassParamNameGidMin: strconv.Itoa(min - 1),
				kubernetes.StorageClassParamNameGidMax: strconv.Itoa(max + 1),
			},
			isErr: true,
		}, {
			describe: "both but upside down",
			pm: map[string]string{
				kubernetes.StorageClassParamNameGidMin: strconv.Itoa(max),
				kubernetes.StorageClassParamNameGidMax: strconv.Itoa(min),
			},
			isErr: true,
		},
	}
	for i := range testCaseList {
		testCaseTest(&testCaseList[i])
	}
}
