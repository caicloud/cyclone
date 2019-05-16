package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/pvc"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/biz/usage"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/config"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// ListIntegrations get integrations the given tenant has access to.
// - ctx Context of the reqeust
// - tenant Tenant
// - query Query params includes start, limit and filter.
func ListIntegrations(ctx context.Context, tenant string, query *types.QueryParams) (*types.ListResponse, error) {
	// TODO: Need a more efficient way to get paged items.
	secrets, err := handler.K8sClient.CoreV1().Secrets(svrcommon.TenantNamespace(tenant)).List(meta_v1.ListOptions{
		LabelSelector: meta.LabelIntegrationType,
	})
	if err != nil {
		log.Errorf("Get integrations from k8s with tenant %s error: %v", tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}

	items := secrets.Items
	var integrations []api.Integration
	size := int64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), integrations), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	if query.Sort {
		sort.Sort(sorter.NewSecretSorter(items, query.Ascending))
	}

	for _, secret := range items {
		integration, err := SecretToIntegration(&secret)
		if err != nil {
			continue
		}
		integrations = append(integrations, *integration)
	}

	return types.NewListResponse(int(size), integrations[query.Start:end]), nil
}

// SecretToIntegration translates secret to integration
func SecretToIntegration(secret *core_v1.Secret) (*api.Integration, error) {
	integration := &api.Integration{
		ObjectMeta: secret.ObjectMeta,
	}

	// retrieve integration name
	integration.Name = svrcommon.SecretIntegration(secret.Name)
	err := json.Unmarshal(secret.Data[svrcommon.SecretKeyIntegration], &integration.Spec)
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
		err := OpenClusterForTenant(in, tenant)
		if err != nil {
			return nil, err
		}
	}

	if in.Spec.Type == api.SCM && in.Spec.SCM != nil && in.Spec.SCM.Type != api.SVN {
		err := scm.GenerateSCMToken(in.Spec.SCM)
		if err != nil {
			return nil, err
		}
	}

	secret, err := buildSecret(tenant, in)
	if err != nil {
		return nil, err
	}

	ns := svrcommon.TenantNamespace(tenant)
	_, err = handler.K8sClient.CoreV1().Secrets(ns).Create(secret)
	if err != nil {
		log.Errorf("Create secret %v for tenant %s error %v", secret.ObjectMeta.Name, tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}

	return in, nil
}

// OpenClusterForTenant opens cluster to run workload.
func OpenClusterForTenant(in *api.Integration, tenantName string) (err error) {
	// Convert the returned error if it is a k8s error.
	defer func() {
		err = cerr.ConvertK8sError(err)
	}()

	cluster := in.Spec.Cluster

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

	if cluster.Namespace != "" && cluster.Namespace != svrcommon.TenantNamespace(tenant.Name) {
		// check if namespace exist
		_, err = client.CoreV1().Namespaces().Get(cluster.Namespace, meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("get namespace %s error %v", cluster.Namespace, err)
			return
		}
	} else {
		// create namespace
		cluster.Namespace = svrcommon.TenantNamespace(tenant.Name)
		err = svrcommon.CreateNamespace(tenant.Name, client)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Errorf("Create namespace for tenant %s error %v", tenantName, err)
				return
			}
			log.Infof("Namespace %s already exist", cluster.Namespace)
		}
	}

	// create resource quota
	err = svrcommon.CreateResourceQuota(tenant, cluster.Namespace, client)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			log.Errorf("Create resource quota for tenant %s error %v", tenantName, err)
			return
		}
		log.Infof("Resource quota for tenant %s already exist", tenantName)
	}

	// If a PVC has been configured, check existence of it.
	if cluster.PVC != "" && cluster.PVC != svrcommon.TenantPVC(tenant.Name) {
		_, err = client.CoreV1().PersistentVolumeClaims(cluster.Namespace).Get(cluster.PVC, meta_v1.GetOptions{})
		if err != nil {
			log.Errorf("Get pvc %s error %v", cluster.PVC, err)
			return
		}
	} else {
		if tenant.Spec.PersistentVolumeClaim.Size == "" {
			tenant.Spec.PersistentVolumeClaim.Size = config.Config.DefaultPVCConfig.Size
		}
		if tenant.Spec.PersistentVolumeClaim.StorageClass == "" {
			tenant.Spec.PersistentVolumeClaim.StorageClass = config.Config.DefaultPVCConfig.StorageClass
		}

		err = pvc.CreatePVC(tenant.Name, tenant.Spec.PersistentVolumeClaim.StorageClass,
			tenant.Spec.PersistentVolumeClaim.Size, cluster.Namespace, client)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				log.Errorf("create pvc for tenant %s error %v", tenantName, err)
				return
			}
			log.Infof("PVC for tenant %s already exist", tenantName)
		}

		cluster.PVC = svrcommon.TenantPVC(tenant.Name)
	}

	clusterName := in.Spec.Cluster.ClusterName
	if in.Spec.Cluster.IsControlCluster {
		clusterName = common.ControlClusterName
	}

	// Create ExecutionCluster resource for Workflow Engine to use
	_, err = handler.K8sClient.CycloneV1alpha1().ExecutionClusters().Create(&v1alpha1.ExecutionCluster{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: clusterName,
		},
		Spec: v1alpha1.ExecutionClusterSpec{
			Credential: cluster.Credential,
		},
	})

	// Execution cluster is system level, so different tenants may try to create ExecutionCluster CR for the same
	// cluster, so if the CR already exists, just ignore it.
	if err != nil && errors.IsAlreadyExists(err) {
		log.Infof("ExecutionCluster resource for %s already exist", in.Name)
		return nil
	}

	return err
}

