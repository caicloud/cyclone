package v1alpha1

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/caicloud/nirvana/log"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/integration"
	"github.com/caicloud/cyclone/pkg/server/biz/integration/cluster"
	"github.com/caicloud/cyclone/pkg/server/biz/integration/sonarqube"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// ListIntegrations get integrations for the given tenant.
func ListIntegrations(ctx context.Context, tenant string, includePublic bool, query *types.QueryParams) (*types.ListResponse, error) {
	// TODO: Need a more efficient way to get paged items.
	secrets, err := handler.K8sClient.CoreV1().Secrets(svrcommon.TenantNamespace(tenant)).List(meta_v1.ListOptions{
		LabelSelector: meta.LabelIntegrationType,
	})
	if err != nil {
		log.Errorf("Get secrets from k8s with tenant %s error: %v", tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}

	items := secrets.Items

	if includePublic {
		systemNamespace := common.GetSystemNamespace()
		publicSecrets, err := handler.K8sClient.CoreV1().Secrets(systemNamespace).List(meta_v1.ListOptions{
			LabelSelector: meta.LabelIntegrationType,
		})
		if err != nil {
			log.Errorf("Get secrets from system namespace %s error: %v", systemNamespace, err)
			return nil, err
		}

		items = append(items, publicSecrets.Items...)
	}

	var integrations []api.Integration

	if query.Sort {
		sort.Sort(sorter.NewSecretSorter(items, query.Ascending))
	}

	for _, secret := range items {
		i, err := integration.FromSecret(&secret)
		if err != nil {
			continue
		}
		integrations = append(integrations, *i)
	}

	results, err := filterIntegrations(integrations, query.Filter)
	if err != nil {
		return nil, err
	}

	size := int64(len(results))
	if query.Start >= size {
		return types.NewListResponse(int(size), []api.Integration{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}
	return types.NewListResponse(int(size), results[query.Start:end]), nil
}

func filterIntegrations(integrations []api.Integration, filter string) ([]api.Integration, error) {
	if filter == "" {
		return integrations, nil
	}

	var results []api.Integration
	// Support multiple filters rules, separated with comma.
	filterParts := strings.Split(filter, ",")
	filters := make(map[string]string)
	for _, part := range filterParts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			return nil, cerr.ErrorQueryParamNotCorrect.Error(filter)
		}

		filters[kv[0]] = strings.ToLower(kv[1])
	}

	var selected bool
	for _, itg := range integrations {
		selected = true
		for key, value := range filters {
			switch key {
			case "type":
				if itg.Labels != nil {
					if itgType, ok := itg.Labels[meta.LabelIntegrationType]; ok {
						if strings.EqualFold(itgType, value) {
							continue
						}
					}
				}
				selected = false
			}
		}

		if selected {
			results = append(results, itg)
		}
	}

	return results, nil
}

// CreateIntegration creates an integration to store external system info for the tenant.
func CreateIntegration(ctx context.Context, tenant string, isPublic bool, in *api.Integration) (*api.Integration, error) {
	modifiers := []CreationModifier{GenerateNameModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, "", "", in)
		if err != nil {
			return nil, err
		}
	}

	return createIntegration(tenant, isPublic, in)
}

