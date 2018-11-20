package testtools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	tntv1a1 "github.com/caicloud/clientset/pkg/apis/tenant/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
	"github.com/caicloud/storage-admin/pkg/util"
)

const (
	TypeNameDefault = "type-test"
	TypeNameOther   = "type-test-other"

	ServiceNameDefault = "service-test"
	ServiceNameOther   = "service-test-other"

	ServiceAliasDefault = "service-test-alias"
	ServiceAliasOther   = "service-test-alias-other"

	ServiceDescriptionDefault = "service-test-description"
	ServiceDescriptionOther   = "service-test-description-other"

	ClassNameDefault = "class-test"
	ClassNameOther   = "class-test-other"

	ClassAliasDefault = "class-test-alias"
	ClassAliasOther   = "class-test-alias-other"

	ClassDescriptionDefault = "class-test-description"
	ClassDescriptionOther   = "class-test-description-other"

	ClusterNameDefault = "cluster-test"

	TenantNameDefault = "tenant-test"

	NamespaceNameDefault = "namespace-test"
	NamespaceNameOther   = "namespace-test-other"

	VolumeNameDefault = "vol-test"
	VolumeNameOther   = "vol-test-other"

	MapServiceDefaultKey   = "mapServiceDefaultKey"
	MapServiceDefaultValue = "mapServiceDefaultValue"
	MapServiceDelKey       = "mapServiceDelKey"
	MapServiceDelValue     = "mapServiceDelValue"
	MapClassDefaultKey     = "mapClassDefaultKey"
	MapClassDefaultValue   = "mapClassDefaultValue"
	MapClassDelKey         = "mapClassDelKey"
	MapClassDelValue       = "mapClassDelValue"
	MapAddKey              = "mapAddKey"
	MapAddValue            = "mapAddValue"
	MapSetValue            = "mapSetValue"

	DefaultGfsRestUser        = "uuu"
	DefaultGfsRestUserKey     = "kkk"
	DefaultGfsSecretName      = "nnn"
	DefaultGfsSecretNamespace = "sss"

	QuotaDefaultResNumName  = "testQuotaDefaultResNumName"
	QuotaDefaultResSizeName = "testQuotaDefaultResSizeName"
	QuotaDefaultResNum      = 666
	QuotaDefaultResSize     = 2333
	RatioDefault            = 150

	ProvisionerDefault = kubernetes.StorageClassProvisionerGlusterfs

	VolumeSizeDefault = 30
)

// storage type

type StorageTypeBuilder resv1b1.StorageType

func (b *StorageTypeBuilder) Get() *resv1b1.StorageType {
	return ((*resv1b1.StorageType)(b)).DeepCopy()
}

func DefaultStorageTypeBuilder() *StorageTypeBuilder {
	return &StorageTypeBuilder{
		ObjectMeta: metav1.ObjectMeta{
			Name: TypeNameDefault,
		},
		Provisioner:        ProvisionerDefault,
		RequiredParameters: GetMapServiceDefault(),
		OptionalParameters: GetMapClassDefault(),
	}
}
func (b *StorageTypeBuilder) SetName(name string) *StorageTypeBuilder {
	b.Name = name
	return b
}
func (b *StorageTypeBuilder) ChangeName() *StorageTypeBuilder { return b.SetName(TypeNameOther) }
func (b *StorageTypeBuilder) SetRequiredParam(k, v string, isDel bool) *StorageTypeBuilder {
	setMap(b.RequiredParameters, k, v, isDel)
	return b
}

func (b *StorageTypeBuilder) AppendGfsParam() *StorageTypeBuilder {
	setMap(b.RequiredParameters, kubernetes.StorageClassParamNameRestUser, DefaultGfsRestUser, false)
	setMap(b.RequiredParameters, kubernetes.StorageClassParamNameRestUserKey, DefaultGfsRestUserKey, false)
	setMap(b.RequiredParameters, kubernetes.StorageClassParamNameSecretName, DefaultGfsSecretName, false)
	setMap(b.RequiredParameters, kubernetes.StorageClassParamNameSecretNamespace, DefaultGfsSecretNamespace, false)
	return b
}