// CloseClusterForTenant close worker cluster for the tenant.
// It is dangerous since all pvc data will lost.
func CloseClusterForTenant(in *api.Integration, tenant string) (err error) {
	// Convert the returned error if it is a k8s error.
	defer func() {
		err = cerr.ConvertK8sError(err)
	}()
	cluster := in.Spec.Cluster

	// new cluster clientset
	client, err := common.NewClusterClient(&cluster.Credential, cluster.IsControlCluster)
	if err != nil {
		log.Errorf("new cluster client error %v", err)
		return
	}

	// delete namespace which is created by cyclone
	if cluster.Namespace == svrcommon.TenantNamespace(tenant) {
		// if is a user cluster and namespace are created by cyclone, delete the namespace directly
		if !cluster.IsControlCluster && cluster.Namespace == svrcommon.TenantNamespace(tenant) {
			err = client.CoreV1().Namespaces().Delete(cluster.Namespace, &meta_v1.DeleteOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					log.Errorf("delete namespace %s error %v", cluster.Namespace, err)
					return
				}
				log.Warningf("namespace %s not found", cluster.Namespace)
				err = nil
			}

			// if namespace is deleted, will exit, no need delete others resources.
			return
		}
	}

	// delete resource quota
	quotaName := svrcommon.TenantResourceQuota(tenant)
	err = client.CoreV1().ResourceQuotas(cluster.Namespace).Delete(quotaName, &meta_v1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Errorf("delete resource quota %s error %v", quotaName, err)
			return
		}
		log.Warningf("resource quota %s not found", quotaName)
		err = nil
	}

	// Delete the PVC watcher deployment.
	err = usage.DeletePVCUsageWatcher(client, cluster.Namespace)
	if err != nil {
		log.Warningf("Delete PVC watcher '%s' error: %v", usage.PVCWatcherName, err)
	}

	// delete pvc which is created by cyclone
	if cluster.PVC == svrcommon.TenantPVC(tenant) {
		err = client.CoreV1().PersistentVolumeClaims(cluster.Namespace).Delete(cluster.PVC, &meta_v1.DeleteOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				log.Errorf("delete pvc %s error %v", cluster.PVC, err)
				return
			}
			log.Warningf("pvc %s not found", cluster.PVC)
			err = nil
		}
		return
	}

	// TODO(ChenDe): Different tenants may use the same execution cluster, so we can't delete the ExecutionCluster simply here.
	//err = handler.K8sClient.CycloneV1alpha1().ExecutionClusters().Delete(in.Spec.Cluster.ClusterName, &meta_v1.DeleteOptions{})
	//if err != nil {
	//	log.Warningf("Delete ExecutionCluster resource error: %v", err)
	//	return nil
	//}

	return
}

func buildSecret(tenant string, in *api.Integration) (*core_v1.Secret, error) {
	objectMeta := in.ObjectMeta
	// build secret name
	objectMeta.Name = svrcommon.IntegrationSecret(in.Name)
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}

	objectMeta.Labels[meta.LabelIntegrationType] = string(in.Spec.Type)
	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil {
		worker := in.Spec.Cluster.IsWorkerCluster
		if worker {
			objectMeta.Labels = meta.AddSchedulableClusterLabel(objectMeta.Labels)
		}
	}

	integration, err := json.Marshal(in.Spec)
	if err != nil {
		log.Errorf("Marshal integration %v for tenant %s error %v", in.Name, tenant, err)
		return nil, err
	}
	data := make(map[string][]byte)
	data[svrcommon.SecretKeyIntegration] = integration

	secret := &core_v1.Secret{
		ObjectMeta: objectMeta,
		Data:       data,
	}

	return secret, nil
}

// GetIntegration gets an integration with the given name under given tenant.
func GetIntegration(ctx context.Context, tenant, name string) (*api.Integration, error) {
	return getIntegration(tenant, name)
}

