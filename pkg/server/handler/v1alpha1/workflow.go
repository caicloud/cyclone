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

// CreateWorkflow ...
func CreateWorkflow(ctx context.Context, project, tenant string, wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	wf.ObjectMeta.Labels = common.AddProjectLabel(wf.ObjectMeta.Labels, project)
	wf.Name = common.BuildResoucesName(project, wf.Name)

	created, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Create(wf)
	if err != nil {
		return nil, err
	}

	created.Name = common.RetrieveResoucesName(project, wf.Name)
	return created, nil
}

// ListWorkflows ...
func ListWorkflows(ctx context.Context, project, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	workflows, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: common.ProjectSelector(project),
	})
	if err != nil {
		log.Errorf("Get workflows from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := workflows.Items
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Stage{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	wfs := make([]v1alpha1.Workflow, size)
	for i, wf := range items[pagination.Start:end] {
		wf.Name = common.RetrieveResoucesName(project, wf.Name)
		wfs[i] = wf
	}
	return types.NewListResponse(int(size), wfs), nil
}

// GetWorkflow ...
func GetWorkflow(ctx context.Context, project, workflow, tenant string) (*v1alpha1.Workflow, error) {
	name := common.BuildResoucesName(project, workflow)
	wf, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	wf.Name = common.RetrieveResoucesName(project, wf.Name)
	return wf, nil
}

// UpdateWorkflow ...
func UpdateWorkflow(ctx context.Context, project, resource, tenant string, wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	name := common.BuildResoucesName(project, resource)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newWf := origin.DeepCopy()
		newWf.Spec = wf.Spec
		_, err = handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Update(newWf)
		return err
	})

	if err != nil {
		return nil, err
	}

	return wf, nil
}

// DeleteWorkflow ...
func DeleteWorkflow(ctx context.Context, project, resource, tenant string) error {
	name := common.BuildResoucesName(project, resource)
	return handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Delete(name, nil)
}
