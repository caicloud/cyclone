package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// ListIntegrations get integrations the given tenant has access to.
// - ctx Context of the reqeust
// - tenant Tenant
// - pagination Pagination with page and limit.
func ListIntegrations(ctx context.Context, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	// TODO: Need a more efficient way to get paged items.
	secrets, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).List(meta_v1.ListOptions{
		LabelSelector: common.LabelIntegrationType,
	})
	if err != nil {
		log.Errorf("Get integrations from k8s with tenant %s error: %v", tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}

	items := secrets.Items
	integrations := []api.Integration{}
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), integrations), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	for _, secret := range items {
		integration, err := SecretToIntegration(&secret)
		if err != nil {
			continue
		}
		integrations = append(integrations, *integration)
	}

	return types.NewListResponse(int(size), integrations[pagination.Start:end]), nil
}

// SecretToIntegration translates secret to integration
func SecretToIntegration(secret *core_v1.Secret) (*api.Integration, error) {
	integration := &api.Integration{
		ObjectMeta: secret.ObjectMeta,
	}

	// retrieve integration name
	integration.Name = common.SecretIntegration(secret.Name)
	err := json.Unmarshal(secret.Data[common.SecretKeyIntegration], &integration.Spec)
	if err != nil {
		return integration, err
	}

	return integration, nil
}

// CreateIntegration creates an integration to store external system info for the tenant.
func CreateIntegration(ctx context.Context, tenant string, in *api.Integration) (*api.Integration, error) {
	modifiers := []CreationModifier{GenerateNameModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, "", "", in)
		if err != nil {
			return nil, err
		}
	}

	return createIntegration(tenant, in)
}

