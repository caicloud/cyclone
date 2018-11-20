package helper

import (
	"fmt"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

func ListStorageService(kc kubernetes.Interface, typeName, name string) (aServiceList []apiv1a1.StorageServiceObject, missTypeList []string, e error) {
	var (
		sMetaList *resv1b1.StorageServiceList
		tpList    *resv1b1.StorageTypeList
	)
	sMetaList, e = kc.ResourceV1beta1().StorageServices().List(metav1.ListOptions{})
	if e != nil {
		return
	}
	tpList, e = kc.ResourceV1beta1().StorageTypes().List(metav1.ListOptions{})
	if e != nil {
		return
	}
	tpMap := make(map[string]*resv1b1.StorageType, len(tpList.Items))
	for i := range tpList.Items {
		tpMap[tpList.Items[i].Name] = &tpList.Items[i]
	}
	ss := util.StorageServiceTypeAndNameFilter(sMetaList.Items, typeName, name)
	aServiceList, missTypeList = util.DefaultTranslatorMetaToApi().StorageServiceList(ss, tpMap)
	return
}

func GetStorageService(name string, kc kubernetes.Interface) (*resv1b1.StorageService, *errors.FormatError) {
	service, e := kc.ResourceV1beta1().StorageServices().Get(name, metav1.GetOptions{})
	if e != nil {
		if kubernetes.IsNotFound(e) {
			return nil, errors.NewError().SetErrorObjectNotFound(name, e)
		} else {
			return nil, errors.NewError().SetErrorInternalServerError(e)
		}
	}
	return service, nil
}

func GetStorageServiceAndType(serviceName string, kc kubernetes.Interface) (*resv1b1.StorageService,
	*resv1b1.StorageType, *errors.FormatError) {
	service, fe := GetStorageService(serviceName, kc)
	if fe != nil {
		return nil, nil, fe
	}
	tp, fe := GetStorageTypeInner(service.TypeName, kc)
	if fe != nil {
		return nil, nil, fe
	}
	return service, tp, nil
}

func CreateStorageService(service *resv1b1.StorageService, kc kubernetes.Interface) (*resv1b1.StorageService, *errors.FormatError) {
	re, e := kc.ResourceV1beta1().StorageServices().Create(service)
	if e != nil {
		if kubernetes.IsAlreadyExists(e) {
			return nil, errors.NewError().SetErrorObjectAlreadyExist(service.Name, e)
		} else if kubernetes.IsInvalid(e) {
			return nil, errors.NewError().SetErrorObjectBadName(service.Name, e)
		} else {
			return nil, errors.NewError().SetErrorInternalServerError(e)
		}
	}
	return re, nil
}

func GetStorageSecret(pm map[string]string, kc kubernetes.Interface) (*corev1.Secret, *errors.FormatError) {
	secretNamespace, secretName, ok := util.GetStorageSecretPath(pm)
	if ok {
		secret, e := kc.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if e != nil {
			return nil, errors.NewError().SetErrorStorageSecretNotFound(secretNamespace, secretName, e)
		}
		return secret, nil
	}
	return nil, nil
}

func CreateSecret(secret *corev1.Secret, kc kubernetes.Interface) (*corev1.Secret, *errors.FormatError) {
	re, e := kc.CoreV1().Secrets(secret.GetNamespace()).Create(secret)
	if e != nil {
		if kubernetes.IsAlreadyExists(e) {
			return nil, errors.NewError().SetErrorObjectAlreadyExist(secret.Name, e)
		} else {
			return nil, errors.NewError().SetErrorInternalServerError(e)
		}
	}
	return re, nil
}

func UpdateStorageService(service *resv1b1.StorageService, kc kubernetes.Interface) (*resv1b1.StorageService, *errors.FormatError) {
	re, e := kc.ResourceV1beta1().StorageServices().Update(service)
	if e != nil {
		if kubernetes.IsNotFound(e) {
			return nil, errors.NewError().SetErrorObjectNotFound(service.Name, e)
		} else {
			return nil, errors.NewError().SetErrorInternalServerError(e)
		}
	}
	return re, nil
}

func TerminateStorageService(name string, kc kubernetes.Interface) *errors.FormatError {
	_, fe := GetStorageService(name, kc)
	if fe != nil {
		return fe
	}
	e := kc.ResourceV1beta1().StorageServices().Delete(name, nil)
	if e != nil {
		if kubernetes.IsNotFound(e) {
			return nil
		}
		return errors.NewError().SetErrorInternalServerError(e)
	}
	return nil
}

func CheckAndPreworkForGlusterFS(sName string, parameters map[string]string,
	tp *resv1b1.StorageType, kc kubernetes.Interface) *errors.FormatError {
	// common check
	fe := util.CheckStorageServiceParameters(parameters, tp.RequiredParameters)
	if fe != nil {
		return fe
	}
	// auth about
	restuser := parameters[kubernetes.StorageClassParamNameRestUser]
	restuserkey := parameters[kubernetes.StorageClassParamNameRestUserKey]
	secretName := parameters[kubernetes.StorageClassParamNameSecretName]
	secretNamespace := parameters[kubernetes.StorageClassParamNameSecretNamespace]
	switch {
	case len(restuser) > 0 && len(restuserkey) > 0 && len(secretNamespace) == 0 && len(secretName) == 0:
		// gen and create secret
		secret := util.GenGlusterfsSecret(sName, restuser, restuserkey)
		parameters[kubernetes.StorageClassParamNameSecretName] = secret.Name
		parameters[kubernetes.StorageClassParamNameSecretNamespace] = secret.Namespace
		delete(parameters, kubernetes.StorageClassParamNameRestUserKey)
		_, fe = CreateSecret(secret, kc)
		if fe != nil {
			return fe
		}
		return nil
	case len(secretNamespace) > 0 && len(secretName) > 0:
		// check exist secret
		secret, e := kc.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if e != nil {
			if kubernetes.IsNotFound(e) {
				return errors.NewError().SetErrorStorageSecretNotFound(secretNamespace, secretName, e)
			}
			return errors.NewError().SetErrorInternalServerError(e)
		}
		pUser, pKey, e := util.ParseGlusterfsSecret(secret)
		if e != nil {
			return errors.NewError().SetErrorBadRequest(e)
		}
		// compare secret and parameter
		if len(restuser) == 0 {
			parameters[kubernetes.StorageClassParamNameRestUser] = pUser
		} else if restuser != pUser {
			return errors.NewError().SetErrorBadRequest(
				fmt.Errorf("%s=%s not match", kubernetes.StorageClassParamNameRestUser, restuser))
		}
		if len(restuserkey) > 0 {
			if restuserkey != pKey {
				return errors.NewError().SetErrorBadRequest(
					fmt.Errorf("%s=%s not match", kubernetes.StorageClassParamNameRestUserKey, restuserkey))
			}
			delete(parameters, kubernetes.StorageClassParamNameRestUserKey)
		}
		return nil
	case (len(secretNamespace) > 0 && len(secretName) == 0) || (len(secretNamespace) == 0 && len(secretName) > 0):
		return errors.NewError().SetErrorBadRequest(
			fmt.Errorf("secret path '%s/%s' not complete", secretName, secretNamespace))
	case ((len(restuser) > 0 && len(restuserkey) == 0) || (len(restuser) == 0 && len(restuserkey) > 0)) &&
		len(secretNamespace) == 0 && len(secretName) == 0:
		return errors.NewError().SetErrorBadRequest(
			fmt.Errorf("rest user and key '%s:%s' not complete", restuser, restuserkey))
	default: // all empty means no auth
		return nil
	}
}
