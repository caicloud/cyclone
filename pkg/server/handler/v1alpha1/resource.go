package v1alpha1

import (
	"context"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

// CreateResource ...
func CreateResource(ctx context.Context, project, tenant string, rsc *v1alpha1.Resource) (*v1alpha1.Resource, error) {
	modifiers := []CreationModifier{GenerateNameModifier, InjectProjectLabelModifier}
	for _, modifier := range modifiers {
		err := modifier(project, tenant, rsc)
		if err != nil {
			return nil, err
		}
	}

	return handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Create(rsc)
}

// ListResources ...
func ListResources(ctx context.Context, project, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	resources, err := handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: common.ProjectSelector(project),
	})
	if err != nil {
		log.Errorf("Get resources from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, cerr.ConvertK8sError(err)
	}

	items := resources.Items
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Resource{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	return types.NewListResponse(int(size), items[pagination.Start:end]), nil
}

// GetResource ...
func GetResource(ctx context.Context, project, resource, tenant string) (*v1alpha1.Resource, error) {
	rsc, err := handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Get(resource, metav1.GetOptions{})

	return rsc, cerr.ConvertK8sError(err)
}

// UpdateResource ...
func UpdateResource(ctx context.Context, project, resource, tenant string, rsc *v1alpha1.Resource) (*v1alpha1.Resource, error) {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Get(resource, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newRsc := origin.DeepCopy()
		newRsc.Spec = rsc.Spec
		newRsc.Annotations = UpdateAnnotations(rsc.Annotations, newRsc.Annotations)
		_, err = handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Update(newRsc)
		return err
	})

	if err != nil {
		return nil, cerr.ConvertK8sError(err)
	}

	return rsc, nil
}

// DeleteResource ...
func DeleteResource(ctx context.Context, project, resource, tenant string) error {
	err := handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Delete(resource, nil)

	return cerr.ConvertK8sError(err)
}
