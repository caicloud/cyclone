package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/handler/common"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	contextutil "github.com/caicloud/cyclone/pkg/util/context"
)

// CreateResource ...
func CreateResource(ctx context.Context) (*v1alpha1.Resource, error) {
	rs := &v1alpha1.Resource{}
	err := contextutil.GetJSONPayload(ctx, rs)
	if err != nil {
		return nil, err
	}

	return common.K8sClient.CycloneV1alpha1().Resources(rs.Namespace).Create(rs)
}

// ListResources ...
func ListResources(ctx context.Context, namespace string) (*v1alpha1.ResourceList, error) {
	return common.K8sClient.CycloneV1alpha1().Resources(namespace).List(metav1.ListOptions{})
}

// GetResource ...
func GetResource(ctx context.Context, name, namespace string) (*v1alpha1.Resource, error) {
	return common.K8sClient.CycloneV1alpha1().Resources(namespace).Get(name, metav1.GetOptions{})
}

// UpdateResource ...
func UpdateResource(ctx context.Context, name string) (*v1alpha1.Resource, error) {
	rs := &v1alpha1.Resource{}
	err := contextutil.GetJSONPayload(ctx, rs)
	if err != nil {
		return nil, err
	}

	if name != rs.Name {
		return nil, cerr.ErrorValidationFailed.Error("Name", "Resource name inconsistent between body and path.")
	}

	return common.K8sClient.CycloneV1alpha1().Resources(rs.Namespace).Update(rs)
}

// DeleteResource ...
func DeleteResource(ctx context.Context, name, namespace string) error {
	return common.K8sClient.CycloneV1alpha1().Resources(namespace).Delete(name, nil)
}
