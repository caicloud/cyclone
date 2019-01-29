package v1alpha1

import (
	"context"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
)

// CreateWorkflowTrigger ...
func CreateWorkflowTrigger(ctx context.Context, project, tenant string, wft *v1alpha1.WorkflowTrigger) (*v1alpha1.WorkflowTrigger, error) {
	err := ModifyResource(project, tenant, wft)
	if err != nil {
		return nil, err
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
	return handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(workflowtrigger, metav1.GetOptions{})
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
		newWft.Annotations = UpdateAnnotations(wft.Annotations, newWft.Annotations)
		_, err = handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Update(newWft)
		return err
	})

	if err != nil {
		return nil, err
	}

	return wft, nil
}

// DeleteWorkflowTrigger ...
func DeleteWorkflowTrigger(ctx context.Context, project, workflowtrigger, tenant string) error {
	return handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Delete(workflowtrigger, nil)
}
