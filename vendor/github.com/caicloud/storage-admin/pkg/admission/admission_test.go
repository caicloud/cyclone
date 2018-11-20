package admission

import (
	"strings"
	"testing"

	admsv1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/kubernetes/fake"
	"github.com/caicloud/storage-admin/pkg/util"
	tt "github.com/caicloud/storage-admin/pkg/util/testtools"
)

func TestParseNamePath(t *testing.T) {
	type testCase struct {
		namePath  string
		isError   bool
		name      string
		namespace string
	}
	testFunc := func(tc *testCase) {
		namespace, name, e := ParseNamePath(tc.namePath, DefaultCertNamespace, DefaultCertName)
		if isErr := e != nil; isErr != tc.isError {
			t.Fatalf("test case <%s> failed, unexpected error %v", tc.namePath, e)
		}
		if !tc.isError {
			if name != tc.name {
				t.Fatalf("test case <%s> failed, name not right, want \"%s\", got \"%s\"", tc.namePath, tc.name, name)
			}
			if namespace != tc.namespace {
				t.Fatalf("test case <%s> failed, namespace not right, want \"%s\", got \"%s\"", tc.namePath, tc.namespace, namespace)
			}
		}
		t.Logf("test case <%s> done", tc.namePath)
	}
	name := "name-t"
	namespace := "namespace-t"
	testCaseList := []testCase{
		{
			namePath:  "",
			isError:   false,
			name:      DefaultCertName,
			namespace: DefaultCertNamespace,
		}, {
			namePath:  name,
			isError:   false,
			name:      name,
			namespace: DefaultCertNamespace,
		}, {
			namePath:  namespace + "/" + name,
			isError:   false,
			name:      name,
			namespace: namespace,
		}, {
			namePath:  "/" + name,
			isError:   false,
			name:      name,
			namespace: DefaultCertNamespace,
		}, {
			namePath:  namespace + "/",
			isError:   false,
			name:      DefaultCertName,
			namespace: namespace,
		}, {
			namePath:  "/",
			isError:   false,
			name:      DefaultCertName,
			namespace: DefaultCertNamespace,
		}, {
			namePath: namespace + "/" + name + "/",
			isError:  true,
		},
	}
	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}

