package util

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	resv1b1 "github.com/caicloud/clientset/pkg/apis/resource/v1beta1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1a1 "github.com/caicloud/storage-admin/pkg/apis/admin/v1alpha1"
	"github.com/caicloud/storage-admin/pkg/constants"
	"github.com/caicloud/storage-admin/pkg/kubernetes"
)

var (
	translatorMetaToApi MetaToApiTranslator
	translatorApiToMeta ApiToMetaTranslator

	startTime = time.Now()
	randSeq   uint64
)

func init() {
	// random about
	rand.Seed(time.Now().UnixNano())
}

type MetaToApiTranslator struct{}
type ApiToMetaTranslator struct{}

func DefaultTranslatorMetaToApi() *MetaToApiTranslator {
	return &translatorMetaToApi
}
func DefaultTranslatorApiToMeta() *ApiToMetaTranslator {
	return &translatorApiToMeta
}

func CommonMetaFromMetaToApi(meta *metav1.ObjectMeta, api *apiv1a1.ObjectMetaData) {
	api.ID = string(meta.UID)
	api.Name = meta.Name
	api.CreationTime = meta.CreationTimestamp.Format(constants.TimeFormatObjectMeta)
	// TODO 原生没有 LastUpdateTime
	api.LastUpdateTime = meta.CreationTimestamp.Format(constants.TimeFormatObjectMeta)
	api.Alias = GetObjectAlias(meta)
	api.Description = GetObjectDescription(meta)
}

func StringMapCopy(src map[string]string) map[string]string {
	re := make(map[string]string, len(src))
	for k, v := range src {
		re[k] = v
	}
	return re
}

// meta to api

// type

func (t *MetaToApiTranslator) StorageTypeSet(api *apiv1a1.StorageTypeObject, meta *resv1b1.StorageType) {
	CommonMetaFromMetaToApi(&meta.ObjectMeta, &api.Metadata)
	api.Provisioner = meta.Provisioner
	api.CommonParameters = StringMapCopy(meta.RequiredParameters)
	api.OptionalParameters = StringMapCopy(meta.OptionalParameters)
}

func (t *MetaToApiTranslator) StorageTypeList(metas []resv1b1.StorageType) []apiv1a1.StorageTypeObject {
	re := make([]apiv1a1.StorageTypeObject, len(metas))
	for i := range metas {
		t.StorageTypeSet(&re[i], &metas[i])
	}
	return re
}

func (t *MetaToApiTranslator) StorageTypeInfoSet(api *apiv1a1.StorageTypeInfoObject, meta *resv1b1.StorageType) {
	api.Name = meta.Name
	api.Provisioner = meta.Provisioner
	api.OptionalParameters = StringMapCopy(meta.OptionalParameters)
}

// service

type UniqueStringCollector struct { // only in one thread
	m map[string]struct{}
}

func NewUniqueStringCollector() *UniqueStringCollector {
	return &UniqueStringCollector{map[string]struct{}{}}
}
func (m *UniqueStringCollector) AddType(name string) { m.m[name] = struct{}{} }
func (m *UniqueStringCollector) GetList() []string {
	if len(m.m) == 0 {
		return nil
	}
	list := make([]string, 0, len(m.m))
	for k := range m.m {
		list = append(list, k)
	}
	return list
}

func (t *MetaToApiTranslator) StorageServiceSet(api *apiv1a1.StorageServiceObject,
	meta *resv1b1.StorageService, tpInfo *resv1b1.StorageType) {
	CommonMetaFromMetaToApi(&meta.ObjectMeta, &api.Metadata)
	t.StorageTypeInfoSet(&api.Type, tpInfo)
	api.Parameters = StringMapCopy(meta.Parameters)
}

func (t *MetaToApiTranslator) StorageServiceList(metas []resv1b1.StorageService,
	typeMap map[string]*resv1b1.StorageType) (re []apiv1a1.StorageServiceObject, missingTypes []string) {
	re = make([]apiv1a1.StorageServiceObject, len(metas))
	missMap := NewUniqueStringCollector()
	wi := 0
	for i := range metas {
		tp := typeMap[metas[i].TypeName]
		if tp != nil {
			t.StorageServiceSet(&re[wi], &metas[i], tp)
			wi++
		} else {
			missMap.AddType(metas[i].TypeName)
		}
	}
	re = re[:wi]
	missingTypes = missMap.GetList()
	return re, missingTypes
}

// class

