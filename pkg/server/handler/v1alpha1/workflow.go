package v1alpha1

import (
	"context"
	"sort"
	"strings"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/meta"
	api "github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/statistic"
	"github.com/caicloud/cyclone/pkg/server/biz/utils"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// CreateWorkflow ...
func CreateWorkflow(ctx context.Context, tenant, project string, wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier, InjectProjectOwnerRefModifier}
	for _, modifier := range modifiers {
		err := modifier(tenant, project, "", wf)
		if err != nil {
			log.Errorf("Failed to create workflow %s for tenant %s in project %s as error: %v", wf.Name, tenant, project, err)
			return nil, err
		}
	}

	return handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Create(wf)
}

// ListWorkflows ...
func ListWorkflows(ctx context.Context, tenant, project string, query *types.QueryParams) (*types.ListResponse, error) {
	workflows, err := handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: meta.ProjectSelector(project),
	})
	if err != nil {
		log.Errorf("Get workflows from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := workflows.Items
	var results []v1alpha1.Workflow
	if query.Filter == "" {
		results = items
	} else {
		// Only support filter by name or alias.
		kv := strings.Split(query.Filter, "=")
		if len(kv) != 2 {
			return nil, cerr.ErrorQueryParamNotCorrect.Error(query.Filter)
		}
		value := strings.ToLower(kv[1])

		if kv[0] == "name" {
			for _, item := range items {
				if strings.Contains(item.Name, value) {
					results = append(results, item)
				}
			}
		} else if kv[0] == "alias" {
			for _, item := range items {
				if item.Annotations != nil {
					if alias, ok := item.Annotations[meta.AnnotationAlias]; ok {
						if strings.Contains(alias, value) {
							results = append(results, item)
						}
					}
				}
			}
		} else {
			// Will not filter results.
			results = items
		}
	}

	size := uint64(len(results))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Workflow{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	if query.Sort {
		sort.Sort(sorter.NewWorkflowSorter(results, query.Ascending))
	}

	return types.NewListResponse(int(size), results[query.Start:end]), nil
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
		newWf.Annotations = utils.MergeMap(wf.Annotations, newWf.Annotations)
		newWf.Labels = utils.MergeMap(wf.Labels, newWf.Labels)
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
	err := deleteCollections(tenant, project, workflow)
	if err != nil {
		return err
	}

	err = handler.K8sClient.CycloneV1alpha1().Workflows(common.TenantNamespace(tenant)).Delete(workflow, nil)
	return cerr.ConvertK8sError(err)

}

// GetWFStatistics handles the request to get a workflow's statistics.
func GetWFStatistics(ctx context.Context, tenant, project, workflow string, start, end string) (*api.Statistic, error) {
	wfrs, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: meta.ProjectSelector(project) + "," + meta.WorkflowSelector(workflow),
	})
	if err != nil {
		return nil, err
	}

	return statistic.Stats(wfrs, start, end)
}