func createIntegration(tenant string, in *api.Integration) (*api.Integration, error) {
	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil && in.Spec.Cluster.IsWorkerCluster {
		// open cluster for the tenant, create namespace and pvc
		err := OpenClusterForTenant(in.Spec.Cluster, tenant)
		if err != nil {
			return nil, err
		}
	}

	if in.Spec.Type == api.SCM && in.Spec.SCM != nil {
		err := scm.GenerateSCMToken(in.Spec.SCM)
		if err != nil {
			return nil, err
		}
	}

	secret, err := buildSecret(tenant, in)
	if err != nil {
		return nil, err
	}

	ns := common.TenantNamespace(tenant)
	_, err = handler.K8sClient.CoreV1().Secrets(ns).Create(secret)
	if err != nil {
		log.Errorf("Create secret %v for tenant %s error %v", secret.ObjectMeta.Name, tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}

	return in, nil
}

// OpenClusterForTenant opens cluster to run workload.
func OpenClusterForTenant(cluster *api.ClusterSource, tenantName string) (err error) {
	// Convert the returned error if it is a k8s error.
	defer func() {
		err = cerr.ConvertK8sError(err)
	}()

	// new cluster clientset
	client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
	if err != nil {
		log.Errorf("new cluster client for tenant %s error %v", tenantName, err)
		return
	}

	tenant, err := getTenant(tenantName)
	if err != nil {
		log.Errorf("get tenant %s info error %v", tenantName, err)
		return
	}

	if cluster.Namespace != "" {
		// check if namespace exist
		_, err = client.CoreV1().Namespaces().Get(cluster.Namespace, meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("get namespace %s error %v", cluster.Namespace, err)
			return
		}
	} else {
		// create namespace
		cluster.Namespace = common.TenantNamespace(tenant.Name)
		err = common.CreateNamespace(cluster.Namespace, client)
		if err != nil {
			log.Errorf("create user cluster namespace for tenant %s error %v", tenantName, err)
			if !errors.IsAlreadyExists(err) {
				return
			}
		}
	}

	// create resource quota
	err = common.CreateResourceQuota(tenant, cluster.Namespace, client)
	if err != nil {
		log.Errorf("create resource quota for tenant %s error %v", tenantName, err)
		if !errors.IsAlreadyExists(err) {
			return
		}
	}

	if cluster.PVC != "" {
		_, err = client.CoreV1().PersistentVolumeClaims(cluster.Namespace).Get(cluster.PVC, meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("get pvc %s error %v", cluster.PVC, err)
			return
		}
	} else {
		// create pvc
		if tenant.Spec.PersistentVolumeClaim.Size == "" {
			tenant.Spec.PersistentVolumeClaim.Size = config.Config.DefaultPVCConfig.Size
		}

		err = common.CreatePVC(tenant.Name, tenant.Spec.PersistentVolumeClaim.StorageClass,
			tenant.Spec.PersistentVolumeClaim.Size, cluster.Namespace, client)
		if err != nil {
			log.Errorf("create pvc for tenant %s error %v", tenantName, err)
			if !errors.IsAlreadyExists(err) {
				return
			}
		}

		cluster.PVC = common.TenantPVC(tenant.Name)
	}

	return nil
}

// CloseClusterForTenant close worker cluster for the tenant.
// It is dangerous since all pvc data will lost.
func CloseClusterForTenant(cluster *api.ClusterSource, tenant string) (err error) {
	// Convert the returned error if it is a k8s error.
	defer func() {
		err = cerr.ConvertK8sError(err)
	}()

	// new cluster clientset
	client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
	if err != nil {
		log.Errorf("new cluster client error %v", err)
		return
	}

	// delete namespace which is created by cyclone
	if cluster.Namespace == common.TenantNamespace(tenant) {
		// cyclone can not support other clusters (except for control cluster)by now, so we can not delete the namespace,
		// since there are metadata(cyclone resources, stage, workflow, workflowrun metadata) under it.

		/*
			err = client.CoreV1().Namespaces().Delete(cluster.Namespace, &meta_v1.DeleteOptions{})
			if err != nil {
				log.Errorf("delete namespace %s error %v", cluster.Namespace, err)
				return
			}
			// if namespace is deleted, will exit, no need delete others resources.
			return
		*/
	}

	// delete resource quota
	quotaName := common.TenantResourceQuota(tenant)
	err = client.CoreV1().ResourceQuotas(cluster.Namespace).Delete(quotaName, &meta_v1.DeleteOptions{})
	if err != nil {
		log.Errorf("delete resource quota %s error %v", quotaName, err)
		return
	}

	// delete pvc which is created by cyclone
	if cluster.PVC == common.TenantPVC(tenant) {
		err = client.CoreV1().PersistentVolumeClaims(cluster.Namespace).Delete(cluster.PVC, &meta_v1.DeleteOptions{})
		if err != nil {
			log.Errorf("delete pvc %s error %v", cluster.Namespace, err)
			return
		}
		return
	}
	return
}

func buildSecret(tenant string, in *api.Integration) (*core_v1.Secret, error) {
	meta := in.ObjectMeta
	// build secret name
	meta.Name = common.IntegrationSecret(in.Name)
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}

	meta.Labels[common.LabelIntegrationType] = string(in.Spec.Type)
	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil {
		worker := in.Spec.Cluster.IsWorkerCluster
		if worker {
			meta.Labels[common.LabelClusterOn] = common.LabelTrueValue
		}
	}

	integration, err := json.Marshal(in.Spec)
	if err != nil {
		log.Errorf("Marshal integration %v for tenant %s error %v", in.Name, tenant, err)
		return nil, err
	}
	data := make(map[string][]byte)
	data[common.SecretKeyIntegration] = integration

	secret := &core_v1.Secret{
		ObjectMeta: meta,
		Data:       data,
	}

	return secret, nil
}

// GetIntegration gets an integration with the given name under given tenant.
func GetIntegration(ctx context.Context, tenant, name string) (*api.Integration, error) {
	return getIntegration(tenant, name)
}

func getIntegration(tenant, name string) (*api.Integration, error) {
	secret, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(
		common.IntegrationSecret(name), meta_v1.GetOptions{})
	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return SecretToIntegration(secret)
}