func getIntegration(tenant, name string) (*api.Integration, error) {
	secret, err := handler.K8sClient.CoreV1().Secrets(svrcommon.TenantNamespace(tenant)).Get(
		svrcommon.IntegrationSecret(name), meta_v1.GetOptions{})
	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return SecretToIntegration(secret)
}

func updateClusterIntegration(tenant, name string, in *api.Integration) error {
	oldIn, err := getIntegration(tenant, name)
	if err != nil {
		log.Errorf("get integration %s error %v", name, err)
		return err
	}

	oldCluster := oldIn.Spec.Cluster
	cluster := in.Spec.Cluster
	if cluster.Namespace == "" {
		cluster.Namespace = oldCluster.Namespace
	} else if cluster.Namespace == oldCluster.Namespace && cluster.PVC == "" {
		cluster.PVC = oldCluster.PVC
	}
	// turn on worker cluster
	if !oldCluster.IsWorkerCluster && cluster.IsWorkerCluster {
		// open cluster for the tenant, create namespace and pvc
		err := OpenClusterForTenant(in, tenant)
		if err != nil {
			return err
		}
	}

	// turn off worker cluster
	if oldCluster.IsWorkerCluster && !cluster.IsWorkerCluster {
		// close cluster for the tenant, delete namespace
		err := CloseClusterForTenant(oldIn, tenant)
		if err != nil {
			return err
		}
	}

	// TODO(zhujian7): namespace or pvc changed
	if oldCluster.IsWorkerCluster && cluster.IsWorkerCluster {
		if oldCluster.Namespace != cluster.Namespace {
			log.Info("can not process updating namespace, namespace changed from %s to %s", oldCluster.Namespace, cluster.Namespace)
		}

		if oldCluster.PVC != cluster.PVC {
			log.Info("can not process updating pvc, pvc changed from %s to %s", oldCluster.PVC, cluster.PVC)
		}
	}

	return nil
}

// UpdateIntegration updates an integration with the given tenant name and integration name.
// If updated successfully, return the updated integration.
func UpdateIntegration(ctx context.Context, tenant, name string, in *api.Integration) (*api.Integration, error) {
	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil {
		err := updateClusterIntegration(tenant, name, in)
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

	err = updateSecret(svrcommon.TenantNamespace(tenant), svrcommon.IntegrationSecret(name), in.Spec.Type, secret)
	if err != nil {
		return nil, err
	}

	return in, nil
}

func updateSecret(namespace, secretName string, inteType api.IntegrationType, secret *core_v1.Secret) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CoreV1().Secrets(namespace).Get(
			secretName, meta_v1.GetOptions{})
		if err != nil {
			return err
		}

		newSecret := origin.DeepCopy()
		newSecret.Annotations = MergeMap(secret.Annotations, newSecret.Annotations)
		newSecret.Labels = MergeMap(secret.Labels, newSecret.Labels)
		newSecret.Labels[meta.LabelIntegrationType] = string(inteType)

		// Only use new datas to overwrite old ones, and keep others not needed to be overwritten, such as repos.
		for key, value := range secret.Data {
			newSecret.Data[key] = value
		}

		_, err = handler.K8sClient.CoreV1().Secrets(namespace).Update(newSecret)
		return err
	})

	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	return nil
}

// DeleteIntegration deletes a integration with the given tenant and name.
func DeleteIntegration(ctx context.Context, tenant, name string) error {
	in, err := getIntegration(tenant, name)
	if err != nil {
		return err
	}

	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil && in.Spec.Cluster.IsWorkerCluster {
		return cerr.ErrorClusterNotClosed.Error(in.Name)
	}

	secretName := svrcommon.IntegrationSecret(name)
	if in.Spec.Type == api.SCM {
		// Cleanup SCM webhooks for integrated SCM.
		secret, err := handler.K8sClient.CoreV1().Secrets(svrcommon.TenantNamespace(tenant)).Get(
			secretName, meta_v1.GetOptions{})
		if err != nil {
			return cerr.ConvertK8sError(err)
		}

		repos := map[string][]string{}
		if d, ok := secret.Data[svrcommon.SecretKeyRepos]; ok {
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
				if err = DeleteSCMWebhook(integration.Spec.SCM, tenant, secretName, repo); err != nil {
					// Only try best to cleanup webhooks, if there are errors, will not block the process.
					log.Error(err)
				}
			}
		}
	}

	err = handler.K8sClient.CoreV1().Secrets(svrcommon.TenantNamespace(tenant)).Delete(
		secretName, &meta_v1.DeleteOptions{})

	return cerr.ConvertK8sError(err)
}

