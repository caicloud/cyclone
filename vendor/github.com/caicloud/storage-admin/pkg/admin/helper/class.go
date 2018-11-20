package helper

import (
	"fmt"
	"strconv"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/storage-admin/pkg/errors"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

func ListStorageClass(kc kubernetes.Interface, typeName, name string) ([]storagev1.StorageClass, error) {
	cList, e := kc.StorageV1().StorageClasses().List(metav1.ListOptions{})
	if e != nil {
		return nil, e
	}
	return util.StorageClassTypeAndNameFilter(cList.Items, typeName, name), nil
}

func CreateStorageClass(class *storagev1.StorageClass, kc kubernetes.Interface) (*storagev1.StorageClass, *errors.FormatError) {
	re, e := kc.StorageV1().StorageClasses().Create(class)
	if e != nil {
		if kubernetes.IsAlreadyExists(e) {
			return nil, errors.NewError().SetErrorObjectAlreadyExist(class.Name, e)
		} else if kubernetes.IsInvalid(e) {
			return nil, errors.NewError().SetErrorObjectBadName(class.Name, e)
		} else {
			return nil, errors.NewError().SetErrorInternalServerError(e)
		}
	}
	return re, nil
}

func GetStorageClass(name string, kc kubernetes.Interface) (*storagev1.StorageClass, *errors.FormatError) {
	re, e := kc.StorageV1().StorageClasses().Get(name, metav1.GetOptions{})
	if e != nil {
		if kubernetes.IsNotFound(e) {
			return nil, errors.NewError().SetErrorObjectNotFound(name, e)
		} else {
			return nil, errors.NewError().SetErrorInternalServerError(e)
		}
	}
	return re, nil
}

func GetStorageClassWithStatusCheck(name string, kc kubernetes.Interface) (*storagev1.StorageClass, *errors.FormatError) {
	re, fe := GetStorageClass(name, kc)
	if fe != nil {
		return nil, fe
	}
	if re.DeletionTimestamp != nil {
		return nil, errors.NewError().SetErrorClassStatusNotActive(
			fmt.Sprintf("marked deleted in %v", re.DeletionTimestamp.String()))
	}
	return re, nil
}

func GetStorageServiceAndTypeWithOptParamMapCheck(serviceName string, pm map[string]string, kc kubernetes.Interface) (*resv1b1.StorageService,
	*resv1b1.StorageType, *errors.FormatError) {
	service, tp, fe := GetStorageServiceAndType(serviceName, kc)
	if fe != nil {
		return nil, nil, fe
	}
	if fe = util.CheckStorageClassParameters(pm, tp.OptionalParameters); fe != nil {
		return nil, nil, fe
	}
	return service, tp, nil
}

func UpdateStorageClass(class *storagev1.StorageClass, kc kubernetes.Interface) (*storagev1.StorageClass, *errors.FormatError) {
	re, e := kc.StorageV1().StorageClasses().Update(class)
	if e != nil {
		if kubernetes.IsNotFound(e) {
			return nil, errors.NewError().SetErrorObjectNotFound(class.Name, e)
		} else {
			return nil, errors.NewError().SetErrorInternalServerError(e)
		}
	}
	return re, nil
}

func TerminateStorageClass(name string, kc kubernetes.Interface) *errors.FormatError {
	class, fe := GetStorageClass(name, kc)
	if fe != nil {
		return fe
	}
	if util.IsObjStorageAdminMarked(class) && !util.HasClassFinalizer(class.Finalizers) && class.DeletionTimestamp == nil {
		class.Finalizers = util.AddClassFinalizer(class.Finalizers)
		class, fe = UpdateStorageClass(class, kc)
		if fe != nil && !errors.IsObjectNotFound(fe) {
			return fe
		}
	}
	e := kc.StorageV1().StorageClasses().Delete(name, nil)
	if e != nil && !kubernetes.IsNotFound(e) {
		return errors.NewError().SetErrorInternalServerError(e)
	}

	return nil
}

func CreateGlusterFsPrework(sc *storagev1.StorageClass, ss *resv1b1.StorageService,
	mc, kc kubernetes.Interface) (*storagev1.StorageClass, *errors.FormatError) {
	// gid
	if e := checkGlusterFsGidRange(sc.Parameters); e != nil {
		return sc, errors.NewError().SetErrorBadRequestBody(e)
	}

	// secret
	secret, fe := GetStorageSecret(ss.Parameters, mc)
	if fe != nil {
		return sc, fe
	}
	if secret == nil {
		return sc, nil
	}
	secret = secret.DeepCopy()
	secret.Name = util.Md5String(sc.Name) + "-" + util.RandomString()
	secret.ResourceVersion = ""
	sc.Parameters[kubernetes.StorageClassParamNameSecretNamespace] = secret.Namespace
	sc.Parameters[kubernetes.StorageClassParamNameSecretName] = secret.Name
	delete(sc.Parameters, kubernetes.StorageClassParamNameRestUserKey)
	if _, fe = CreateSecret(secret, kc); fe != nil {
		return sc, fe
	}
	return sc, nil
}

func checkGlusterFsGidRange(pm map[string]string) error {
	var (
		ids []int
		pns = []string{kubernetes.StorageClassParamNameGidMin, kubernetes.StorageClassParamNameGidMax}
	)
	for _, pn := range pns {
		v, ok := pm[pn]
		if ok {
			id, e := strconv.Atoi(v)
			if e != nil {
				return fmt.Errorf("%s parse failed", pn)
			}
			if id < kubernetes.StorageClassParamGidRangeMin || kubernetes.StorageClassParamGidRangeMax < id {
				return fmt.Errorf("%s=%d not in range [%d, %d]", pn, id,
					kubernetes.StorageClassParamGidRangeMin, kubernetes.StorageClassParamGidRangeMax)
			}
			ids = append(ids, id)
		}
	}
	if len(ids) == 2 && ids[0] > ids[1] {
		return fmt.Errorf("%s > %s",
			kubernetes.StorageClassParamNameGidMin, kubernetes.StorageClassParamNameGidMax)
	}
	return nil
}
