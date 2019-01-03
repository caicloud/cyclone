package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// POST /api/v1alpha1/workflowtriggers
func CreateWorkflowTrigger(ctx context.Context) (*v1alpha1.WorkflowTrigger, error) {
	wft := &v1alpha1.WorkflowTrigger{}
	err := contextutil.GetJsonPayload(ctx, wft)
	if err != nil {
		return nil, err
	}

	return k8sClient.CycloneV1alpha1().WorkflowTriggers(wft.Namespace).Create(wft)
}

// GET /apis/v1alpha1/workflowtriggers/
func ListWorkflowTriggers(ctx context.Context, namespace string) (*v1alpha1.WorkflowTriggerList, error) {
	return k8sClient.CycloneV1alpha1().WorkflowTriggers(namespace).List(metav1.ListOptions{})
}

// GET /apis/v1alpha1/workflowtriggers/{workflowtrigger}
func GetWorkflowTrigger(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowTrigger, error) {
	return k8sClient.CycloneV1alpha1().WorkflowTriggers(namespace).Get(name, metav1.GetOptions{})
}

// PUT /apis/v1alpha1/workflowtriggers/{workflowtrigger}
func UpdateWorkflowTrigger(ctx context.Context, name string) (*v1alpha1.WorkflowTrigger, error) {
	wft := &v1alpha1.WorkflowTrigger{}
	err := contextutil.GetJsonPayload(ctx, wft)
	if err != nil {
		return nil, err
	}

	if name != wft.Name {
		return nil, cerr.ErrorValidationFailed.Error("Name", "WorkflowTrigger name inconsistent between body and path.")
	}

	return k8sClient.CycloneV1alpha1().WorkflowTriggers(wft.Namespace).Update(wft)
}

// DELETE /apis/v1alpha1/workflowtriggers/{workflowtrigger}
func DeleteWorkflowTrigger(ctx context.Context, name, namespace string) error {
	return k8sClient.CycloneV1alpha1().WorkflowTriggers(namespace).Delete(name, nil)
}
