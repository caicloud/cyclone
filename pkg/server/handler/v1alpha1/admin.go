package v1alpha1

import (
	"context"
	"sort"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/handler/v1alpha1/sorter"
	"github.com/caicloud/cyclone/pkg/server/types"
)

// AllWorkflows lists all workflows matched the provided label from all projects and tenants.
func AllWorkflows(ctx context.Context, label string, query *types.QueryParams) (*types.ListResponse, error) {
	workflows, err := handler.K8sClient.CycloneV1alpha1().Workflows("").List(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Errorf("Get workflows from k8s error: %v", err)
		return nil, err
	}

	items := workflows.Items
	size := int64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Workflow{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[query.Start:end]), nil
}

// AllStages lists all stages matched the provided label from all projects and tenants.
func AllStages(ctx context.Context, label string, query *types.QueryParams) (*types.ListResponse, error) {
	stages, err := handler.K8sClient.CycloneV1alpha1().Stages("").List(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Errorf("Get workflows from k8s error: %v", err)
		return nil, err
	}

	items := stages.Items
	size := int64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Stage{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[query.Start:end]), nil
}

// AllResources lists all resources matched the provided label from all projects and tenants.
func AllResources(ctx context.Context, label string, query *types.QueryParams) (*types.ListResponse, error) {
	resources, err := handler.K8sClient.CycloneV1alpha1().Resources("").List(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Errorf("Get workflows from k8s error: %v", err)
		return nil, err
	}

	items := resources.Items
	size := int64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Resource{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[query.Start:end]), nil
}

// AllWorkflowRuns lists all workflowruns matched the provided label from all projects and tenants.
func AllWorkflowRuns(ctx context.Context, label string, query *types.QueryParams) (*types.ListResponse, error) {
	workflowruns, err := handler.K8sClient.CycloneV1alpha1().WorkflowRuns("").List(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Errorf("Get workflows from k8s error: %v", err)
		return nil, err
	}

	items := workflowruns.Items
	sort.Sort(sorter.NewWorkflowRunSorter(items, false))
	size := int64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.WorkflowRun{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[query.Start:end]), nil
}

// AllWorkflowTriggers lists all workflow triggers matched the provided label from all projects and tenants.
func AllWorkflowTriggers(ctx context.Context, label string, query *types.QueryParams) (*types.ListResponse, error) {
	triggers, err := handler.K8sClient.CycloneV1alpha1().WorkflowTriggers("").List(metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		log.Errorf("Get workflows from k8s error: %v", err)
		return nil, err
	}

	items := triggers.Items
	size := int64(len(items))
	if query.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.WorkflowTrigger{}), nil
	}

	end := query.Start + query.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[query.Start:end]), nil
}
