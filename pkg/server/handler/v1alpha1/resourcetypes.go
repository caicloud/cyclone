package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
	wfcommon "github.com/caicloud/cyclone/pkg/workflow/common"
)

// ListResourceTypes ...
func ListResourceTypes(ctx context.Context, tenant string, operation string) (*types.ListResponse, error) {
	namespaces := []string{svrcommon.TenantNamespace(tenant)}
	if tenant != svrcommon.DefaultTenant {
		namespaces = append(namespaces, common.GetSystemNamespace())
	}

	resources, err := wfcommon.GetResourceTypes(handler.K8sClient, namespaces, operation)
	if err != nil {
		log.Errorf("Get resource type from tenant namespaces %v error: %v", namespaces, err)
		return nil, cerr.ConvertK8sError(err)
	}

	return types.NewListResponse(len(resources), resources), nil
}

// GetResourceType ...
func GetResourceType(ctx context.Context, tenant, resourceType string) (*v1alpha1.Resource, error) {
	namespaces := []string{svrcommon.TenantNamespace(tenant)}
	if tenant != svrcommon.DefaultTenant {
		namespaces = append(namespaces, common.GetSystemNamespace())
	}
	resources, err := wfcommon.GetResourceTypes(handler.K8sClient, namespaces, "")
	if err != nil {
		log.Errorf("Get resource type from tenant namespaces %v error: %v", namespaces, err)
		return nil, cerr.ConvertK8sError(err)
	}

	for _, resource := range resources {
		if strings.EqualFold(resource.Spec.Type, resourceType) {
			return &resource, nil
		}
	}

	return nil, cerr.ErrorContentNotFound.Error(fmt.Sprintf("resource type '%s'", resourceType))
}

// CreateResourceType ...
func CreateResourceType(ctx context.Context, tenant string, resource *v1alpha1.Resource) (*v1alpha1.Resource, error) {
	rsc, err := handler.K8sClient.CycloneV1alpha1().Resources(svrcommon.TenantNamespace(tenant)).Create(resource)
	return rsc, cerr.ConvertK8sError(err)
}

// UpdateResourceType ...
func UpdateResourceType(ctx context.Context, tenant string, resourceType string, resource *v1alpha1.Resource) (*v1alpha1.Resource, error) {
	types, err := wfcommon.GetResourceTypes(handler.K8sClient, []string{svrcommon.TenantNamespace(tenant)}, "")
	if err != nil {
		return nil, err
	}
	for _, t := range types {
		if !strings.EqualFold(t.Spec.Type, resourceType) {
			continue
		}

		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			origin, err := handler.K8sClient.CycloneV1alpha1().Resources(svrcommon.TenantNamespace(tenant)).Get(t.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			origin.Spec = resource.Spec
			_, err = handler.K8sClient.CycloneV1alpha1().Resources(svrcommon.TenantNamespace(tenant)).Update(origin)
			return err
		})
		return resource, err
	}

	return nil, cerr.ErrorContentNotFound.Error(fmt.Sprintf("resource type '%s'", resourceType))
}

// DeleteResourceType ...
func DeleteResourceType(ctx context.Context, tenant string, resourceType string) error {
	types, err := wfcommon.GetResourceTypes(handler.K8sClient, []string{svrcommon.TenantNamespace(tenant)}, "")
	if err != nil {
		return err
	}

	for _, t := range types {
		if !strings.EqualFold(t.Spec.Type, resourceType) {
			continue
		}

		err := handler.K8sClient.CycloneV1alpha1().Resources(svrcommon.TenantNamespace(tenant)).Delete(t.Name, &metav1.DeleteOptions{})
		return cerr.ConvertK8sError(err)
	}

	return cerr.ErrorContentNotFound.Error(fmt.Sprintf("resource type '%s'", resourceType))
}
