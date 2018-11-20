package helper

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

func ListDataVolume(namespace, storageClass, name string, kc kubernetes.Interface) (
	re []*corev1.PersistentVolumeClaim, e error) {
	vList, e := kc.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	lowName := strings.ToLower(name)
	isNameOk := func(pvc *corev1.PersistentVolumeClaim) bool {
		return strings.Contains(strings.ToLower(pvc.Name), lowName)
	}
	isStorageClassOk := func(pvc *corev1.PersistentVolumeClaim) bool {
		return util.GetPVCClass(pvc) == storageClass
	}
	switch {
	case len(storageClass) == 0 && len(name) == 0:
		re = make([]*corev1.PersistentVolumeClaim, 0, len(vList.Items))
		for i := range vList.Items {
			re = append(re, &vList.Items[i])
		}
	case len(storageClass) > 0 && len(name) == 0:
		for i := range vList.Items {
			if isStorageClassOk(&vList.Items[i]) {
				re = append(re, &vList.Items[i])
			}
		}
	case len(storageClass) == 0 && len(name) > 0:
		for i := range vList.Items {
			if isNameOk(&vList.Items[i]) {
				re = append(re, &vList.Items[i])
			}
		}
	case len(storageClass) > 0 && len(name) > 0:
		for i := range vList.Items {
			if isStorageClassOk(&vList.Items[i]) && isNameOk(&vList.Items[i]) {
				re = append(re, &vList.Items[i])
			}
		}
	}
	return re, nil
}

func CreateDataVolume(volume *corev1.PersistentVolumeClaim, kc kubernetes.Interface) (re *corev1.PersistentVolumeClaim,
	fe *errors.FormatError) {
	re, e := kc.CoreV1().PersistentVolumeClaims(volume.Namespace).Create(volume)
	if e != nil {
		if kubernetes.IsAlreadyExists(e) {
			return nil, errors.NewError().SetErrorObjectAlreadyExist(volume.Name, e)
		}
		ae := errors.ParseQuotaNotCompleteApiErr(e)
		if ae != nil {
			return nil, errors.NewError().SetErrorQuotaNotCompleteFromApi(ae)
		}
		if errors.IsQuotaExceeded(e) {
			return nil, errors.NewError().SetErrorQuotaExceeded(e)
		}
		if kubernetes.IsInvalid(e) {
			return nil, errors.NewError().SetErrorObjectBadName(volume.Name, e)
		}
		return nil, errors.NewError().SetErrorInternalServerError(e)
	}
	return re, nil
}

func GetDataVolume(namespace, name string, kc kubernetes.Interface) (re *corev1.PersistentVolumeClaim,
	fe *errors.FormatError) {
	re, e := kc.CoreV1().PersistentVolumeClaims(namespace).Get(name, metav1.GetOptions{})
	if e != nil {
		if kubernetes.IsNotFound(e) {
			return nil, errors.NewError().SetErrorObjectNotFound(name, e)
		}
		return nil, errors.NewError().SetErrorInternalServerError(e)
	}
	return re, nil
}

func DeleteDataVolume(namespace, name string, kc kubernetes.Interface) *errors.FormatError {
	e := kc.CoreV1().PersistentVolumeClaims(namespace).Delete(name, nil)
	if kubernetes.IsNotFound(e) {
		return errors.NewError().SetErrorObjectNotFound(name, e)
	} else if e != nil {
		return errors.NewError().SetErrorInternalServerError(e)
	}
	return nil
}