func (t *MetaToApiTranslator) StorageClassSet(api *apiv1a1.StorageClassObject, meta *storagev1.StorageClass) {
	// meta.StorageClass.DeepCopyInto(&api.StorageClass)
	api.TypeMeta = meta.TypeMeta
	api.ObjectMeta = meta.ObjectMeta
	api.Provisioner = meta.Provisioner
	api.Parameters = MapCopy(meta.Parameters)
	if meta.DeletionTimestamp != nil {
		api.Status = apiv1a1.StorageClassTerminating
	} else {
		api.Status = apiv1a1.StorageClassActive
	}
}

func (t *MetaToApiTranslator) StorageClassList(metas []storagev1.StorageClass) []apiv1a1.StorageClassObject {
	re := make([]apiv1a1.StorageClassObject, len(metas))
	for i := range metas {
		t.StorageClassSet(&re[i], &metas[i])
	}
	return re
}

func (t *MetaToApiTranslator) StorageClassListActiveOnly(metas []storagev1.StorageClass) []apiv1a1.StorageClassObject {
	// TODO: just temporary, for fast fix website bugs
	re := make([]apiv1a1.StorageClassObject, len(metas))
	l := 0
	for i := range metas {
		if metas[i].DeletionTimestamp == nil {
			t.StorageClassSet(&re[l], &metas[i])
			l++
		}
	}
	return re[:l]
}

func MapCopy(src map[string]string) (dst map[string]string) {
	dst = make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// volume

func (t *MetaToApiTranslator) DataVolumeSet(api *apiv1a1.DataVolumeObject, meta *corev1.PersistentVolumeClaim) {
	meta.DeepCopyInto(api)
}

func (t *MetaToApiTranslator) DataVolumeList(metas []*corev1.PersistentVolumeClaim) []apiv1a1.DataVolumeObject {
	re := make([]apiv1a1.DataVolumeObject, len(metas))
	for i := range metas {
		t.DataVolumeSet(&re[i], metas[i])
	}
	return re
}

// name and description

func setObjectLabel(obj metav1.Object, k, v string) {
	if obj == nil {
		return
	}
	am := obj.GetAnnotations()
	if am == nil {
		am = make(map[string]string, 1)
	}
	am[k] = v
	obj.SetAnnotations(am)
}
func getObjectLabel(obj metav1.Object, k string) (v string, ok bool) {
	if obj == nil {
		return "", false
	}
	am := obj.GetAnnotations()
	if am == nil {
		return "", false
	}
	v, ok = am[k]
	return v, ok
}

func SetObjectAlias(obj metav1.Object, name string) {
	setObjectLabel(obj, constants.LabelKeyStorageAdminAlias, name)
}
func GetObjectAlias(obj metav1.Object) string {
	name, _ := getObjectLabel(obj, constants.LabelKeyStorageAdminAlias)
	return name
}

func SetObjectDescription(obj metav1.Object, name string) {
	setObjectLabel(obj, constants.LabelKeyStorageAdminDescription, name)
}
func GetObjectDescription(obj metav1.Object) string {
	name, _ := getObjectLabel(obj, constants.LabelKeyStorageAdminDescription)
	return name
}

// api to meta

func (t *ApiToMetaTranslator) StorageServiceNewFromCreate(api *apiv1a1.CreateStorageServiceRequest) *resv1b1.StorageService {
	meta := new(resv1b1.StorageService)
	meta.ObjectMeta.Name = api.Name
	meta.ObjectMeta.Finalizers = []string{constants.StorageServiceControllerFinalizerName}
	meta.ObjectMeta.Annotations = make(map[string]string, 2)
	SetObjectAlias(meta, api.Alias)
	SetObjectDescription(meta, api.Description)
	meta.TypeName = api.Type
	meta.Parameters = StringMapCopy(api.Parameters)
	return meta
}

func (t *ApiToMetaTranslator) StorageClassNewFromCreate(api *apiv1a1.CreateStorageClassRequest,
	service *resv1b1.StorageService, tp *resv1b1.StorageType) *storagev1.StorageClass {
	meta := new(storagev1.StorageClass)
	meta.ObjectMeta.Name = api.Name
	meta.ObjectMeta.Annotations = make(map[string]string, 5)
	SetClassService(meta, service.Name)
	SetObjStorageAdminMark(meta)
	SetClassType(meta, tp.Name)
	SetObjectAlias(meta, api.Alias)
	SetObjectDescription(meta, api.Description)
	meta.ObjectMeta.Finalizers = []string{constants.StorageClassControllerFinalizerName}
	meta.Provisioner = tp.Provisioner
	meta.Parameters = StringMapCopy(api.Parameters)
	for k, v := range service.Parameters {
		meta.Parameters[k] = v
	}
	return meta
}

func (t *ApiToMetaTranslator) StorageClassSetFromUpdate(class *storagev1.StorageClass, api *apiv1a1.UpdateStorageClassRequest) {
	SetObjectAlias(class, api.Alias)
	SetObjectDescription(class, api.Description)
}

func (t *ApiToMetaTranslator) DataVolumeNewFromCreate(api *apiv1a1.CreateDataVolumeRequest,
	sc *storagev1.StorageClass, namespace string) *corev1.PersistentVolumeClaim {
	meta := new(corev1.PersistentVolumeClaim)
	resourceList := corev1.ResourceList{
		corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", api.Size)),
	}
	meta.ObjectMeta.Name = api.Name
	meta.ObjectMeta.Namespace = namespace
	meta.ObjectMeta.Annotations = map[string]string{
		kubernetes.LabelKeyKubeStorageProvisioner: sc.Provisioner,
	}
	meta.Spec.AccessModes = api.AccessModes
	meta.Spec.Resources = corev1.ResourceRequirements{
		Limits:   resourceList,
		Requests: resourceList,
	}
	SetPVCClass(meta, sc.Name)
	SetObjStorageAdminMark(meta)
	return meta
}

// special pre work
func GenGlusterfsSecret(serviceName, restuser, restuserkey string) *corev1.Secret {
	// gen name
	name := Md5String(serviceName) + "-" + RandomString()
	// set secret
	rawValue := fmt.Sprintf("%s:%s", restuser, restuserkey)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceSystem,
			Labels: map[string]string{
				constants.LabelKeyStorageService: serviceName,
			},
		},
		Data: map[string][]byte{
			"key": []byte(rawValue),
		},
		Type: corev1.SecretType(kubernetes.StorageClassProvisionerGlusterfs),
	}
	SetObjStorageAdminMark(secret)
	return secret
}

