package v1alpha1

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/caicloud/nirvana/log"
	"k8s.io/api/core/v1"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// CreateTenant creates a cyclone tenant
func CreateTenant(ctx context.Context, tenant *api.Tenant) (*api.Tenant, error) {
	modifiers := []CreationModifier{GenerateNameModifier, TenantModifier}
	for _, modifier := range modifiers {
		err := modifier("", "", "", tenant)
		if err != nil {
			return nil, err
		}
	}

	return tenant, createTenant(tenant)
}

// ListTenants list all tenants' information
func ListTenants(ctx context.Context, query *types.QueryParams) (*types.ListResponse, error) {
	namespaces, err := handler.K8sClient.CoreV1().Namespaces().List(meta_v1.ListOptions{
		LabelSelector: common.LabelOwnerCyclone(),
	})
	if err != nil {
		log.Errorf("List cyclone namespace error %v", err)
		return nil, cerr.ConvertK8sError(err)
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

	size := int64(len(tenants))
	if query.Start >= size {
		return types.NewListResponse(int(size), []api.Tenant{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), tenants[query.Start:end]), nil
}

// GetTenant gets information for a specific tenant
func GetTenant(ctx context.Context, name string) (*api.Tenant, error) {
	return getTenant(name)
}

func getTenant(name string) (*api.Tenant, error) {
	namespace, err := handler.K8sClient.CoreV1().Namespaces().Get(common.TenantNamespace(name), meta_v1.GetOptions{})
	if err != nil {
		log.Errorf("Get namespace for tenant %s error %v", name, err)
		return nil, cerr.ConvertK8sError(err)
	}

	return NamespaceToTenant(namespace)
}

// NamespaceToTenant trans namespace to tenant
func NamespaceToTenant(namespace *core_v1.Namespace) (*api.Tenant, error) {
	tenant := &api.Tenant{
		ObjectMeta: namespace.ObjectMeta,
	}

	// retrieve tenant name
	tenant.Name = common.NamespaceTenant(namespace.Name)
	annotationTenant := namespace.Annotations[common.AnnotationTenant]
	err := json.Unmarshal([]byte(annotationTenant), &tenant.Spec)
	if err != nil {
		log.Errorf("Unmarshal tenant annotation %s error %v", annotationTenant, err)
		return tenant, err
	}

	// delete tenant annotation
	delete(tenant.Annotations, common.AnnotationTenant)
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
				log.Warningf("cluster of integration %s is nil", integration.Name)
				continue
			}

			client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
			if err != nil {
				log.Warningf("new cluster client for integration %s error %v", integration.Name, err)
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
				log.Warningf("cluster of integration %s is nil", integration.Name)
				continue
			}

			client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
			if err != nil {
				log.Warningf("new cluster client for integration %s error %v", integration.Name, err)
				continue
			}

			newPVC := newTenant.Spec.PersistentVolumeClaim
			err = common.UpdatePVC(tenant.Name, newPVC.StorageClass, newPVC.Size, cluster.Namespace, client)
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
	// close workload cluster
	integrations, err := GetWokerClusters(name)
	if err != nil {
		log.Errorf("get workload clusters for tenant %s error %v", name, err)
		return err
	}

	for _, integration := range integrations {
		cluster := integration.Spec.Cluster
		if cluster == nil {
			log.Warningf("cluster of integration %s is nil", integration.Name)
			continue
		}

		err := CloseClusterForTenant(cluster, name)
		if err != nil {
			log.Warningf("close cluster %s for tenant %s error %v", integration.Name, name, err)
			continue
		}
	}

	err = handler.K8sClient.CoreV1().Namespaces().Delete(common.TenantNamespace(name), &meta_v1.DeleteOptions{})
	if err != nil {
		log.Errorf("Delete namespace for tenant %s error %v", name, err)
		return cerr.ConvertK8sError(err)
	}

	return nil
}

// CreateAdminTenant creates cyclone admin tenant and initialize the tenant:
// - Create namespace
// - Create PVC
// - Load and create stage templates
func CreateAdminTenant() error {
	ns := common.TenantNamespace(common.AdminTenant)
	_, err := handler.K8sClient.CoreV1().Namespaces().Get(ns, meta_v1.GetOptions{})
	if err == nil {
		log.Infof("Default namespace %s already exist", ns)
		return nil
	}

	annotations := make(map[string]string)
	annotations[common.AnnotationDescription] = "This is the administrator tenant."
	annotations[common.AnnotationAlias] = common.AdminTenant

	tenant := &api.Tenant{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        common.AdminTenant,
			Annotations: annotations,
		},
		Spec: api.TenantSpec{
			PersistentVolumeClaim: api.PersistentVolumeClaim{
				Size: config.Config.DefaultPVCConfig.Size,
			},
			ResourceQuota: config.Config.WorkerNamespaceQuota,
		},
	}

	if config.Config.DefaultPVCConfig.StorageClass != "" {
		tenant.Spec.PersistentVolumeClaim.StorageClass = config.Config.DefaultPVCConfig.StorageClass
	}

	return createTenant(tenant)
}