// storage service

type StorageServiceBuilder resv1b1.StorageService

func (b *StorageServiceBuilder) Get() *resv1b1.StorageService {
	return ((*resv1b1.StorageService)(b)).DeepCopy()
}

func DefaultStorageServiceBuilder() *StorageServiceBuilder {
	b := &StorageServiceBuilder{
		ObjectMeta: metav1.ObjectMeta{
			Name: ServiceNameDefault,
		},
		TypeName:   TypeNameDefault,
		Parameters: GetMapServiceDefault(),
	}
	b.SetAlias(ServiceAliasDefault)
	b.SetDescription(ServiceDescriptionDefault)
	return b
}

func (b *StorageServiceBuilder) SetName(name string) *StorageServiceBuilder {
	b.Name = name
	return b
}
func (b *StorageServiceBuilder) ChangeName() *StorageServiceBuilder {
	return b.SetName(ServiceNameOther)
}

func (b *StorageServiceBuilder) SetParam(k, v string, isDel bool) *StorageServiceBuilder {
	setMap(b.Parameters, k, v, isDel)
	return b
}
func (b *StorageServiceBuilder) AppendGfsParam() *StorageServiceBuilder {
	// setMap(b.Parameters, kubernetes.StorageClassParamNameRestUser, DefaultGfsRestUser, false)
	// setMap(b.Parameters, kubernetes.StorageClassParamNameRestUserKey, DefaultGfsRestUserKey, false)
	setMap(b.Parameters, kubernetes.StorageClassParamNameSecretName, DefaultGfsSecretName, false)
	setMap(b.Parameters, kubernetes.StorageClassParamNameSecretNamespace, DefaultGfsSecretNamespace, false)
	return b
}

func (b *StorageServiceBuilder) SetType(typeName string) *StorageServiceBuilder {
	b.TypeName = typeName
	return b
}
func (b *StorageServiceBuilder) ChangeType() *StorageServiceBuilder { return b.SetType(TypeNameOther) }

func (b *StorageServiceBuilder) SetAlias(alias string) *StorageServiceBuilder {
	util.SetObjectAlias((*resv1b1.StorageService)(b), alias)
	return b
}
func (b *StorageServiceBuilder) ChangeAlias() *StorageServiceBuilder {
	return b.SetAlias(ServiceAliasOther)
}
func (b *StorageServiceBuilder) SetDescription(description string) *StorageServiceBuilder {
	util.SetObjectDescription((*resv1b1.StorageService)(b), description)
	return b
}

// storage class

type StorageClassBuilder storagev1.StorageClass

func (b *StorageClassBuilder) Get() *storagev1.StorageClass {
	return ((*storagev1.StorageClass)(b)).DeepCopy()
}

func DefaultStorageClassBuilder() *StorageClassBuilder {
	b := &StorageClassBuilder{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ClassNameDefault,
			Annotations: make(map[string]string, 5),
			Finalizers:  []string{constants.StorageClassControllerFinalizerName},
		},
		Provisioner: ProvisionerDefault,
		Parameters:  GetMapServiceDefault(),
	}
	b.SetClassMark()
	b.SetService(ServiceNameDefault)
	b.SetType(TypeNameDefault)
	b.SetAlias(ClassAliasDefault)
	b.SetDescription(ClassDescriptionDefault)
	return b
}

func (b *StorageClassBuilder) SetName(name string) *StorageClassBuilder {
	b.Name = name
	return b
}
func (b *StorageClassBuilder) ChangeName() *StorageClassBuilder {
	return b.SetName(ServiceNameOther)
}

func (b *StorageClassBuilder) SetParam(k, v string, isDel bool) *StorageClassBuilder {
	setMap(b.Parameters, k, v, isDel)
	return b
}

