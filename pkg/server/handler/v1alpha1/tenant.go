package v1alpha1

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/caicloud/nirvana/log"
	"k8s.io/api/core/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
)

// CreateTenant creates a cyclone tenant
func CreateTenant(ctx context.Context, tenant *api.Tenant) (*api.Tenant, error) {
	return tenant, createTenant(tenant)
}

// ListTenants list all tenants' information
func ListTenants(ctx context.Context) ([]api.Tenant, error) {
	namespaces, err := handler.K8sClient.CoreV1().Namespaces().List(meta_v1.ListOptions{
		LabelSelector: common.LabelOwnerCyclone(),
	})
	if err != nil {
		log.Errorf("List cyclone namespace error %v", err)
		return nil, err
	}

	tenants := []api.Tenant{}
	for _, namespace := range namespaces.Items {
		t, err := NamespaceToTenant(&namespace)
		if err != nil {
			log.Errorf("Unmarshal tenant annotation error %v", err)
			continue
		}
		tenants = append(tenants, *t)
	}
	return tenants, nil
}

// GetTenant gets information for a specific tenant
func GetTenant(ctx context.Context, name string) (*api.Tenant, error) {
	return getTenant(name)
}

func getTenant(name string) (*api.Tenant, error) {
	namespace, err := handler.K8sClient.CoreV1().Namespaces().Get(common.TenantNamespace(name), meta_v1.GetOptions{})
	if err != nil {
		log.Errorf("Get namespace for tenant %s error %v", name, err)
		return nil, err
	}

	return NamespaceToTenant(namespace)
}

// NamespaceToTenant trans namespace to tenant
func NamespaceToTenant(namespace *core_v1.Namespace) (*api.Tenant, error) {
	annotationTenant := namespace.Annotations[common.AnnotationTenant]

	tenant := &api.Tenant{}
	err := json.Unmarshal([]byte(annotationTenant), tenant)
	if err != nil {
		log.Errorf("Unmarshal tenant annotation error %v", err)
		return tenant, err
	}

	tenant.Metadata.CreationTime = namespace.ObjectMeta.CreationTimestamp.String()
	return tenant, nil
}

// UpdateTenant updates information for a specific tenant
func UpdateTenant(ctx context.Context, name string, newTenant *api.Tenant) (*api.Tenant, error) {
	// get old tenant
	tenant, err := getTenant(name)
	if err != nil {
		log.Errorf("get old tenant %s error %v", name, err)
		return nil, err
	}

	integrations := []api.Integration{}
	// update resource quota if necessary
	if !reflect.DeepEqual(tenant.Spec.ResourceQuota, newTenant.Spec.ResourceQuota) {
		integrations, err = GetWokerClusters(name)
		if err != nil {
			return nil, err
		}

		for _, integration := range integrations {
			cluster := integration.Spec.Cluster
			if cluster == nil {
				log.Warningf("cluster of integration %s is nil", integration.Metadata.Name)
				continue
			}

			client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
			if err != nil {
				log.Warningf("new cluster client for integration %s error %v", integration.Metadata.Name, err)
				continue
			}

			err = common.UpdateResourceQuota(newTenant, cluster.Namespace, client)
			if err != nil {
				log.Errorf("Update resource quota for tenant %s error %v", name, err)
				return nil, err
			}
		}
	}

	// update pvc if necessary
	if !reflect.DeepEqual(tenant.Spec.PersistentVolumeClaim, newTenant.Spec.PersistentVolumeClaim) {
		if len(integrations) == 0 {
			integrations, err = GetWokerClusters(name)
			if err != nil {
				return nil, err
			}
		}

		for _, integration := range integrations {
			cluster := integration.Spec.Cluster
			if cluster == nil {
				log.Warningf("cluster of integration %s is nil", integration.Metadata.Name)
				continue
			}

			client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
			if err != nil {
				log.Warningf("new cluster client for integration %s error %v", integration.Metadata.Name, err)
				continue
			}

			newPVC := newTenant.Spec.PersistentVolumeClaim
			err = common.UpdatePVC(tenant.Metadata.Name, newPVC.StorageClass, newPVC.Size, cluster.Namespace, client)
			if err != nil {
				log.Errorf("Update resource quota for tenant %s error %v", name, err)
				return nil, err
			}
		}
	}

	// update namespace
	err = updateTenantNamespace(newTenant)
	if err != nil {
		log.Errorf("Update namespace for tenant %s error %v", name, err)
		return nil, err
	}
	return newTenant, nil
}