func createControlClusterIntegration(tenant string) error {
	annotations := make(map[string]string)
	annotations[common.AnnotationDescription] = "This cluster is integrated by cyclone while creating tenant."
	annotations[common.AnnotationAlias] = common.ControlClusterName
	in := &api.Integration{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        common.ControlClusterName,
			Annotations: annotations,
		},
		Spec: api.IntegrationSpec{
			Type: api.Cluster,
			IntegrationSource: api.IntegrationSource{
				Cluster: &api.ClusterSource{
					IsControlCluster: true,
					// Cluster by default is not enabled to run workflow, need users to enable it explicitly.
					IsWorkerCluster: false,
					Namespace:       common.TenantNamespace(tenant),
				},
			},
		},
	}

	_, err := createIntegration(tenant, in)
	if err != nil {
		return cerr.ErrorCreateIntegration.Error(err)
	}

	return nil
}

func createTenant(tenant *api.Tenant) error {
	// create namespace
	err := createTenantNamespace(tenant)
	if err != nil {
		return err
	}

	// create cluster integration for control cluster
	err = createControlClusterIntegration(tenant.Name)
	if err != nil {
		return err
	}

	return nil
}

func createTenantNamespace(tenant *api.Tenant) error {
	meta := tenant.ObjectMeta

	// build namespace name
	meta.Name = common.TenantNamespace(tenant.Name)

	// marshal tenant and set it into namespace annotation
	t, err := json.Marshal(tenant.Spec)
	if err != nil {
		log.Warningf("Marshal tenant %s error %v", tenant.Name, err)
		return err
	}
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	meta.Annotations[common.AnnotationTenant] = string(t)

	// set labels
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[common.LabelOwner] = common.OwnerCyclone

	_, err = handler.K8sClient.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: meta,
	})
	if err != nil {
		log.Errorf("Create namespace for tenant %s error %v", tenant.Name, err)
		if errors.IsAlreadyExists(err) {
			tenant.Labels = meta.Labels
			return updateTenantNamespace(tenant)
		}
		return cerr.ConvertK8sError(err)
	}

	return nil
}

func updateTenantNamespace(tenant *api.Tenant) error {
	t, err := json.Marshal(tenant.Spec)
	if err != nil {
		log.Warningf("Marshal tenant %s error %v", tenant.Name, err)
		return err
	}

	// update namespace annotation with retry
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CoreV1().Namespaces().Get(common.TenantNamespace(tenant.Name), meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Get namespace for tenant %s error %v", tenant.Name, err)
			return cerr.ConvertK8sError(err)
		}

		newNs := origin.DeepCopy()
		newNs.Annotations = MergeMap(tenant.Annotations, newNs.Annotations)
		newNs.Labels = MergeMap(tenant.Labels, newNs.Labels)
		newNs.Annotations[common.AnnotationTenant] = string(t)

		_, err = handler.K8sClient.CoreV1().Namespaces().Update(newNs)
		if err != nil {
			log.Errorf("Update namespace for tenant %s error %v", tenant.Name, err)
			return cerr.ConvertK8sError(err)
		}
		return nil
	})

}
