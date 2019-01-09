package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/caicloud/nirvana/log"
	"k8s.io/api/core/v1"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1alpha1 "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/handler/common"
)

// CreateTenant creates a cyclone tenant
func CreateTenant(ctx context.Context, tenant *apiv1alpha1.Tenant) (*apiv1alpha1.Tenant, error) {
	return tenant, createCycloneTenant(tenant)
}

// ListTenants list all tenants' information
func ListTenants(ctx context.Context) ([]apiv1alpha1.Tenant, error) {
	namespaces, err := common.K8sClient.CoreV1().Namespaces().List(meta_v1.ListOptions{
		LabelSelector: common.LabelOwner + "=" + common.OwnerCyclone,
	})
	if err != nil {
		log.Errorf("List cyclone namespace error %v", err)
		return nil, err
	}

	tenants := []apiv1alpha1.Tenant{}
	for _, namespace := range namespaces.Items {
		annotationTenant := namespace.Annotations[common.AnnotationTenant]

		t := apiv1alpha1.Tenant{}
		err := json.Unmarshal([]byte(annotationTenant), &t)
		if err != nil {
			log.Errorf("Unmarshal tenant annotation error %v", err)
			continue
		}

		tenants = append(tenants, t)
	}
	return tenants, nil
}

// GetTenant gets information for a specific tenant
func GetTenant(ctx context.Context, name string) (*apiv1alpha1.Tenant, error) {
	namespace, err := common.K8sClient.CoreV1().Namespaces().Get(common.TeanntNamespacePrefix+name, meta_v1.GetOptions{})
	if err != nil {
		log.Errorf("Get namespace for tenant %s error %v", name, err)
		return nil, err
	}

	annotationTenant := namespace.Annotations[common.AnnotationTenant]

	tenant := &apiv1alpha1.Tenant{}
	err = json.Unmarshal([]byte(annotationTenant), tenant)
	if err != nil {
		log.Errorf("Unmarshal tenant annotation error %v", err)
		return tenant, err
	}

	return tenant, nil
}

// UpdateTenant updates information for a specific tenant
func UpdateTenant(ctx context.Context, name string, newTenant *apiv1alpha1.Tenant) (*apiv1alpha1.Tenant, error) {
	// update namespace
	err := updateTenantNamespace(newTenant)
	if err != nil {
		log.Errorf("Update namespace for tenant %s error %v", name, err)
		return nil, err
	}

	// TODO , update pvc

	// update resource quota
	err = updateResourceQuota(newTenant)
	if err != nil {
		log.Errorf("Update resource quota for tenant %s error %v", name, err)
		return nil, err
	}

	return newTenant, nil
}

// DeleteTenant deletes a tenant
func DeleteTenant(ctx context.Context, name string) error {
	err := common.K8sClient.CoreV1().Namespaces().Delete(common.TeanntNamespacePrefix+name, &meta_v1.DeleteOptions{})
	if err != nil {
		log.Errorf("Delete namespace for tenant %s error %v", name, err)
		return err
	}
	return nil
}

// CreateDefaultTenant creates cyclone default tenant
// First create namespace, then create pvc
func CreateDefaultTenant() error {
	_, err := common.K8sClient.CoreV1().Namespaces().Get(common.DefaultTenantNamespace, meta_v1.GetOptions{})
	if err == nil {
		log.Infof("Default namespace %s already exist", common.DefaultTenantNamespace)
		return nil
	}

	quota := map[core_v1.ResourceName]string{
		core_v1.ResourceLimitsCPU:      common.QuotaCPULimit,
		core_v1.ResourceLimitsMemory:   common.QuotaMemoryLimit,
		core_v1.ResourceRequestsCPU:    common.QuotaCPURequest,
		core_v1.ResourceRequestsMemory: common.QuotaMemoryRequest,
	}

	tenant := &apiv1alpha1.Tenant{
		Metadata: apiv1alpha1.TenantMetadata{
			Name: common.DefaultTenant,
		},
		Spec: apiv1alpha1.TenantSpec{
			// TODO(zhujian7), read from configmap
			PersistentVolumeClaim: apiv1alpha1.PersistentVolumeClaim{
				StorageClass: "", // use default storageclass
				Size:         common.DefaultPVCSize,
			},
			ResourceQuota: quota,
		},
	}

	return createCycloneTenant(tenant)
}

func createCycloneTenant(tenant *apiv1alpha1.Tenant) error {
	// create namespace
	err := createTenantNamespace(tenant)
	if err != nil {
		return err
	}

	// create resouce quota
	err = createResourceQuota(tenant)
	if err != nil {
		return err
	}

	// TODO(zhujian7), create cluster integration for control cluster

	// create pvc
	if tenant.Spec.PersistentVolumeClaim.Size == "" {
		tenant.Spec.PersistentVolumeClaim.Size = common.DefaultPVCSize
	}

	return createTenantPVC(tenant.Metadata.Name,
		tenant.Spec.PersistentVolumeClaim.StorageClass, tenant.Spec.PersistentVolumeClaim.Size)

}

