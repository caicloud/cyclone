package v1alpha1

import (
	"context"
	"reflect"
	"sort"

	"github.com/caicloud/nirvana/log"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/integration/cluster"
	"github.com/caicloud/cyclone/pkg/server/biz/pvc"
	"github.com/caicloud/cyclone/pkg/server/biz/tenant"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
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
		LabelSelector: meta.LabelExistsSelector(meta.LabelTenantName),
	})
	if err != nil {
		log.Errorf("List cyclone namespace error %v", err)
		return nil, cerr.ConvertK8sError(err)
	}

	items := namespaces.Items
	if query.Sort {
		sort.Sort(sorter.NewNamespaceSorter(items, query.Ascending))
	}

	var tenants []api.Tenant
	for _, namespace := range items {
		t, err := tenant.FromNamespace(&namespace)
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
	return tenant.Get(handler.K8sClient, name)
}

// UpdateTenant updates information for a specific tenant
func UpdateTenant(ctx context.Context, name string, newTenant *api.Tenant) (*api.Tenant, error) {
	// Get old tenant
	t, err := tenant.Get(handler.K8sClient, name)
	if err != nil {
		log.Errorf("get old tenant %s error %v", name, err)
		return nil, err
	}

	integrations := []api.Integration{}
	// Update resource quota if necessary
	if !reflect.DeepEqual(t.Spec.ResourceQuota, newTenant.Spec.ResourceQuota) {
		integrations, err = cluster.GetSchedulableClusters(handler.K8sClient, name)
		if err != nil {
			return nil, err
		}

		for _, integration := range integrations {
			cluster := integration.Spec.Cluster
			if cluster == nil {
				log.Warningf("cluster of integration %s is nil", integration.Name)
				continue
			}

			if !cluster.IsWorkerCluster {
				log.Infof("%s is not worker cluster, skip", integration.Name)
				continue
			}

			client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
			if err != nil {
				log.Warningf("new cluster client for integration %s error %v", integration.Name, err)
				continue
			}

			err = svrcommon.UpdateResourceQuota(newTenant, cluster.Namespace, client)
			if err != nil {
				log.Errorf("update resource quota for tenant %s error %v", name, err)
				return nil, err
			}
		}
	}

	// Update pvc if necessary
	if !reflect.DeepEqual(t.Spec.PersistentVolumeClaim, newTenant.Spec.PersistentVolumeClaim) {
		if len(integrations) == 0 {
			integrations, err = cluster.GetSchedulableClusters(handler.K8sClient, name)
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

			if !cluster.IsWorkerCluster {
				log.Infof("%s is not worker cluster, skip", integration.Name)
				continue
			}

			client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
			if err != nil {
				log.Warningf("new cluster client for integration %s error %v", integration.Name, err)
				continue
			}

			log.Infof("old PersistentVolumeClaim: %+v ", t.Spec.PersistentVolumeClaim)
			log.Infof("new PersistentVolumeClaim: %+v ", newTenant.Spec.PersistentVolumeClaim)
			newPVC := newTenant.Spec.PersistentVolumeClaim
			err = pvc.UpdatePVC(t.Name, newPVC.StorageClass, newPVC.Size, cluster.Namespace, client)
			if err != nil {
				log.Errorf("update pvc for tenant %s error %v", name, err)
				return nil, err
			}
		}
	}

	// Update namespace
	err = tenant.UpdateNamespace(handler.K8sClient, newTenant)
	if err != nil {
		log.Errorf("Update namespace for tenant %s error %v", name, err)
		return nil, err
	}
	return newTenant, nil
}

// DeleteTenant deletes a tenant
func DeleteTenant(ctx context.Context, name string) error {
	// close workload cluster
	integrations, err := cluster.GetSchedulableClusters(handler.K8sClient, name)
	if err != nil {
		log.Errorf("get workload clusters for tenant %s error %v", name, err)
		return err
	}

	for _, integration := range integrations {
		if integration.Spec.Cluster == nil {
			log.Warningf("cluster of integration %s is nil", integration.Name)
			continue
		}

		err := cluster.Close(handler.K8sClient, &integration, name)
		if err != nil {
			log.Warningf("close cluster %s for tenant %s error %v", integration.Name, name, err)
			continue
		}
	}

	err = deleteCollections(name)
	if err != nil {
		return err
	}

	err = handler.K8sClient.CoreV1().Namespaces().Delete(svrcommon.TenantNamespace(name), &meta_v1.DeleteOptions{})
	if err != nil {
		log.Errorf("Delete namespace for tenant %s error %v", name, err)
		return cerr.ConvertK8sError(err)
	}

	return nil
}

// CreateDefaultTenant creates cyclone default tenant and initialize the tenant:
// - Create namespace
// - Create PVC
func CreateDefaultTenant() error {
	ns := svrcommon.TenantNamespace(svrcommon.DefaultTenant)
	_, err := handler.K8sClient.CoreV1().Namespaces().Get(ns, meta_v1.GetOptions{})
	if err == nil {
		log.Infof("Default namespace %s already exist", ns)
		return nil
	}

	annotations := make(map[string]string)
	annotations[meta.AnnotationDescription] = "This is the cyclone default tenant."
	annotations[meta.AnnotationAlias] = svrcommon.DefaultTenant

	tenant := &api.Tenant{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:        svrcommon.DefaultTenant,
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
	annotations[meta.AnnotationDescription] = "This cluster is integrated by cyclone while creating tenant."
	annotations[meta.AnnotationAlias] = common.ControlClusterName
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
					Namespace:       svrcommon.TenantNamespace(tenant),
				},
			},
		},
	}

	in, err := createIntegration(tenant, false, in)
	if err != nil {
		return cerr.ErrorCreateIntegration.Error(err)
	}

	if config.Config.OpenControlCluster {
		err := cluster.Open(handler.K8sClient, in, tenant)
		if err != nil {
			return err
		}
	}

	return nil
}

func createTenant(t *api.Tenant) error {
	// create namespace
	err := tenant.CreateNamespace(handler.K8sClient, t)
	if err != nil {
		return err
	}

	// create cluster integration for control cluster
	err = createControlClusterIntegration(t.Name)
	if err != nil {
		return err
	}

	return nil
}