// DeleteTenant deletes a tenant
func DeleteTenant(ctx context.Context, name string) error {
	err := handler.K8sClient.CoreV1().Namespaces().Delete(common.TenantNamespace(name), &meta_v1.DeleteOptions{})
	if err != nil {
		log.Errorf("Delete namespace for tenant %s error %v", name, err)
		return err
	}
	return nil
}

// CreateAdminTenant creates cyclone admin tenant
// First create namespace, then create pvc
func CreateAdminTenant() error {
	ns := common.TenantNamespace(common.AdminTenant)
	_, err := handler.K8sClient.CoreV1().Namespaces().Get(ns, meta_v1.GetOptions{})
	if err == nil {
		log.Infof("Default namespace %s already exist", ns)
		return nil
	}

	quota := map[core_v1.ResourceName]string{
		core_v1.ResourceLimitsCPU:      common.QuotaCPULimit,
		core_v1.ResourceLimitsMemory:   common.QuotaMemoryLimit,
		core_v1.ResourceRequestsCPU:    common.QuotaCPURequest,
		core_v1.ResourceRequestsMemory: common.QuotaMemoryRequest,
	}

	tenant := &api.Tenant{
		Metadata: api.Metadata{
			Name: common.AdminTenant,
		},
		Spec: api.TenantSpec{
			// TODO(zhujian7), read from configmap
			PersistentVolumeClaim: api.PersistentVolumeClaim{
				StorageClass: "", // use default storageclass
				Size:         common.DefaultPVCSize,
			},
			ResourceQuota: quota,
		},
	}

	return createTenant(tenant)
}

func createControlClusterIntegration(tenant string) error {
	in := &api.Integration{
		Metadata: api.Metadata{
			Name:        common.ControlClusterName,
			Description: "This is cluster is integrated by cyclone while creating tenant.",
		},
		Spec: api.IntegrationSpec{
			Type: api.Cluster,
			IntegrationSource: api.IntegrationSource{
				Cluster: &api.ClusterSource{
					IsControlCluster: true,
					IsWorkerCluster:  true,
					Namespace:        common.TenantNamespace(tenant),
				},
			},
		},
	}

	_, err := createIntegration(tenant, in)
	return err
}

func createTenant(tenant *api.Tenant) error {
	// create namespace
	err := createTenantNamespace(tenant)
	if err != nil {
		return err
	}

	// create cluster integration for control cluster
	err = createControlClusterIntegration(tenant.Metadata.Name)
	if err != nil {
		return err
	}

	// TODO(zhujian7): create built-in template-stage if tenant is admin

	return nil
}

func createTenantNamespace(tenant *api.Tenant) error {
	// marshal tenant and set it into namespace annotation
	namespace, err := buildNamespace(tenant)
	if err != nil {
		log.Warningf("Build namespace for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	_, err = handler.K8sClient.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		log.Errorf("Create namespace for tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	return nil
}

func updateTenantNamespace(tenant *api.Tenant) error {
	t, err := json.Marshal(tenant)
	if err != nil {
		log.Warningf("Marshal tenant %s error %v", tenant.Metadata.Name, err)
		return err
	}

	// update namespace annotation with retry
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		ns, err := handler.K8sClient.CoreV1().Namespaces().Get(common.TenantNamespace(tenant.Metadata.Name), meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Get namespace for tenant %s error %v", tenant.Metadata.Name, err)
			return err
		}

		ns.ObjectMeta.Annotations[common.AnnotationTenant] = string(t)

		_, err = handler.K8sClient.CoreV1().Namespaces().Update(ns)
		if err != nil {
			log.Errorf("Update namespace for tenant %s error %v", tenant.Metadata.Name, err)
			return err
		}
		return nil
	})

}

func buildNamespace(tenant *api.Tenant) (*v1.Namespace, error) {
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

	nsname := common.TenantNamespace(tenant.Metadata.Name)
	namespace := &v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        nsname,
			Labels:      label,
			Annotations: annotation,
		},
	}

	return namespace, nil
}
