package helper

import (
	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	errors "github.com/caicloud/storage-admin/pkg/errors"
	kubernetes "github.com/caicloud/storage-admin/pkg/kubernetes"
)

func ListStorageType(kc kubernetes.Interface) (*resv1b1.StorageTypeList, *errors.FormatError) {
	cList, e := kc.ResourceV1beta1().StorageTypes().List(metav1.ListOptions{})
	if e != nil {
		return nil, errors.NewError().SetErrorInternalServerError(e)
	}
	return cList, nil
}

func GetStorageTypeInner(name string, kc kubernetes.Interface) (*resv1b1.StorageType, *errors.FormatError) {
	tp, e := kc.ResourceV1beta1().StorageTypes().Get(name, metav1.GetOptions{})
	if e != nil {
		if kubernetes.IsNotFound(e) {
			return nil, errors.NewError().SetErrorTypeNotFound(name)
		} else {
			return nil, errors.NewError().SetErrorInternalServerError(e)
		}
	}
	return tp, nil
}