func (b *StorageClassBuilder) AppendGfsParam() *StorageClassBuilder {
	setMap(b.Parameters, kubernetes.StorageClassParamNameRestUser, DefaultGfsRestUser, false)
	setMap(b.Parameters, kubernetes.StorageClassParamNameRestUserKey, DefaultGfsRestUserKey, false)
	setMap(b.Parameters, kubernetes.StorageClassParamNameSecretName, DefaultGfsSecretName, false)
	setMap(b.Parameters, kubernetes.StorageClassParamNameSecretNamespace, DefaultGfsSecretNamespace, false)
	return b
}

func (b *StorageClassBuilder) SetType(typeName string) *StorageClassBuilder {
	util.SetClassType((*storagev1.StorageClass)(b), typeName)
	return b
}
func (b *StorageClassBuilder) ChangeType() *StorageClassBuilder { return b.SetType(TypeNameOther) }

func (b *StorageClassBuilder) SetService(typeName string) *StorageClassBuilder {
	util.SetClassService((*storagev1.StorageClass)(b), typeName)
	return b
}
func (b *StorageClassBuilder) ChangeService() *StorageClassBuilder {
	return b.SetService(ServiceNameOther)
}

func (b *StorageClassBuilder) SetAlias(alias string) *StorageClassBuilder {
	util.SetObjectAlias((*storagev1.StorageClass)(b), alias)
	return b
}
func (b *StorageClassBuilder) ChangeAlias() *StorageClassBuilder {
	return b.SetAlias(ClassAliasOther)
}
func (b *StorageClassBuilder) SetDescription(description string) *StorageClassBuilder {
	util.SetObjectDescription((*storagev1.StorageClass)(b), description)
	return b
}

func (b *StorageClassBuilder) SetClassMark() *StorageClassBuilder {
	util.SetObjStorageAdminMark((*storagev1.StorageClass)(b))
	return b
}

func (b *StorageClassBuilder) SetTerminated() *StorageClassBuilder {
	b.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	return b
}

// volume

type PvcBuilder corev1.PersistentVolumeClaim

func (b *PvcBuilder) Get() *corev1.PersistentVolumeClaim {
	return ((*corev1.PersistentVolumeClaim)(b)).DeepCopy()
}

func DefaultPvcBuilder() *PvcBuilder {
	resourceList := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", VolumeSizeDefault)),
	}
	b := &PvcBuilder{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VolumeNameDefault,
			Namespace: NamespaceNameDefault,
			Annotations: map[string]string{
				kubernetes.LabelKeyKubeStorageProvisioner: ProvisionerDefault,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: GetAccessModeVolumeDefault(),
			Resources: corev1.ResourceRequirements{
				Requests: resourceList,
				Limits:   resourceList,
			},
		},
	}
	b.SetClass(ClassNameDefault)
	return b
}

func (b *PvcBuilder) SetName(name string) *PvcBuilder {
	b.Name = name
	return b
}
func (b *PvcBuilder) ChangeName() *PvcBuilder {
	return b.SetName(VolumeNameOther)
}

func (b *PvcBuilder) SetNamespace(namespace string) *PvcBuilder {
	b.Namespace = namespace
	return b
}
func (b *PvcBuilder) ChangeNamespace() *PvcBuilder {
	return b.SetNamespace(NamespaceNameOther)
}

func (b *PvcBuilder) SetClass(scName string) *PvcBuilder {
	util.SetPVCClass((*corev1.PersistentVolumeClaim)(b), scName)
	return b
}
func (b *PvcBuilder) ChangeClass() *PvcBuilder {
	return b.SetClass(ClassNameOther)
}

// namespace

type NamespaceBuilder corev1.Namespace

func (b *NamespaceBuilder) Get() *corev1.Namespace {
	return ((*corev1.Namespace)(b)).DeepCopy()
}