func ParseGlusterfsSecret(secret *corev1.Secret) (restuser, restuserkey string, e error) {
	if secret == nil {
		return "", "", fmt.Errorf("secret is nil")
	}
	sType := string(secret.Type)
	if sType != kubernetes.StorageClassProvisionerGlusterfs {
		return "", "", fmt.Errorf("secret type unexpected %v", sType)
	}
	key, ok := secret.Data["key"]
	if !ok || len(key) == 0 {
		return "", "", fmt.Errorf("secret key is empty")
	}
	vec := strings.SplitN(string(key), ":", 2)
	if len(vec) != 2 {
		return "", "", fmt.Errorf("secret key in bad format")
	}
	return vec[0], vec[1], nil
}

// finalizer

func HasClassFinalizer(finalizers []string) bool {
	return HasFinalizer(finalizers, constants.StorageClassControllerFinalizerName)
}
func HasFinalizer(finalizers []string, finalizer string) bool {
	for i := range finalizers {
		if finalizers[i] == finalizer {
			return true
		}
	}
	return false
}

func RemoveClassFinalizer(old []string) []string {
	return RemoveFinalizer(old, constants.StorageClassControllerFinalizerName)
}
func RemoveServiceFinalizer(old []string) []string {
	return RemoveFinalizer(old, constants.StorageServiceControllerFinalizerName)
}
func RemoveFinalizer(old []string, finalizer string) []string {
	finalizers := make([]string, 0, len(old))
	for i := range old {
		if old[i] != finalizer {
			finalizers = append(finalizers, old[i])
		}
	}
	return finalizers
}

func AddClassFinalizer(old []string) []string {
	return AddFinalizer(old, constants.StorageClassControllerFinalizerName)
}
func AddFinalizer(old []string, finalizer string) []string {
	if HasFinalizer(old, finalizer) {
		return old
	}
	return append(old, finalizer)
}

// filter

func objectAliasFilter(obj metav1.Object, lowName string) bool {
	alias := strings.ToLower(GetObjectAlias(obj))
	return strings.Contains(alias, lowName)
}