// UpdateIntegration updates an integration with the given tenant name and integration name.
// If updated successfully, return the updated integration.
func UpdateIntegration(ctx context.Context, tenant, name string, in *api.Integration) (*api.Integration, error) {
	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil {
		oldIn, err := getIntegration(tenant, name)
		if err != nil {
			log.Errorf("get integration %s error %v", name, err)
			return nil, err
		}

		// turn on worker cluster
		if !oldIn.Spec.Cluster.IsWorkerCluster && in.Spec.Cluster.IsWorkerCluster {
			// open cluster for the tenant, create namespace and pvc
			err := OpenClusterForTenant(in.Spec.Cluster, tenant)
			if err != nil {
				return nil, err
			}
		}

		// turn off worker cluster
		if oldIn.Spec.Cluster.IsWorkerCluster && !in.Spec.Cluster.IsWorkerCluster {
			// close cluster for the tenant, delete namespace
			err := CloseClusterForTenant(in.Spec.Cluster, tenant)
			if err != nil {
				return nil, err
			}
		}

		// TODO(zhujian7): namespace or pvc changed
		if oldIn.Spec.Cluster.IsWorkerCluster && in.Spec.Cluster.IsWorkerCluster {
		}
	}

	if in.Spec.Type == api.SCM && in.Spec.SCM != nil {
		err := scm.GenerateSCMToken(in.Spec.SCM)
		if err != nil {
			return nil, err
		}
	}

	ns := common.TenantNamespace(tenant)
	secret, err := buildSecret(tenant, in)
	if err != nil {
		return nil, err
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CoreV1().Secrets(ns).Get(
			common.IntegrationSecret(name), meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		newSecret := origin.DeepCopy()
		newSecret.Annotations = MergeMap(secret.Annotations, newSecret.Annotations)
		newSecret.Labels = MergeMap(secret.Labels, newSecret.Labels)
		newSecret.Labels[common.LabelIntegrationType] = string(in.Spec.Type)

		// Only use new datas to overwrite old ones, and keep others not needed to be overwritten, such as repos.
		for key, value := range secret.Data {
			newSecret.Data[key] = value
		}

		_, err = handler.K8sClient.CoreV1().Secrets(ns).Update(newSecret)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return in, nil
}

// DeleteIntegration deletes a integration with the given tenant and name.
func DeleteIntegration(ctx context.Context, tenant, name string) error {
	isName := common.IntegrationSecret(name)
	in, err := getIntegration(tenant, isName)
	if err != nil {
		return err
	}

	if in.Spec.Type == api.SCM {
		// Cleanup SCM webhooks for integrated SCM.
		secret, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(
			isName, meta_v1.GetOptions{})
		if err != nil {
			return cerr.ConvertK8sError(err)
		}

		repos := map[string][]string{}
		if d, ok := secret.Data[common.SecretKeyRepos]; ok {
			log.Infof("repos data of secret %s: %s\n", secret.Name, d)
			if err = json.Unmarshal(d, &repos); err != nil {
				log.Errorf("Failed to unmarshal repos from secret")
				return err
			}
		}

		if len(repos) > 0 {
			log.Infoln("Delete webhook.")
			integration, err := SecretToIntegration(secret)
			if err != nil {
				log.Error(err)
				return err
			}

			for repo := range repos {
				if err = DeleteSCMWebhook(integration.Spec.SCM, tenant, isName, repo); err != nil {
					// Only try best to cleanup webhooks, if there are errors, will not block the process.
					log.Error(err)
				}
			}
		}
	}

	err = handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Delete(
		isName, &meta_v1.DeleteOptions{})

	return cerr.ConvertK8sError(err)
}

// GetWokerClusters gets all clusters which are use to perform workload
func GetWokerClusters(tenant string) ([]api.Integration, error) {
	secrets, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).List(meta_v1.ListOptions{
		LabelSelector: common.WorkerClustersSelector(),
	})
	if err != nil {
		log.Errorf("Get integrations from k8s with tenant %s error: %v", tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}

	integrations := []api.Integration{}
	for _, secret := range secrets.Items {
		integration, err := SecretToIntegration(&secret)
		if err != nil {
			continue
		}
		integrations = append(integrations, *integration)
	}

	return integrations, nil
}
