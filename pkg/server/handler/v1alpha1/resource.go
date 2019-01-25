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
)

// CreateResource ...
func CreateResource(ctx context.Context, project, tenant string, rsc *v1alpha1.Resource) (*v1alpha1.Resource, error) {
	labels := rsc.ObjectMeta.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[common.LabelProject] = project

	rsc.Name = common.BuildResoucesName(project, rsc.Name)
	rsc.ObjectMeta.Labels = labels
	return handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Create(rsc)
}

// ListResources ...
func ListResources(ctx context.Context, project, tenant string, pagination *types.Pagination) (*types.ListResponse, error) {
	resources, err := handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).List(metav1.ListOptions{
		LabelSelector: common.ProjectSelector(project),
	})
	if err != nil {
		log.Errorf("Get resources from k8s with tenant %s, project %s error: %v", tenant, project, err)
		return nil, err
	}

	items := resources.Items
	size := int64(len(items))
	if pagination.Start >= size {
		return types.NewListResponse(int(size), []v1alpha1.Stage{}), nil
	}

	end := pagination.Start + pagination.Limit
	if end > size {
		end = size
	}

	rscs := make([]v1alpha1.Resource, size)
	for i, rsc := range items[pagination.Start:end] {
		rsc.Name = common.RetrieveResoucesName(project, rsc.Name)
		rscs[i] = rsc
	}
	return types.NewListResponse(int(size), rscs), nil
}

// GetResource ...
func GetResource(ctx context.Context, project, resource, tenant string) (*v1alpha1.Resource, error) {
	name := common.BuildResoucesName(project, resource)
	rsc, err := handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	rsc.Name = common.RetrieveResoucesName(project, rsc.Name)
	return rsc, nil
}

// UpdateResource ...
func UpdateResource(ctx context.Context, project, resource, tenant string, rsc *v1alpha1.Resource) (*v1alpha1.Resource, error) {
	name := common.BuildResoucesName(project, resource)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		origin, err := handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		newRsc := origin.DeepCopy()
		newRsc.Spec = rsc.Spec
		_, err = handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Update(newRsc)
		return err
	})

	if err != nil {
		return nil, err
	}

	return rsc, nil
}

// DeleteResource ...
func DeleteResource(ctx context.Context, project, resource, tenant string) error {
	name := common.BuildResoucesName(project, resource)
	return handler.K8sClient.CycloneV1alpha1().Resources(common.TenantNamespace(tenant)).Delete(name, nil)
}
