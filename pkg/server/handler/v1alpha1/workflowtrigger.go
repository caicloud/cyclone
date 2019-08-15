package v1alpha1

import (
	"context"
	"sort"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/server/biz/hook"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// CreateWorkflowTrigger ...
func CreateWorkflowTrigger(ctx context.Context, tenant, project, workflow string, wft *v1alpha1.WorkflowTrigger) (*v1alpha1.WorkflowTrigger, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier, InjectWorkflowLabelModifier, InjectWorkflowOwnerRefModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, project, workflow, wft)
		if err != nil {
			return nil, err
		}
	}

	if wft.Spec.WorkflowRef == nil {
		wft.Spec.WorkflowRef = workflowReference(tenant, workflow)
	}

	if wft.Spec.Type == v1alpha1.TriggerTypeWebhook || wft.Spec.Type == v1alpha1.TriggerTypeSCM {
		hookManager, err := hook.GetManager(wft.Spec.Type)
		if err != nil {
			return nil, err
		}

		err = hookManager.Register(tenant, *wft)
		if err != nil {
			return nil, err
		}
	}

	hook.LabelSCMTrigger(wft)
	return handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Create(wft)
}

// ListWorkflowTriggers ...
func ListWorkflowTriggers(ctx context.Context, tenant, project, workflow string, query *types.QueryParams) (*types.ListResponse, error) {
	workflowTriggers, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: meta.ProjectSelector(project) + "," + meta.WorkflowSelector(workflow),
	})
	if err != nil {
		log.Errorf("Get workflowtrigger from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := workflowTriggers.Items

	size := uint64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.WorkflowTrigger{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	if query.Sort {
		sort.Sort(sorter.NewWorkflowTriggerSorter(items, query.Ascending))
	}

	return types.NewListResponse(int(size), items[query.Start:end]), nil
}

// GetWorkflowTrigger ...
func GetWorkflowTrigger(ctx context.Context, tenant, project, workflow, workflowtrigger string) (*v1alpha1.WorkflowTrigger, error) {
	wft, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(workflowtrigger, metav1.GetOptions{})
	return wft, cerr.ConvertK8sError(err)
}

// UpdateWorkflowTrigger ...
func UpdateWorkflowTrigger(ctx context.Context, tenant, project, workflow, workflowtrigger string, wft *v1alpha1.WorkflowTrigger) (*v1alpha1.WorkflowTrigger, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(workflowtrigger, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newWft := origin.DeepCopy()
		newWft.Spec = wft.Spec
		newWft.Annotations = utils.MergeMap(wft.Annotations, newWft.Annotations)
		newWft.Labels = utils.MergeMap(wft.Labels, newWft.Labels)
		if newWft.Spec.WorkflowRef == nil {
			newWft.Spec.WorkflowRef = workflowReference(tenant, workflow)
		}

		// Handle trigger type change and repo change when SCM type.
		// Do not care about the change of secret.
		oldSpec := origin.Spec
		newSpec := newWft.Spec
		unregisterOld, registerNew := false, false
		if oldSpec.Type == v1alpha1.TriggerTypeSCM {
			// Need to unregister old SCM webhook only when:
			// * new trigger is not SCM type
			// * repo of new trigger is different from old
			if newSpec.Type != v1alpha1.TriggerTypeSCM {
				unregisterOld = true
			} else if oldSpec.SCM.Repo != newSpec.SCM.Repo {
				unregisterOld = true
				registerNew = true
			}
		} else if newSpec.Type == v1alpha1.TriggerTypeSCM {
			registerNew = true
		}

		var hookManager hook.Manager
		if registerNew || unregisterOld {
			hookManager, err = hook.GetManager(wft.Spec.Type)
			if err != nil {
				return err
			}
		}

		if unregisterOld {
			if err = hookManager.Unregister(tenant, *origin); err != nil {
				return err
			}
		}

		if registerNew {
			if err = hookManager.Register(tenant, *newWft); err != nil {
				return err
			}
		}

		hook.LabelSCMTrigger(newWft)
		_, err = handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Update(newWft)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return wft, nil
}

// DeleteWorkflowTrigger ...
func DeleteWorkflowTrigger(ctx context.Context, tenant, project, workflow, workflowtrigger string) error {
	wft, err := GetWorkflowTrigger(ctx, tenant, project, workflow, workflowtrigger)
	if err != nil {
		return cerr.ConvertK8sError(err)
	}

	if wft.Spec.Type == v1alpha1.TriggerTypeWebhook || wft.Spec.Type == v1alpha1.TriggerTypeSCM {
		var hookManager hook.Manager
		hookManager, err = hook.GetManager(wft.Spec.Type)
		if err != nil {
			return err
		}

		if err = hookManager.Unregister(tenant, *wft); err != nil {
			return err
		}

		err = handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Delete(workflowtrigger, nil)
	}

	return cerr.ConvertK8sError(err)
}