// OpenCluster opens cluster type integration to execute workflow
func OpenCluster(ctx context.Context, tenant, name string) error {
	in, err := getIntegration(tenant, name)
	if err != nil {
		return err
	}

	if in.Spec.Type != api.Cluster {
		return cerr.ErrorIntegrationTypeNotCorrect.Error(name, api.Cluster, in.Spec.Type)
	}

	if in.Spec.Cluster == nil {
		return cerr.ErrorUnknownInternal.Error("cluster is nil")
	}

	if in.Spec.Cluster.IsWorkerCluster {
		return nil
	}

	// open cluster for the tenant, create namespace and pvc
	err = OpenClusterForTenant(in, tenant)
	if err != nil {
		return err
	}

	in.Spec.Cluster.IsWorkerCluster = true
	secret, err := buildSecret(tenant, in)
	if err != nil {
		return err
	}
	secret.Labels[meta.LabelIntegrationSchedulableCluster] = meta.LabelValueTrue

	return updateSecret(svrcommon.TenantNamespace(tenant), svrcommon.IntegrationSecret(name), in.Spec.Type, secret)
}

// CloseCluster closes cluster type integration that used to execute workflow
func CloseCluster(ctx context.Context, tenant, name string) error {
	in, err := getIntegration(tenant, name)
	if err != nil {
		return err
	}

	if in.Spec.Type != api.Cluster {
		return cerr.ErrorIntegrationTypeNotCorrect.Error(name, api.Cluster, in.Spec.Type)
	}

	if in.Spec.Cluster == nil {
		return cerr.ErrorUnknownInternal.Error("cluster is nil")
	}

	cluster := in.Spec.Cluster
	if !cluster.IsWorkerCluster {
		return nil
	}

	// close cluster for the tenant, delete namespace
	err = CloseClusterForTenant(in, tenant)
	if err != nil {
		return err
	}

	in.Spec.Cluster.IsWorkerCluster = false
	secret, err := buildSecret(tenant, in)
	if err != nil {
		return err
	}
	if _, ok := secret.Labels[meta.LabelIntegrationSchedulableCluster]; ok {
		delete(secret.Labels, meta.LabelIntegrationSchedulableCluster)
	}

	return updateSecret(svrcommon.TenantNamespace(tenant), svrcommon.IntegrationSecret(name), in.Spec.Type, secret)
}

// GetSchedulableClusters gets all clusters which are used to perform workload for this tenant.
func GetSchedulableClusters(tenant string) ([]api.Integration, error) {
	secrets, err := handler.K8sClient.CoreV1().Secrets(svrcommon.TenantNamespace(tenant)).List(meta_v1.ListOptions{
		LabelSelector: meta.SchedulableClusterSelector(),
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

func getSCMSourceFromIntegration(tenant, integrationName string) (*api.SCMSource, error) {
	integration, err := getIntegration(tenant, integrationName)
	if err != nil {
		return nil, err
	}

	if integration.Spec.Type != api.SCM {
		return nil, cerr.ErrorValidationFailed.Error(fmt.Sprintf("type of integration %s", integration.Name),
			fmt.Sprintf("only support %s type", api.SCM))
	}

	return integration.Spec.SCM, nil
}

// ListSCMRepos lists repos for integrated SCM under given tenant.
func ListSCMRepos(ctx context.Context, tenant, integrationName string) (*types.ListResponse, error) {
	scmSource, err := getSCMSourceFromIntegration(tenant, integrationName)
	if err != nil {
		return nil, err
	}

	repos, err := listSCMRepos(scmSource)
	if err != nil {
		log.Errorf("Failed to list repos for integration %s as %v", integrationName, err)
		return nil, err
	}

	return types.NewListResponse(len(repos), repos), nil

}

// ListSCMBranches lists branches for specified repo of integrated SCM under given tenant.
func ListSCMBranches(ctx context.Context, tenant, integrationName, repo string) (*types.ListResponse, error) {
	scmSource, err := getSCMSourceFromIntegration(tenant, integrationName)
	if err != nil {
		return nil, err
	}

	repo, err = url.PathUnescape(repo)
	if err != nil {
		return nil, err
	}
	branches, err := listSCMBranches(scmSource, repo)
	if err != nil {
		log.Errorf("Failed to list branches for integration %s's repo %s as %v", integrationName, repo, err)
		return nil, err
	}

	return types.NewListResponse(len(branches), branches), nil
}

// ListSCMTags lists tags for specified repo of integrated SCM under given tenant.
func ListSCMTags(ctx context.Context, tenant, integrationName, repo string) (*types.ListResponse, error) {
	scmSource, err := getSCMSourceFromIntegration(tenant, integrationName)
	if err != nil {
		return nil, err
	}

	repo, err = url.PathUnescape(repo)
	if err != nil {
		return nil, err
	}
	tags, err := listSCMTags(scmSource, repo)
	if err != nil {
		log.Errorf("Failed to list tags for integration %s's repo %s as %v", integrationName, repo, err)
		return nil, err
	}

	return types.NewListResponse(len(tags), tags), nil
}
