package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// POST /apis/v1alpha1/resources/
// X-Tenant: any
func CreateResource(ctx context.Context) (*v1alpha1.Resource, error) {
	rs := &v1alpha1.Resource{}
	err := contextutil.GetJsonPayload(ctx, rs)
	if err != nil {
		return nil, err
	}

	return k8sClient.CycloneV1alpha1().Resources(rs.Namespace).Create(rs)
}

// POST /apis/v1alpha1/resources/{resource-name}
// X-Tenant: any
func GetResource(ctx context.Context, name, namespace string) (*v1alpha1.Resource, error) {
	return k8sClient.CycloneV1alpha1().Resources(namespace).Get(name, metav1.GetOptions{})
}