func DefaultNamespaceBuilder() *NamespaceBuilder {
	return &NamespaceBuilder{
		ObjectMeta: metav1.ObjectMeta{
			Name: NamespaceNameDefault,
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}
}

func (b *NamespaceBuilder) SetName(name string) *NamespaceBuilder {
	b.Name = name
	return b
}
func (b *NamespaceBuilder) ChangeName() *NamespaceBuilder {
	return b.SetName(NamespaceNameOther)
}

// storage type

func GetMapServiceDefault() map[string]string {
	return map[string]string{
		MapServiceDefaultKey: MapServiceDefaultValue,
		MapServiceDelKey:     MapServiceDelValue,
	}
}

func GetMapClassDefault() map[string]string {
	return map[string]string{
		MapClassDefaultKey: MapClassDefaultValue,
		MapClassDelKey:     MapClassDelValue,
	}
}

// storage class

func ClassRemoveMark(obj *storagev1.StorageClass) *storagev1.StorageClass {
	delete(obj.Annotations, constants.LabelKeyStorageAdminMarkKey)
	return obj
}

// volume

func GetAccessModeVolumeDefault() []corev1.PersistentVolumeAccessMode {
	return []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
}

// quota

func ObjectMakeQuota(name, namespace string, resNum, resSize int64) *corev1.ResourceQuota {
	return &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				QuotaDefaultResNumName:  *resource.NewQuantity(resNum, resource.DecimalExponent),
				QuotaDefaultResSizeName: *resource.NewQuantity(resSize, resource.BinarySI),
			},
		},
	}
}

func ObjectDefaultQuota() *corev1.ResourceQuota {
	return ObjectMakeQuota(namespaceQuotaName(NamespaceNameDefault), NamespaceNameDefault,
		QuotaDefaultResNum, QuotaDefaultResSize)
}

func GetQuantityResNumDefault() resource.Quantity {
	return *resource.NewQuantity(QuotaDefaultResNum, resource.DecimalExponent)
}
func GetQuantityResSizeDefault() resource.Quantity {
	return *resource.NewQuantity(QuotaDefaultResSize, resource.BinarySI)
}

func ResListSetClassQuota(className string, objs ...corev1.ResourceList) {
	for i := range objs {
		objs[i][corev1.ResourceName(util.LabelKeyStorageQuotaNum(className))] = GetQuantityResNumDefault()
		objs[i][corev1.ResourceName(util.LabelKeyStorageQuotaSize(className))] = GetQuantityResSizeDefault()
	}
}
func CheckClassInResList(className string, objs ...corev1.ResourceList) bool {
	if objs == nil {
		return false
	}
	labelNameNum := corev1.ResourceName(util.LabelKeyStorageQuotaNum(className))
	labelNameSize := corev1.ResourceName(util.LabelKeyStorageQuotaSize(className))
	for i := range objs {
		if _, ok := objs[i][labelNameNum]; !ok {
			if _, ok = objs[i][labelNameSize]; ok {
				return true
			}
		}
	}
	return false
}

func QuotaSetNamespace(obj *corev1.ResourceQuota, namespace string) *corev1.ResourceQuota {
	obj.Namespace = namespace
	return obj
}
func QuotaSetChangeNamespace(obj *corev1.ResourceQuota) *corev1.ResourceQuota {
	return QuotaSetNamespace(obj, NamespaceNameOther)
}

// partition

func CloneResourceList(old corev1.ResourceList) corev1.ResourceList {
	nHard := make(map[corev1.ResourceName]resource.Quantity, len(old))
	for k, v := range old {
		nHard[k] = v
	}
	return nHard
}

func ObjectMakePartition(name string, hard corev1.ResourceList) *tntv1a1.Partition {
	return &tntv1a1.Partition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: tntv1a1.PartitionSpec{
			Tenant: name,
			Quota:  CloneResourceList(hard),
		},
		Status: tntv1a1.PartitionStatus{
			Phase:      tntv1a1.PartitionActive,
			Conditions: []tntv1a1.PartitionCondition{},
			Hard:       CloneResourceList(hard),
			Used:       CloneResourceList(hard),
		},
	}
}

func ObjectDefaultPartition() *tntv1a1.Partition {
	return ObjectMakePartition(NamespaceNameDefault, ObjectDefaultQuota().Spec.Hard)
}

