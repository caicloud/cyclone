package handler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// POST /apis/v1alpha1/resources/
func CreateResource(ctx context.Context) (*v1alpha1.Resource, error) {
	rs := &v1alpha1.Resource{}
	err := contextutil.GetJsonPayload(ctx, rs)
	if err != nil {
		return nil, err
	}

	return k8sClient.CycloneV1alpha1().Resources(rs.Namespace).Create(rs)
}

// GET /apis/v1alpha1/resources/
func ListResources(ctx context.Context, namespace string) (*v1alpha1.ResourceList, error) {
	return k8sClient.CycloneV1alpha1().Resources(namespace).List(metav1.ListOptions{})
}

// GET /apis/v1alpha1/resources/{resource}
func GetResource(ctx context.Context, name, namespace string) (*v1alpha1.Resource, error) {
	return k8sClient.CycloneV1alpha1().Resources(namespace).Get(name, metav1.GetOptions{})
}

// PUT /apis/v1alpha1/resources/{resource}
func UpdateResource(ctx context.Context, name string) (*v1alpha1.Resource, error) {
	rs := &v1alpha1.Resource{}
	err := contextutil.GetJsonPayload(ctx, rs)
	if err != nil {
		return nil, err
	}

	if name != rs.Name {
		return nil, cerr.ErrorValidationFailed.Error("Name", "Resource name inconsistent between body and path.")
	}

	return k8sClient.CycloneV1alpha1().Resources(rs.Namespace).Update(rs)
}

// DELETE /apis/v1alpha1/resources/{resource}
func DeleteResource(ctx context.Context, name, namespace string) error {
	return k8sClient.CycloneV1alpha1().Resources(namespace).Delete(name, nil)
}
