package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// POST /api/v1alpha1/workflowruns
func CreateWorkflowRun(ctx context.Context) (*v1alpha1.WorkflowRun, error) {
	wfr := &v1alpha1.WorkflowRun{}
	err := contextutil.GetJsonPayload(ctx, wfr)
	if err != nil {
		return nil, err
	}

	return k8sClient.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Create(wfr)
}

// GET /apis/v1alpha1/workflowruns/
func ListWorkflowRuns(ctx context.Context, namespace string) (*v1alpha1.WorkflowRunList, error) {
	return k8sClient.CycloneV1alpha1().WorkflowRuns(namespace).List(metav1.ListOptions{})
}

// GET /apis/v1alpha1/workflowruns/{workflowrun}
func GetWorkflowRun(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowRun, error) {
	return k8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Get(name, metav1.GetOptions{})
}

// PUT /apis/v1alpha1/workflowruns/{workflowrun}
func UpdateWorkflowRun(ctx context.Context, name string) (*v1alpha1.WorkflowRun, error) {
	wfr := &v1alpha1.WorkflowRun{}
	err := contextutil.GetJsonPayload(ctx, wfr)
	if err != nil {
		return nil, err
	}

	if name != wfr.Name {
		return nil, cerr.ErrorValidationFailed.Error("Name", "WorkflowRun name inconsistent between body and path.")
	}

	return k8sClient.CycloneV1alpha1().WorkflowRuns(wfr.Namespace).Update(wfr)
}

// DELETE /apis/v1alpha1/workflowruns/{workflowrun}
func DeleteWorkflowRun(ctx context.Context, name, namespace string) error {
	return k8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Delete(name, nil)
}

// CancelWorkflowRun updates the workflowrun overall status to Cancelled.
func CancelWorkflowRun(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowRun, error) {
	data := `[{"op":"replace","path":"/status/overall/status","value":"Cancelled"}]`
	return k8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Patch(name, types.JSONPatchType, []byte(data))
}

// CancelWorkflowRun updates the workflowrun overall status to Running.
func ContinueWorkflowRun(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowRun, error) {
	data := `[{"op":"replace","path":"/status/overall/status","value":"Running"}]`
	return k8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Patch(name, types.JSONPatchType, []byte(data))
}