func PartitionSetClassQuota(obj *tntv1a1.Partition, className string) *tntv1a1.Partition {
	ResListSetClassQuota(className, obj.Spec.Quota, obj.Status.Hard, obj.Status.Used)
	return obj
}

// cluster quota

func GetRatioDefault() map[corev1.ResourceName]int64 {
	return map[corev1.ResourceName]int64{
		QuotaDefaultResNumName:  RatioDefault,
		QuotaDefaultResSizeName: RatioDefault,
	}
}
func RatioSetClassQuota(className string, objs ...map[corev1.ResourceName]int64) {
	for i := range objs {
		objs[i][corev1.ResourceName(util.LabelKeyStorageQuotaNum(className))] = RatioDefault
		objs[i][corev1.ResourceName(util.LabelKeyStorageQuotaSize(className))] = RatioDefault
	}
}
func CheckClassInRatio(className string, objs ...map[corev1.ResourceName]int64) bool {
	if objs == nil {
		return false
	}
	labelNameNum := corev1.ResourceName(util.LabelKeyStorageQuotaNum(className))
	labelNameSize := corev1.ResourceName(util.LabelKeyStorageQuotaSize(className))
	for i := range objs {
		if _, ok := objs[i][labelNameNum]; !ok {
			if _, ok = objs[i][labelNameSize]; ok {
				return true
			}
		}
	}
	return false
}

func ObjectMakeClusterQuota(name string, hard corev1.ResourceList, ratio map[corev1.ResourceName]int64) *tntv1a1.ClusterQuota {
	return &tntv1a1.ClusterQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: tntv1a1.ClusterQuotaSpec{
			Ratio: ratio,
		},
		Status: tntv1a1.ClusterQuotaStatus{
			Logical: tntv1a1.Logical{
				Total:      CloneResourceList(hard),
				Allocated:  CloneResourceList(hard),
				SystemUsed: CloneResourceList(hard),
				Used:       CloneResourceList(hard),
			},
			Physical: tntv1a1.Physical{
				Capacity:    CloneResourceList(hard),
				Allocatable: CloneResourceList(hard),
				Unavailable: CloneResourceList(hard),
			},
		},
	}
}

func ObjectDefaultClusterQuota() *tntv1a1.ClusterQuota {
	return ObjectMakeClusterQuota(ClusterNameDefault, ObjectDefaultQuota().Spec.Hard, GetRatioDefault())
}

func ClusterQuotaSetClassQuota(obj *tntv1a1.ClusterQuota, className string) *tntv1a1.ClusterQuota {
	RatioSetClassQuota(className, obj.Spec.Ratio)
	return obj
}
func ClusterQuotaSetClassRatio(obj *tntv1a1.ClusterQuota, className string) *tntv1a1.ClusterQuota {
	RatioSetClassQuota(className, obj.Spec.Ratio)
	return obj
}

// tenant

func ObjectMakeTenant(name string, hard corev1.ResourceList) *tntv1a1.Tenant {
	return &tntv1a1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: tntv1a1.TenantSpec{
			Quota: CloneResourceList(hard),
		},
		Status: tntv1a1.TenantStatus{
			Phase:      tntv1a1.TenantActive,
			Hard:       CloneResourceList(hard),
			Used:       CloneResourceList(hard),
			ActualUsed: CloneResourceList(hard),
		},
	}
}

func ObjectDefaultTenant() *tntv1a1.Tenant {
	return ObjectMakeTenant(TenantNameDefault, ObjectDefaultQuota().Spec.Hard)
}

func TenantSetClassQuota(obj *tntv1a1.Tenant, className string) *tntv1a1.Tenant {
	ResListSetClassQuota(className, obj.Spec.Quota, obj.Status.Hard, obj.Status.Used, obj.Status.ActualUsed)
	return obj
}

// secret

type SecretBuilder corev1.Secret

