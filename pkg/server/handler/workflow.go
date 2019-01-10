package handler

import (
	"context"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// CreateWorkflow ... POST /apis/v1alpha1/workflows/
func CreateWorkflow(ctx context.Context) (*v1alpha1.Workflow, error) {
	wf := &v1alpha1.Workflow{}
	err := contextutil.GetJSONPayload(ctx, wf)
	if err != nil {
		return nil, err
	}

	wc, err := k8sClient.CycloneV1alpha1().Workflows(wf.Namespace).Create(wf)
	if err != nil {
		log.Errorf("Create workflow %s error:%v", wf.Name, err)
	}
	return wc, nil
}

// ListWorkflows ... GET /apis/v1alpha1/workflows/
func ListWorkflows(ctx context.Context, namespace string) (*v1alpha1.WorkflowList, error) {
	return k8sClient.CycloneV1alpha1().Workflows(namespace).List(metav1.ListOptions{})
}

// GetWorkflow ... GET /apis/v1alpha1/workflows/{workflow}
func GetWorkflow(ctx context.Context, name, namespace string) (*v1alpha1.Workflow, error) {
	return k8sClient.CycloneV1alpha1().Workflows(namespace).Get(name, metav1.GetOptions{})
}

// UpdateWorkflow ... PUT /apis/v1alpha1/workflows/{workflow}
func UpdateWorkflow(ctx context.Context, name string) (*v1alpha1.Workflow, error) {
	wf := &v1alpha1.Workflow{}
	err := contextutil.GetJSONPayload(ctx, wf)
	if err != nil {
		return nil, err
	}

	if name != wf.Name {
		return nil, cerr.ErrorValidationFailed.Error("Name", "Workflow name inconsistent between body and path.")
	}

	return k8sClient.CycloneV1alpha1().Workflows(wf.Namespace).Update(wf)
}

// DeleteWorkflow ... DELETE /apis/v1alpha1/workflows/{workflow}
func DeleteWorkflow(ctx context.Context, name, namespace string) error {
	return k8sClient.CycloneV1alpha1().Workflows(namespace).Delete(name, nil)
}
