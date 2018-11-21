package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
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

// POST /apis/v1alpha1/workflowtriggers/{workflowtrigger-name}
// X-Tenant: any
func GetWorkflowTrigger(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowTrigger, error) {
	return k8sClient.CycloneV1alpha1().WorkflowTriggers(namespace).Get(name, metav1.GetOptions{})
}