func DefaultSecretBuilder() *SecretBuilder {
	b := &SecretBuilder{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultGfsSecretName,
			Namespace: DefaultGfsSecretNamespace,
			Labels: map[string]string{
				constants.LabelKeyStorageService: ServiceNameDefault,
			},
		},
		Data: map[string][]byte{
			"key": []byte(DefaultGfsRestUser + ":" + DefaultGfsRestUserKey),
		},
		Type: corev1.SecretType(kubernetes.StorageClassProvisionerGlusterfs),
	}
	util.SetObjStorageAdminMark(b)
	return b
}

func (b *SecretBuilder) SetGfsKey(user, key string) *SecretBuilder {
	b.Data["key"] = []byte(user + ":" + key)
	return b
}

func (b *SecretBuilder) SetKey(key string) *SecretBuilder {
	b.Data["key"] = []byte(key)
	return b
}

func (b *SecretBuilder) SetType(typeName string) *SecretBuilder {
	b.Type = corev1.SecretType(typeName)
	return b
}

func (b *SecretBuilder) DelLabel(key string) *SecretBuilder {
	delete(b.Labels, key)
	return b
}
func (b *SecretBuilder) SetLabel(key, value string) *SecretBuilder {
	b.Labels[key] = value
	return b
}

func (b *SecretBuilder) Get() *corev1.Secret {
	return (*corev1.Secret)(b)
}

func CheckServiceSecret(kc kubernetes.Interface, name, user, pwd string) error {
	ss, e := kc.ResourceV1beta1().StorageServices().Get(name, metav1.GetOptions{})
	if e != nil {
		return e
	}
	sNamespace, sName, ok := util.GetStorageSecretPath(ss.Parameters)
	if !ok {
		return fmt.Errorf("parse secret from storage service failed (%s/%s)", sNamespace, sName)
	}
	saved, e := kc.CoreV1().Secrets(sNamespace).Get(sName, metav1.GetOptions{})
	if e != nil {
		return e
	}
	userKeys := strings.SplitN(string(saved.Data["key"]), ":", 2)
	if len(userKeys) != 2 {
		return fmt.Errorf("bad secret format: %v", string(saved.Data["key"]))
	}
	if userKeys[0] != user {
		return fmt.Errorf("secret user not right, want %s got %s", user, userKeys[0])
	}
	if userKeys[1] != pwd {
		return fmt.Errorf("secret pwd not right, want %s got %s", pwd, userKeys[1])
	}
	return nil
}

func CheckClassSecret(kc kubernetes.Interface, name string, secret *corev1.Secret) error {
	sc, e := kc.StorageV1().StorageClasses().Get(name, metav1.GetOptions{})
	if e != nil {
		return e
	}
	sNamespace, sName, ok := util.GetStorageSecretPath(sc.Parameters)
	switch {
	case secret == nil && ok:
		return fmt.Errorf("should not have secret %s/%s", sNamespace, sName)
	case secret != nil && !ok:
		return fmt.Errorf("parse secret from storage service failed (%s/%s)", sNamespace, sName)
	}
	saved, e := kc.CoreV1().Secrets(sNamespace).Get(sName, metav1.GetOptions{})
	if e != nil {
		return e
	}
	if bytes.Compare(saved.Data["key"], secret.Data["key"]) != 0 {
		return fmt.Errorf("secret not same, want %s got %s", string(secret.Data["key"]), string(saved.Data["key"]))
	}
	return nil
}

// service create request

type ServiceCRBuilder apiv1a1.CreateStorageServiceRequest

func NewDefaultServiceCRBuilder() *ServiceCRBuilder {
	return &ServiceCRBuilder{
		Name:        ServiceNameDefault,
		Alias:       ServiceAliasDefault,
		Description: ServiceDescriptionDefault,
		Type:        TypeNameDefault,
		Parameters:  GetMapServiceDefault(),
	}
}

func (b *ServiceCRBuilder) Get() *apiv1a1.CreateStorageServiceRequest {
	return (*apiv1a1.CreateStorageServiceRequest)(b)
}

