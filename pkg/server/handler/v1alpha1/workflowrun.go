package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// CreateWorkflowRun ...
func CreateWorkflowRun(ctx context.Context) (*v1alpha1.WorkflowRun, error) {
	wfr := &v1alpha1.WorkflowRun{}
	err := contextutil.GetJSONPayload(ctx, wfr)
	if err != nil {
		return nil, err
	}

	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Create(wfr)
}

// ListWorkflowRuns ...
func ListWorkflowRuns(ctx context.Context, namespace string) (*v1alpha1.WorkflowRunList, error) {
	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(namespace).List(metav1.ListOptions{})
}

// GetWorkflowRun ...
func GetWorkflowRun(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowRun, error) {
	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Get(name, metav1.GetOptions{})
}

// UpdateWorkflowRun ...
func UpdateWorkflowRun(ctx context.Context, name string) (*v1alpha1.WorkflowRun, error) {
	wfr := &v1alpha1.WorkflowRun{}
	err := contextutil.GetJSONPayload(ctx, wfr)
	if err != nil {
		return nil, err
	}

	if name != wfr.Name {
		return nil, cerr.ErrorValidationFailed.Error("Name", "WorkflowRun name inconsistent between body and path.")
	}

	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(wfr)
}

// DeleteWorkflowRun ...
func DeleteWorkflowRun(ctx context.Context, name, namespace string) error {
	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Delete(name, nil)
}

// CancelWorkflowRun updates the workflowrun overall status to Cancelled.
func CancelWorkflowRun(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowRun, error) {
	data := `[{"op":"replace","path":"/status/overall/status","value":"Cancelled"}]`
	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Patch(name, types.JSONPatchType, []byte(data))
}

// ContinueWorkflowRun updates the workflowrun overall status to Running.
func ContinueWorkflowRun(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowRun, error) {
	data := `[{"op":"replace","path":"/status/overall/status","value":"Running"}]`
	return handler.K8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Patch(name, types.JSONPatchType, []byte(data))
}