func createTenantNamespace(tenant *apiv1alpha1.Tenant) error {
	// marshal tenant and set it into namespace annotation
	namespace, err := buildNamespace(tenant)
	if err != nil {
		log.Warningf("Build namespace for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	_, err = common.K8sClient.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		log.Errorf("Create namespace for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	return nil
}

func updateTenantNamespace(tenant *apiv1alpha1.Tenant) error {
	// marshal tenant and set it into namespace annotation
	namespace, err := buildNamespace(tenant)
	if err != nil {
		log.Warningf("Build namespace for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	_, err = common.K8sClient.CoreV1().Namespaces().Update(namespace)
	if err != nil {
		log.Errorf("Update namespace for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	return nil
}

func buildNamespace(tenant *apiv1alpha1.Tenant) (*v1.Namespace, error) {
	// marshal tenant and set it into namespace annotation
	annotation := make(map[string]string)
	t, err := json.Marshal(tenant)
	if err != nil {
		log.Warningf("Marshal tenant %s error %v", tenant.Metadata.Name, err)
		return nil, err
	}
	annotation[common.AnnotationTenant] = string(t)

	// set labels
	label := make(map[string]string)
	label[common.LabelOwner] = common.OwnerCyclone

	nsname := common.TeanntNamespacePrefix + tenant.Metadata.Name
	namespace := &v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        nsname,
			Labels:      label,
			Annotations: annotation,
		},
	}

	return namespace, nil
}

func createResourceQuota(tenant *apiv1alpha1.Tenant) error {
	quota, err := buildResourceQuota(tenant)
	if err != nil {
		log.Warningf("Build resource quota for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}
	nsname := common.TeanntNamespacePrefix + tenant.Metadata.Name

	_, err = common.K8sClient.CoreV1().ResourceQuotas(nsname).Create(quota)
	if err != nil {
		log.Errorf("Create ResourceQuota for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	return nil
}

func buildResourceQuota(tenant *apiv1alpha1.Tenant) (*v1.ResourceQuota, error) {
	// parse resource list
	rl, err := ParseResourceList(tenant.Spec.ResourceQuota)
	if err != nil {
		log.Warningf("Parse resource quota for tenant %s error %v", tenant.Metadata.Name, err)
		return nil, err
	}

	nsname := common.TeanntNamespacePrefix + tenant.Metadata.Name
	quota := &core_v1.ResourceQuota{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      tenant.Metadata.Name,
			Namespace: nsname,
		},
		Spec: core_v1.ResourceQuotaSpec{
			Hard: rl,
		},
	}

	return quota, nil
}

func updateResourceQuota(tenant *apiv1alpha1.Tenant) error {
	quota, err := buildResourceQuota(tenant)
	if err != nil {
		log.Warningf("Build resource quota for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}
	nsname := common.TeanntNamespacePrefix + tenant.Metadata.Name

	_, err = common.K8sClient.CoreV1().ResourceQuotas(nsname).Update(quota)
	if err != nil {
		log.Errorf("Update ResourceQuota for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	return nil
}

func createTenantPVC(tenantName, storageClass, size string) error {
	// parse quantity
	resources := make(map[core_v1.ResourceName]resource.Quantity)
	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		log.Errorf("Parse Quantity %s error %v", size, err)
		return err
	}
	resources[core_v1.ResourceStorage] = quantity

	// create pvc
	pvcName := common.TenantPVCPrefix + tenantName
	namespace := common.TeanntNamespacePrefix + tenantName
	volume := &core_v1.PersistentVolumeClaim{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
		},
		Spec: core_v1.PersistentVolumeClaimSpec{
			AccessModes: []core_v1.PersistentVolumeAccessMode{core_v1.ReadWriteMany},
			Resources: core_v1.ResourceRequirements{
				Limits:   resources,
				Requests: resources,
			},
		},
	}

	if storageClass != "" {
		volume.Spec.StorageClassName = &storageClass
	}

	_, err = common.K8sClient.CoreV1().PersistentVolumeClaims(namespace).Create(volume)
	if err != nil {
		log.Errorf("Create persistent volume claim %s error %v", pvcName, err)
		return err
	}

	return nil
}

// ParseResourceList parse resouces from 'map[string]string' to 'ResourceList'
func ParseResourceList(resources map[core_v1.ResourceName]string) (map[core_v1.ResourceName]resource.Quantity, error) {
	rl := make(map[core_v1.ResourceName]resource.Quantity)

	for r, q := range resources {
		quantity, err := resource.ParseQuantity(q)
		if err != nil {
			log.Errorf("Parse %s Quantity %s error %v", r, q, err)
			return nil, err
		}
		rl[r] = quantity
	}

	return rl, nil
}