func (b *ServiceCRBuilder) SetName(name string) *ServiceCRBuilder {
	b.Name = name
	return b
}
func (b *ServiceCRBuilder) SetAlias(alias string) *ServiceCRBuilder {
	b.Alias = alias
	return b
}
func (b *ServiceCRBuilder) SetDescription(description string) *ServiceCRBuilder {
	b.Description = description
	return b
}
func (b *ServiceCRBuilder) SetType(tpName string) *ServiceCRBuilder {
	b.Type = tpName
	return b
}
func (b *ServiceCRBuilder) DelKey(k string) *ServiceCRBuilder {
	setMap(b.Parameters, k, "", true)
	return b
}
func (b *ServiceCRBuilder) SetKeyValue(k, v string) *ServiceCRBuilder {
	setMap(b.Parameters, k, v, false)
	return b
}
func (b *ServiceCRBuilder) SetAddParam() *ServiceCRBuilder {
	return b.SetKeyValue(MapAddKey, MapAddValue)
}
func (b *ServiceCRBuilder) SetDelParam() *ServiceCRBuilder     { return b.DelKey(MapServiceDelKey) }
func (b *ServiceCRBuilder) SetReplaceParam() *ServiceCRBuilder { return b.SetAddParam().SetDelParam() }

func (b *ServiceCRBuilder) ServiceCRAppendGfsUserKey() *ServiceCRBuilder {
	setMap(b.Parameters, kubernetes.StorageClassParamNameRestUser, "", false)
	setMap(b.Parameters, kubernetes.StorageClassParamNameRestUserKey, "", false)
	setMap(b.Parameters, kubernetes.StorageClassParamNameSecretName, "", false)
	setMap(b.Parameters, kubernetes.StorageClassParamNameSecretNamespace, "", false)
	return b
}

func ServiceCRSetAlias(obj *apiv1a1.CreateStorageServiceRequest, alias string) *apiv1a1.CreateStorageServiceRequest {
	obj.Alias = alias
	return obj
}

func ServiceCRSetDescription(obj *apiv1a1.CreateStorageServiceRequest, description string) *apiv1a1.CreateStorageServiceRequest {
	obj.Description = description
	return obj
}

// service update request

func ObjectMakeStorageServiceUR(alias, description string) *apiv1a1.UpdateStorageServiceRequest {
	return &apiv1a1.UpdateStorageServiceRequest{
		Alias:       alias,
		Description: description,
	}
}

func ObjectDefaultStorageServiceUR() *apiv1a1.UpdateStorageServiceRequest {
	return ObjectMakeStorageServiceUR(ServiceAliasDefault, ServiceDescriptionDefault)
}

// class create request

type ClassCRBuilder apiv1a1.CreateStorageClassRequest

func NewDefaultClassCRBuilder() *ClassCRBuilder {
	return &ClassCRBuilder{
		Name:       ClassNameDefault,
		Service:    ServiceNameDefault,
		Parameters: GetMapClassDefault(),
	}
}

func ObjectMakeStorageClassCR(name, svName, alias, description string, pm map[string]string) *apiv1a1.CreateStorageClassRequest {
	return &apiv1a1.CreateStorageClassRequest{
		Name:        name,
		Alias:       alias,
		Description: description,
		Service:     svName,
		Parameters:  pm,
	}
}

func ObjectDefaultStorageClassCR() *apiv1a1.CreateStorageClassRequest {
	return ObjectMakeStorageClassCR(ClassNameDefault,
		ServiceNameDefault,
		ClassAliasDefault,
		ClassDescriptionDefault,
		GetMapClassDefault())
}

func (b *ClassCRBuilder) Get() *apiv1a1.CreateStorageClassRequest {
	return (*apiv1a1.CreateStorageClassRequest)(b)
}

