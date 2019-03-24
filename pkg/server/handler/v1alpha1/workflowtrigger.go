package v1alpha1

import (
	"context"
	"encoding/json"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// CreateWorkflowTrigger ...
func CreateWorkflowTrigger(ctx context.Context, project, tenant string, wft *v1alpha1.WorkflowTrigger) (*v1alpha1.WorkflowTrigger, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, project, "", wft)
		if err != nil {
			return nil, err
		}
	}

	if wft.Spec.Type == v1alpha1.TriggerTypeSCM {
		// Create webhook for integrated SCM.
		SCMTrigger := wft.Spec.SCM
		if err := registerSCMWebhook(tenant, wft.Name, SCMTrigger.Secret, SCMTrigger.Repo); err != nil {
			return nil, err
		}
	}

	return handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Create(wft)
}

// ListWorkflowTriggers ...
func ListWorkflowTriggers(ctx context.Context, project, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	workflowTriggers, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: common.ProjectSelector(project),
	})
	if err != nil {
		log.Errorf("Get workflowtrigger from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := workflowTriggers.Items
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.WorkflowTrigger{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}

// GetWorkflowTrigger ...
func GetWorkflowTrigger(ctx context.Context, project, workflowtrigger, tenant string) (*v1alpha1.WorkflowTrigger, error) {
	wft, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(workflowtrigger, metav1.GetOptions{})

	return wft, cerr.ConvertK8sError(err)
}

// UpdateWorkflowTrigger ...
func UpdateWorkflowTrigger(ctx context.Context, project, workflowtrigger, tenant string, wft *v1alpha1.WorkflowTrigger) (*v1alpha1.WorkflowTrigger, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(workflowtrigger, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newWft := origin.DeepCopy()
		newWft.Spec = wft.Spec
		newWft.Annotations = MergeMap(wft.Annotations, newWft.Annotations)
		newWft.Labels = MergeMap(wft.Labels, newWft.Labels)

		// Handle trigger type change and repo change when SCM type.
		// Do not care about the change of secret.
		oldSpec := origin.Spec
		newSpec := wft.Spec
		unregisterOld, registerNew := false, false
		if oldSpec.Type == v1alpha1.TriggerTypeSCM {
			// Need to unregister old SCM webhook only when:
			// * new trigger is not SCM type
			// * repo of new trigger is different from old
			if newSpec.Type == v1alpha1.TriggerTypeCron {
				unregisterOld = true
			} else if oldSpec.SCM.Repo != newSpec.SCM.Repo {
				unregisterOld = true
				registerNew = true
			}
		} else if newSpec.Type == v1alpha1.TriggerTypeSCM {
			registerNew = true
		}

		if unregisterOld {
			if err = unregisterSCMWebhook(tenant, wft.Name, oldSpec.SCM.Secret, oldSpec.SCM.Repo); err != nil {
				return err
			}
		}

		if registerNew {
			if err = registerSCMWebhook(tenant, wft.Name, newSpec.SCM.Secret, newSpec.SCM.Repo); err != nil {
				return err
			}
		}

		_, err = handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Update(newWft)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return wft, nil
}

// DeleteWorkflowTrigger ...
func DeleteWorkflowTrigger(ctx context.Context, project, workflowtrigger, tenant string) error {
	wft, err := GetWorkflowTrigger(ctx, project, workflowtrigger, tenant)
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	if wft.Spec.Type == v1alpha1.TriggerTypeSCM {
		// Unregister webhook for integrated SCM.
		SCMTrigger := wft.Spec.SCM
		if err := unregisterSCMWebhook(tenant, wft.Name, SCMTrigger.Secret, SCMTrigger.Repo); err != nil {
			return err
		}
	}

	err = handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Delete(workflowtrigger, nil)

	return cerr.ConvertK8sError(err)
}

func registerSCMWebhook(tenant, wftName, secretName, repo string) error {
	secret, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(
		common.IntegrationSecret(secretName), metav1.GetOptions{})
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	repos := map[string][]string{}
	if d, ok := secret.Data[common.SecretKeyRepos]; ok {
		log.Infof("repos data of secret %s: %s\n", secretName, d)
		if err = json.Unmarshal(d, &repos); err != nil {
			log.Errorf("Failed to unmarshal repos from secret")
			return err
		}
	}

	found := false
	wfts, ok := repos[repo]
	if ok {
		for _, wft := range wfts {
			if wft == wftName {
				found = true
				break
			}
		}
		if !found {
			wfts = append(wfts, wftName)
			repos[repo] = wfts
		}
	} else {
		// Create webhook for this repo.
		log.Infof("Create webhook for repo %s", repo)
		integration, err := SecretToIntegration(secret)
		if err != nil {
			log.Error(err)
			return err
		}

		err = CreateSCMWebhook(integration.Spec.SCM, tenant, secretName, repo)
		if err != nil {
			log.Error(err)
			return err
		}

		repos[repo] = []string{wftName}
	}

	if found {
		log.Warningf("Found wft %s for repo %s in secret %s", wftName, repo, secretName)
		return nil
	}

	reposStr, err := json.Marshal(repos)
	if err != nil {
		log.Errorf("Failed to marshal repos for secret")
		return err
	}

	secret.Data[common.SecretKeyRepos] = reposStr

	secret, err = handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Update(secret)
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	return nil
}

func unregisterSCMWebhook(tenant, wftName, secretName, repo string) error {
	secret, err := handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Get(
		common.IntegrationSecret(secretName), metav1.GetOptions{})
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	repos := map[string][]string{}
	if d, ok := secret.Data[common.SecretKeyRepos]; ok {
		log.Infof("repos data of secret %s: %s\n", secretName, d)
		if err = json.Unmarshal(d, &repos); err != nil {
			log.Errorf("Failed to unmarshal repos from secret")
			return err
		}
	}

	found := false
	wfts, ok := repos[repo]
	if ok {
		if len(wfts) == 1 {
			if wfts[0] == wftName {
				found = true
				// Delete webhook for repo.
				log.Infof("Delete webhook for repo %s", repo)
				integration, err := SecretToIntegration(secret)
				if err != nil {
					log.Error(err)
					return err
				}

				err = DeleteSCMWebhook(integration.Spec.SCM, tenant, secretName, repo)
				if err != nil {
					log.Error(err)
					return err
				}

				delete(repos, repo)
			}
		} else {
			for i, wft := range wfts {
				if wft == wftName {
					found = true
					wfts = append(wfts[:i], wfts[i+1:]...)
					break
				}
			}

			if found {
				repos[repo] = wfts
			}
		}

	}

	if !found {
		log.Warningf("Not found wft %s for repo %s in secret %s", wftName, repo, secretName)
		return nil
	}

	reposStr, err := json.Marshal(repos)
	if err != nil {
		log.Errorf("Failed to marshal repos for secret")
		return err
	}

	secret.Data[common.SecretKeyRepos] = reposStr
	secret, err = handler.K8sClient.CoreV1().Secrets(common.TenantNamespace(tenant)).Update(secret)
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	return nil
}
