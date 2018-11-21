package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
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

// POST /apis/v1alpha1/workflowruns/{workflowrun-name}
// X-Tenant: any
func GetWorkflowRun(ctx context.Context, name, namespace string) (*v1alpha1.WorkflowRun, error) {
	if namespace == "" {
		namespace = "default"
	}
	return k8sClient.CycloneV1alpha1().WorkflowRuns(namespace).Get(name, metav1.GetOptions{})
}