func (b *ClassCRBuilder) SetName(name string) *ClassCRBuilder {
	b.Name = name
	return b
}
func (b *ClassCRBuilder) SetService(svName string) *ClassCRBuilder {
	b.Service = svName
	return b
}
func (b *ClassCRBuilder) DelKey(k string) *ClassCRBuilder {
	setMap(b.Parameters, k, "", true)
	return b
}
func (b *ClassCRBuilder) SetKeyValue(k, v string) *ClassCRBuilder {
	setMap(b.Parameters, k, v, false)
	return b
}
func (b *ClassCRBuilder) SetAddParam() *ClassCRBuilder {
	return b.SetKeyValue(MapAddKey, MapAddValue)
}
func (b *ClassCRBuilder) SetDelParam() *ClassCRBuilder     { return b.DelKey(MapClassDelKey) }
func (b *ClassCRBuilder) SetReplaceParam() *ClassCRBuilder { return b.SetAddParam().SetDelParam() }

func ClassCRSetAlias(obj *apiv1a1.CreateStorageClassRequest, alias string) *apiv1a1.CreateStorageClassRequest {
	obj.Alias = alias
	return obj
}

func ClassCRSetDescription(obj *apiv1a1.CreateStorageClassRequest, description string) *apiv1a1.CreateStorageClassRequest {
	obj.Description = description
	return obj
}

// class update request

func ObjectMakeStorageClassUR(alias, description string) *apiv1a1.UpdateStorageClassRequest {
	return &apiv1a1.UpdateStorageClassRequest{
		Alias:       alias,
		Description: description,
	}
}

func ObjectDefaultStorageClassUR() *apiv1a1.UpdateStorageClassRequest {
	return ObjectMakeStorageClassUR(ClassAliasOther, ClassDescriptionOther)
}

// volume create request

func ObjectMakeVolumeCR(name, scName string,
	accessMode []corev1.PersistentVolumeAccessMode, size int) *apiv1a1.CreateDataVolumeRequest {
	return &apiv1a1.CreateDataVolumeRequest{
		Name:         name,
		StorageClass: scName,
		AccessModes:  accessMode,
		Size:         size,
	}
}

func ObjectDefaultVolumeCR() *apiv1a1.CreateDataVolumeRequest {
	return ObjectMakeVolumeCR(VolumeNameDefault,
		ClassNameDefault,
		GetAccessModeVolumeDefault(),
		VolumeSizeDefault)
}

func VolumeCRSetName(obj *apiv1a1.CreateDataVolumeRequest, name string) *apiv1a1.CreateDataVolumeRequest {
	obj.Name = name
	return obj
}

func VolumeCRSetClass(obj *apiv1a1.CreateDataVolumeRequest, cName string) *apiv1a1.CreateDataVolumeRequest {
	obj.StorageClass = cName
	return obj
}

func VolumeCRSetAccessModes(obj *apiv1a1.CreateDataVolumeRequest, accessMode []corev1.PersistentVolumeAccessMode) *apiv1a1.CreateDataVolumeRequest {
	obj.AccessModes = accessMode
	return obj
}

func VolumeCRSetSize(obj *apiv1a1.CreateDataVolumeRequest, size int) *apiv1a1.CreateDataVolumeRequest {
	obj.Size = size
	return obj
}

// tools

type TestClientSetGetter struct {
	Client kubernetes.Interface
}

func (csg *TestClientSetGetter) KubeClient() kubernetes.Interface { return csg.Client }

func namespaceQuotaName(nsName string) string {
	// TODO maybe should ref to real one
	return nsName
}

func setMap(m map[string]string, k, v string, isDel bool) {
	if isDel {
		delete(m, k)
	} else {
		m[k] = v
	}
}

func IsKeyValOk(k, v string, m map[string]string) bool {
	return m[k] == v
}

func IsKeyNotExist(k string, m map[string]string) bool {
	_, ok := m[k]
	return !ok
}

func ToJson(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func GetClassNameList(objs []storagev1.StorageClass) []string {
	nl := make([]string, len(objs))
	for i := range objs {
		nl[i] = objs[i].GetName()
	}
	return nl
}

func IsStringsSame(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	getSorted := func(ss []string) []string {
		if sort.StringsAreSorted(ss) {
			return ss
		}
		nss := make([]string, len(ss))
		for i := range ss {
			nss[i] = ss[i]
		}
		sort.Strings(nss)
		return nss
	}
	a = getSorted(a)
	b = getSorted(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