func StorageServiceTypeAndNameFilter(in []resv1b1.StorageService, typeName, name string) []resv1b1.StorageService {
	var (
		filters []func(service *resv1b1.StorageService) bool
		lowName = strings.ToLower(name)
		tmp     []*resv1b1.StorageService
	)
	// init
	if len(typeName) > 0 {
		filters = append(filters, func(service *resv1b1.StorageService) bool {
			return service.TypeName == typeName
		})
	}
	if len(name) > 0 {
		filters = append(filters, func(service *resv1b1.StorageService) bool {
			return objectAliasFilter(service, lowName)
		})
	}
	// do
	doFilter := func(service *resv1b1.StorageService) bool {
		for _, f := range filters {
			if f(service) == false {
				return false
			}
		}
		return true
	}
	for i := range in {
		if ok := doFilter(&in[i]); ok {
			tmp = append(tmp, &in[i])
		}
	}
	// sort
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Name < tmp[j].Name
	})
	// return
	out := make([]resv1b1.StorageService, len(tmp))
	for i := range tmp {
		tmp[i].DeepCopyInto(&out[i])
	}
	return out
}

func StorageClassTypeAndNameFilter(in []storagev1.StorageClass, typeName, name string) []storagev1.StorageClass {
	var (
		filters []func(service *storagev1.StorageClass) bool
		lowName = strings.ToLower(name)
		tmp     []*storagev1.StorageClass
	)
	// init
	if len(typeName) > 0 {
		filters = append(filters, func(sc *storagev1.StorageClass) bool {
			return GetClassType(sc) == typeName
		})
	}
	if len(name) > 0 {
		filters = append(filters, func(sc *storagev1.StorageClass) bool {
			return objectAliasFilter(sc, lowName)
		})
	}
	// do
	doFilter := func(sc *storagev1.StorageClass) bool {
		for _, f := range filters {
			if f(sc) == false {
				return false
			}
		}
		return true
	}
	for i := range in {
		if ok := doFilter(&in[i]); ok {
			tmp = append(tmp, &in[i])
		}
	}
	// sort
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Name < tmp[j].Name
	})
	// return
	out := make([]storagev1.StorageClass, len(tmp))
	for i := range tmp {
		tmp[i].DeepCopyInto(&out[i])
	}
	return out
}

// label

func SetObjStorageAdminMark(obj metav1.Object) {
	setObjectLabel(obj, constants.LabelKeyStorageAdminMarkKey, constants.LabelKeyStorageAdminMarkVal)
}

func IsObjStorageAdminMarked(obj metav1.Object) bool {
	mark, _ := getObjectLabel(obj, constants.LabelKeyStorageAdminMarkKey)
	return mark == constants.LabelKeyStorageAdminMarkVal
}

func SetClassService(sc *storagev1.StorageClass, svName string) {
	setObjectLabel(sc, constants.LabelKeyStorageService, svName)
}

func GetClassService(sc *storagev1.StorageClass) string {
	svName, _ := getObjectLabel(sc, constants.LabelKeyStorageService)
	return svName
}

func SetClassType(sc *storagev1.StorageClass, tpName string) {
	setObjectLabel(sc, constants.LabelKeyStorageType, tpName)
}

func GetClassType(sc *storagev1.StorageClass) string {
	tpName, _ := getObjectLabel(sc, constants.LabelKeyStorageType)
	return tpName
}

func GetPVCClass(pvc *corev1.PersistentVolumeClaim) string {
	scName, _ := getObjectLabel(pvc, kubernetes.LabelKeyKubeStorageClass)
	if len(scName) == 0 && pvc.Spec.StorageClassName != nil {
		scName = *pvc.Spec.StorageClassName
	}
	return scName
}

func SetPVCClass(pvc *corev1.PersistentVolumeClaim, scName string) {
	setObjectLabel(pvc, kubernetes.LabelKeyKubeStorageClass, scName)
	pvc.Spec.StorageClassName = &scName
}

func GetStorageSecretPath(pm map[string]string) (secretNamespace, secretName string, ok bool) {
	if pm != nil {
		secretName = pm[kubernetes.StorageClassParamNameSecretName]
		secretNamespace = pm[kubernetes.StorageClassParamNameSecretNamespace]
		ok = len(secretName) > 0 && len(secretNamespace) > 0
	}
	return
}

// random
func RandomString() string {
	return strings.Join([]string{
		strconv.FormatInt(startTime.UnixNano(), 36),
		strconv.FormatUint(atomic.AddUint64(&randSeq, 1), 36),
		strconv.FormatInt(time.Now().UnixNano(), 36),
	}, "-")
}

func Md5String(in string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(in)))
}
