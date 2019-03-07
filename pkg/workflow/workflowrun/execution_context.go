package workflowrun

import (
	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

// GetExecutionContext gets execution context from WorkflowRun, if not found, use the
// default context in workflow controller configuration.
func GetExecutionContext(wfr *v1alpha1.WorkflowRun) *v1alpha1.ExecutionContext {
	if wfr.Spec.ExecutionContext != nil {
		return wfr.Spec.ExecutionContext
	}

	return &v1alpha1.ExecutionContext{
		Namespace: controller.Config.ExecutionContext.Namespace,
		PVC:       controller.Config.ExecutionContext.PVC,
	}
}