func Test_admissionReviewCheckerPvc(t *testing.T) {
	kc := fake.NewSimpleClientset()

	labelNameNum := corev1.ResourceName(util.LabelKeyStorageQuotaNum(tt.ClassNameDefault))
	labelNameSize := corev1.ResourceName(util.LabelKeyStorageQuotaSize(tt.ClassNameDefault))
	emptyRq := tt.ObjectDefaultQuota()
	existRq := tt.ObjectDefaultQuota() // rq.DeepCopy()
	existRq.Spec.Hard[labelNameNum] = tt.GetQuantityResNumDefault()
	existRq.Spec.Hard[labelNameSize] = tt.GetQuantityResSizeDefault()
	numRq := tt.ObjectDefaultQuota() // rq.DeepCopy()
	numRq.Spec.Hard[labelNameNum] = tt.GetQuantityResNumDefault()
	sizeRq := tt.ObjectDefaultQuota() // rq.DeepCopy()
	sizeRq.Spec.Hard[labelNameSize] = tt.GetQuantityResSizeDefault()
	existRqNs := tt.QuotaSetChangeNamespace(tt.ObjectDefaultQuota()) // rq.DeepCopy() // namespace changed
	existRqNs.Spec.Hard[labelNameNum] = tt.GetQuantityResNumDefault()
	existRqNs.Spec.Hard[labelNameSize] = tt.GetQuantityResSizeDefault()
	existRqNs.Namespace = tt.NamespaceNameOther

	ns := tt.DefaultNamespaceBuilder().Get()
	pvc := tt.DefaultPvcBuilder().Get()
	ncPvc := tt.DefaultPvcBuilder().Get() // no class pvc
	delete(ncPvc.Annotations, kubernetes.LabelKeyKubeStorageClass)
	ncPvc.Spec.StorageClassName = nil

	getPvcRawString := func(pvc *corev1.PersistentVolumeClaim) string {
		sr := pvc.Spec.Resources.Requests["storage"]
		className := util.GetPVCClass(pvc)

		s := `{
	"name":"` + pvc.Name + `",
	"namespace":"` + pvc.Namespace + `",
	"creationTimestamp":null,
	"annotations":{`
		if len(className) > 0 {
			s += `"volume.beta.kubernetes.io/storage-class":"` + className + `",`
		}
		s += `"volume.beta.kubernetes.io/storage-provisioner":"` + pvc.Annotations["volume.beta.kubernetes.io/storage-provisioner"] + `"
	},
	"Spec":{
		"AccessModes":["ReadWriteMany"],
		"Selector":null,
		"Resources":{
			"Limits":null,
			"Requests":{"storage":"` + sr.String() + `"}
		},
		"VolumeName":""`

		if len(className) > 0 {
			s += `,"StorageClassName":"` + *pvc.Spec.StorageClassName + `"`
		}
		s += `},
	"Status":{
		"Phase":"Pending",
		"AccessModes":null,
		"Capacity":null
	}
}`
		return s
	}

	getPvcAdmissionReview := func(pvc *corev1.PersistentVolumeClaim) *admsv1.AdmissionReview {
		return &admsv1.AdmissionReview{
			Request: &admsv1.AdmissionRequest{
				Kind: metav1.GroupVersionKind{},
				Object: runtime.RawExtension{
					Raw:    []byte(getPvcRawString(pvc)),
					Object: nil,
				},
				OldObject: runtime.RawExtension{
					Raw:    nil,
					Object: nil,
				},
				Operation: admsv1.Create,
				Name:      pvc.Name,
				Namespace: pvc.Namespace,
				Resource:  *pvcResource,
			},
			Response: &admsv1.AdmissionResponse{
				Allowed: false,
			},
		}
	}

	type testCase struct {
		describe string
		objs     []runtime.Object
		pvc      *corev1.PersistentVolumeClaim
		wlClass  []string
		wlSpace  []string
		isError  bool
	}
	testFunc := func(tc *testCase) {
		kc.Clientset = fake.NewSimpleFakeKubeClientset(tc.objs...)
		pvcIgnoreNamespaces = getWhiteListMap(strings.Join(tc.wlSpace, ","))
		pvcIgnoreStorageClass = getWhiteListMap(strings.Join(tc.wlClass, ","))
		ar := getPvcAdmissionReview(tc.pvc)
		e := admissionReviewCheckerPvc(kc, ar)
		if isErr := e != nil; isErr != tc.isError {
			t.Fatalf("test case <%s> failed, unexpected error %v", tc.describe, e)
		}
		t.Logf("test case <%s> done", tc.describe)
	}
	testCaseList := []testCase{
		{
			describe: "normal",
			objs:     []runtime.Object{ns, existRq},
			pvc:      pvc,
			wlClass:  []string{""},
			wlSpace:  []string{""},
			isError:  false,
		}, {
			describe: "quota not exist",
			objs:     []runtime.Object{ns, emptyRq},
			pvc:      pvc,
			wlClass:  []string{""},
			wlSpace:  []string{""},
			isError:  true,
		}, {
			describe: "quota num only",
			objs:     []runtime.Object{ns, numRq},
			pvc:      pvc,
			wlClass:  []string{""},
			wlSpace:  []string{""},
			isError:  true,
		}, {
			describe: "quota size only",
			objs:     []runtime.Object{ns, sizeRq},
			pvc:      pvc,
			wlClass:  []string{""},
			wlSpace:  []string{""},
			isError:  true,
		}, {
			describe: "namespace not exist",
			objs:     []runtime.Object{tt.DefaultNamespaceBuilder().ChangeName().Get(), existRqNs},
			pvc:      pvc,
			wlClass:  []string{""},
			wlSpace:  []string{""},
			isError:  true,
		}, {
			describe: "class empty",
			objs:     []runtime.Object{ns, existRq},
			pvc:      ncPvc,
			wlClass:  []string{""},
			wlSpace:  []string{""},
			isError:  false,
		}, {
			describe: "class in white list",
			objs:     []runtime.Object{ns, emptyRq},
			pvc:      pvc,
			wlClass:  []string{*pvc.Spec.StorageClassName},
			wlSpace:  []string{""},
			isError:  false,
		}, {
			describe: "namespace in white list",
			objs:     []runtime.Object{ns, emptyRq},
			pvc:      pvc,
			wlClass:  []string{""},
			wlSpace:  []string{ns.Name},
			isError:  false,
		},
	}
	for i := range testCaseList {
		testFunc(&testCaseList[i])
	}
}
