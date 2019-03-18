package v1alpha1

import (
	"context"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/statistic"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// CreateWorkflow ...
func CreateWorkflow(ctx context.Context, tenant, project string, wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, project, "", wf)
		if err != nil {
			return nil, err
		}
	}

	return handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Create(wf)
}

// ListWorkflows ...
func ListWorkflows(ctx context.Context, tenant, project string, pagination *types.Pagination) (*types.ListResponse, error) {
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
		return types.NewListResponse(int(size), []v1alpha1.Workflow{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}

// GetWorkflow ...
func GetWorkflow(ctx context.Context, tenant, project, workflow string) (*v1alpha1.Workflow, error) {
	wf, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Get(workflow, metav1.GetOptions{})

	return wf, cerr.ConvertK8sError(err)
}

// UpdateWorkflow ...
func UpdateWorkflow(ctx context.Context, tenant, project, workflow string, wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Get(workflow, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newWf := origin.DeepCopy()
		newWf.Spec = wf.Spec
		newWf.Annotations = MergeMap(wf.Annotations, newWf.Annotations)
		newWf.Labels = MergeMap(wf.Labels, newWf.Labels)
		_, err = handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Update(newWf)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return wf, nil
}

// DeleteWorkflow ...
func DeleteWorkflow(ctx context.Context, tenant, project, workflow string) error {
	err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Delete(workflow, nil)

	return cerr.ConvertK8sError(err)

}

// GetWFStatistics handles the request to get a workflow's statistics.
func GetWFStatistics(ctx context.Context, tenant, project, workflow string, start, end string) (*api.StatusStats, error) {
	wfrs, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: common.ProjectSelector(project) + "," + common.WorkflowSelector(workflow),
	})
	if err != nil {
		return nil, err
	}

	return statistic.Stats(wfrs, start, end)
}
