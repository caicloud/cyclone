package handler

import (
	"context"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// POST /apis/v1alpha1/workflows/
// X-Tenant: any
func CreateWorkflow(ctx context.Context) (*v1alpha1.Workflow, error) {
	wf := &v1alpha1.Workflow{}
	err := contextutil.GetJsonPayload(ctx, wf)
	if err != nil {
		return nil, err
	}

	wc, err := k8sClient.CycloneV1alpha1().Workflows(wf.Namespace).Create(wf)
	if err != nil {
		log.Errorf("Create workflow %s error:%v", wf.Name, err)
	}
	return wc, nil
}

// POST /apis/v1alpha1/workflows/{workflow-name}
// X-Tenant: any
func GetWorkflow(ctx context.Context, name, namespace string) (*v1alpha1.Workflow, error) {
	if namespace == "" {
		namespace = "default"
	}
	return k8sClient.CycloneV1alpha1().Workflows(namespace).Get(name, metav1.GetOptions{})
}