func createIntegration(tenant string, isPublic bool, in *api.Integration) (*api.Integration, error) {
	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil && in.Spec.Cluster.IsWorkerCluster {
		// Open cluster for the tenant, create namespace and pvc
		err := cluster.Open(handler.K8sClient, in, tenant)
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

	if in.Spec.Type == api.SonarQube && in.Spec.SonarQube != nil {
		_, err := sonarqube.NewSonar(in.Spec.SonarQube.Server, in.Spec.SonarQube.Token)
		if err != nil {
			return nil, err
		}
	}

	secret, err := integration.ToSecret(tenant, in)
	if err != nil {
		return nil, err
	}

	ns := svrcommon.TenantNamespace(tenant)
	if isPublic {
		ns = common.GetSystemNamespace()
	}
	_, err = handler.K8sClient.CoreV1().Secrets(ns).Create(secret)
	if err != nil {
		log.Errorf("Create secret %v for tenant %s error %v", secret.ObjectMeta.Name, tenant, err)
		return nil, cerr.ConvertK8sError(err)
	}

	return in, nil
}

// GetIntegration gets an integration with the given name under given tenant.
func GetIntegration(ctx context.Context, tenant, name string, isPublic bool) (*api.Integration, error) {
	ns := svrcommon.TenantNamespace(tenant)
	if isPublic {
		ns = common.GetSystemNamespace()
	}

	return getIntegration(ns, name)
}

func getIntegration(namespace, name string) (*api.Integration, error) {
	secret, err := handler.K8sClient.CoreV1().Secrets(namespace).Get(
		integration.GetSecretName(name), meta_v1.GetOptions{})
	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return integration.FromSecret(secret)
}

// UpdateIntegration updates an integration with the given tenant name and integration name.
// If updated successfully, return the updated integration.
func UpdateIntegration(ctx context.Context, tenant, name string, isPublic bool, in *api.Integration) (*api.Integration, error) {
	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil {
		err := cluster.UpdateClusterIntegration(handler.K8sClient, tenant, name, in)
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

	if in.Spec.Type == api.SonarQube && in.Spec.SonarQube != nil {
		_, err := sonarqube.NewSonar(in.Spec.SonarQube.Server, in.Spec.SonarQube.Token)
		if err != nil {
			return nil, err
		}
	}

	secret, err := integration.ToSecret(tenant, in)
	if err != nil {
		return nil, err
	}

	ns := svrcommon.TenantNamespace(tenant)
	if isPublic {
		ns = common.GetSystemNamespace()
	}
	err = updateSecret(ns, integration.GetSecretName(name), in.Spec.Type, secret)
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
		newSecret.Annotations = utils.MergeMap(secret.Annotations, newSecret.Annotations)
		newSecret.Labels = utils.MergeMap(secret.Labels, newSecret.Labels)
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
func DeleteIntegration(ctx context.Context, tenant, name string, isPublic bool) error {
	ns := svrcommon.TenantNamespace(tenant)
	if isPublic {
		ns = common.GetSystemNamespace()
	}

	in, err := getIntegration(ns, name)
	if err != nil {
		return err
	}

	if in.Spec.Type == api.Cluster && in.Spec.Cluster != nil && in.Spec.Cluster.IsWorkerCluster {
		return cerr.ErrorClusterNotClosed.Error(in.Name)
	}

	err = handler.K8sClient.CoreV1().Secrets(ns).Delete(integration.GetSecretName(name), &meta_v1.DeleteOptions{})
	return cerr.ConvertK8sError(err)
}

// OpenCluster opens cluster type integration to execute workflow
func OpenCluster(ctx context.Context, tenant, name string) error {
	ns := svrcommon.TenantNamespace(tenant)
	in, err := getIntegration(ns, name)
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
	err = cluster.Open(handler.K8sClient, in, tenant)
	if err != nil {
		return err
	}

	in.Spec.Cluster.IsWorkerCluster = true
	secret, err := integration.ToSecret(tenant, in)
	if err != nil {
		return err
	}

	return updateSecret(ns, integration.GetSecretName(name), in.Spec.Type, secret)
}

// CloseCluster closes cluster type integration that used to execute workflow
func CloseCluster(ctx context.Context, tenant, name string) error {
	ns := svrcommon.TenantNamespace(tenant)
	in, err := getIntegration(ns, name)
	if err != nil {
		return err
	}

	if in.Spec.Type != api.Cluster {
		return cerr.ErrorIntegrationTypeNotCorrect.Error(name, api.Cluster, in.Spec.Type)
	}

	if in.Spec.Cluster == nil {
		return cerr.ErrorUnknownInternal.Error("cluster is nil")
	}

	if !in.Spec.Cluster.IsWorkerCluster {
		return nil
	}

	// close cluster for the tenant, delete namespace
	err = cluster.Close(handler.K8sClient, in, tenant)
	if err != nil {
		return err
	}

	in.Spec.Cluster.IsWorkerCluster = false
	secret, err := integration.ToSecret(tenant, in)
	if err != nil {
		return err
	}

	return updateSecret(ns, integration.GetSecretName(name), in.Spec.Type, secret)
}

func getSCMSourceFromIntegration(tenant, integrationName string) (*api.SCMSource, error) {
	integration, err := getIntegration(svrcommon.TenantNamespace(tenant), integrationName)
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
