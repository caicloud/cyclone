package common

import (
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

func CheckAndDeleteGlusterfsSecret(kc kubernetes.Interface, namespace, name, logPrefix string) error {
	secret, e := kc.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
	if e != nil {
		if kubernetes.IsNotFound(e) {
			return nil
		}
		return e
	}
	if !util.IsObjStorageAdminMarked(secret) {
		glog.Warningf("%s secret %s/%s is not created by storage admin, leave it alone", logPrefix, namespace, name)
		return nil
	}
	e = kc.CoreV1().Secrets(namespace).Delete(name, nil)
	if e != nil && !kubernetes.IsNotFound(e) {
		return e
	}
	return nil
}
