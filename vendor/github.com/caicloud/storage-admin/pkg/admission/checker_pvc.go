package admission

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	admsv1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

var (
	pvcIgnoreNamespaces   = getWhiteListMap(DefaultIgnoredNamespaces)
	pvcIgnoreStorageClass = getWhiteListMap(DefaultIgnoredClasses)
)

func admissionReviewCheckerPvc(kc kubernetes.Interface, ar *admsv1.AdmissionReview) error {
	logPath := fmt.Sprintf("admissionReviewCheckerPvc %v", admissionReviewToString(ar))
	// check and get pvc
	pvc, e := getPvcFromAr(ar)
	if e != nil {
		glog.Errorf("%s getPvcFromAr failed, %v", logPath, e)
		return errors.NewError().SetErrorQuotaNotCompleteByError(e).Api()
	}

	// check ignores
	namespace := pvc.Namespace
	if pvcIgnoreNamespaces[namespace] == true {
		glog.Infof("%s in ignored namespace %s", logPath, namespace)
		return nil
	}
	storageClass := util.GetPVCClass(pvc)
	if pvcIgnoreStorageClass[storageClass] == true {
		// TODO check default class instead, but seems there is no default class set
		glog.Infof("%s in ignored storage class %s", logPath, storageClass)
		return nil
	}

	// get quota
	quotaName := util.NamespaceQuotaName(namespace)
	rq, e := kc.CoreV1().ResourceQuotas(namespace).Get(quotaName, metav1.GetOptions{})
	if e != nil {
		glog.Errorf("%s ResourceQuotas %s/%s Get failed, %v", logPath, namespace, quotaName, e)
		return errors.NewError().SetErrorQuotaNotComplete(storageClass, namespace).Api()
	}

	// check
	if !hasStorageClassResourceQuota(rq, storageClass) {
		return errors.NewError().SetErrorQuotaNotComplete(storageClass, namespace).Api()
	}
	return nil
}

func getPvcFromAr(ar *admsv1.AdmissionReview) (pvc *corev1.PersistentVolumeClaim, e error) {
	pvc = new(corev1.PersistentVolumeClaim)
	e = json.Unmarshal(ar.Request.Object.Raw, &pvc)
	if e != nil {
		return nil, e
	}
	if len(pvc.Namespace) == 0 {
		pvc.Namespace = ar.Request.Namespace
	}
	return pvc, nil
}

func hasStorageClassResourceQuota(rq *corev1.ResourceQuota, scName string) bool {
	quotaNames := []string{
		util.LabelKeyStorageQuotaSize(scName),
		util.LabelKeyStorageQuotaNum(scName),
	}

	for i := range quotaNames {
		quotaName := corev1.ResourceName(quotaNames[i])
		if _, ok := rq.Spec.Hard[quotaName]; !ok {
			return false
		}
	}
	return true
}

func getWhiteListMap(envVal string) map[string]bool {
	ss := strings.Split(envVal, ",")
	m := make(map[string]bool, len(ss))
	for _, k := range ss {
		m[k] = true
	}
	return m
}

func SetPvcIgnoreNamespaces(envVal string) {
	// lock not needed now
	pvcIgnoreNamespaces = getWhiteListMap(envVal)
}
func SetPvcIgnoreClasses(envVal string) {
	// lock not needed now
	pvcIgnoreStorageClass = getWhiteListMap(envVal)
}
