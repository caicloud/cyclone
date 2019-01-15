package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/handler/common"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// CreateWorkflowTrigger ...
func CreateWorkflowTrigger(ctx context.Context) (*v1alpha1.WorkflowTrigger, error) {
	wft := &v1alpha1.WorkflowTrigger{}
	err := contextutil.GetJSONPayload(ctx, wft)
	if err != nil {
		return nil, err
	}

	return common.K8sClient.CycloneV1alpha1().WorkflowTriggers(wft.Namespace).Create(wft)
}

// ListWorkflowTriggers ...
func ListWorkflowTriggers(ctx context.Context, namespace string) (*v1alpha1.WorkflowTriggerList, error) {
	return common.K8sClient.CycloneV1alpha1().WorkflowTriggers(namespace).List(metav1.ListOptions{})
}

// GetWorkflowTrigger ...
func GetWorkflowTrigger(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowTrigger, error) {
	return common.K8sClient.CycloneV1alpha1().WorkflowTriggers(namespace).Get(name, metav1.GetOptions{})
}

// UpdateWorkflowTrigger ...
func UpdateWorkflowTrigger(ctx context.Context, name string) (*v1alpha1.WorkflowTrigger, error) {
	wft := &v1alpha1.WorkflowTrigger{}
	err := contextutil.GetJSONPayload(ctx, wft)
	if err != nil {
		return nil, err
	}

	if name != wft.Name {
		return nil, cerr.ErrorValidationFailed.Error("Name", "WorkflowTrigger name inconsistent between body and path.")
	}

	return common.K8sClient.CycloneV1alpha1().WorkflowTriggers(wft.Namespace).Update(wft)
}

// DeleteWorkflowTrigger ...
func DeleteWorkflowTrigger(ctx context.Context, name, namespace string) error {
	return common.K8sClient.CycloneV1alpha1().WorkflowTriggers(namespace).Delete(name, nil)
}
