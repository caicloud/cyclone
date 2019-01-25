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
	wft.ObjectMeta.Labels = common.AddProjectLabel(wft.ObjectMeta.Labels, project)
	wft.Name = common.BuildResoucesName(project, wft.Name)

	created, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Create(wft)
	if err != nil {
		return nil, err
	}

	created.Name = common.RetrieveResoucesName(project, wft.Name)
	return created, nil
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
		return types.NewListResponse(int(size), []v1alpha1.Stage{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	wfts := make([]v1alpha1.WorkflowTrigger, size)
	for i, wft := range items[pagination.Start:end] {
		wft.Name = common.RetrieveResoucesName(project, wft.Name)
		wfts[i] = wft
	}
	return types.NewListResponse(int(size), wfts), nil
}

// GetWorkflowTrigger ...
func GetWorkflowTrigger(ctx context.Context, project, workflowtrigger, tenant string) (*v1alpha1.WorkflowTrigger, error) {
	name := common.BuildResoucesName(project, workflowtrigger)
	wft, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	wft.Name = common.RetrieveResoucesName(project, wft.Name)
	return wft, nil
}

// UpdateWorkflowTrigger ...
func UpdateWorkflowTrigger(ctx context.Context, project, workflowtrigger, tenant string, wft *v1alpha1.WorkflowTrigger) (*v1alpha1.WorkflowTrigger, error) {
	name := common.BuildResoucesName(project, workflowtrigger)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newWft := origin.DeepCopy()
		newWft.Spec = wft.Spec
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
	name := common.BuildResoucesName(project, workflowtrigger)
	return handler.K8sClient.CycloneV1alpha1().WorkflowTriggers(common.TenantNamespace(tenant)).Delete(name, nil)
}
