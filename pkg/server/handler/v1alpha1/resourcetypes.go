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
	"github.com/caicloud/cyclone/pkg/meta"
	svrcommon "github.com/caicloud/cyclone/pkg/server/common"
	"github.com/caicloud/cyclone/pkg/server/handler"
	"github.com/caicloud/cyclone/pkg/server/types"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

func getResourceTypes(namespaces []string, operation string) ([]v1alpha1.Resource, error) {
	var results []v1alpha1.Resource
	for _, ns := range namespaces {
		resources, err := handler.K8sClient.CycloneV1alpha1().Resources(ns).List(metav1.ListOptions{
			LabelSelector: meta.ResourceTypeSelector(),
		})
		if err != nil {
			log.Errorf("Get resource type from namespace %s error: %v", ns, err)
			return nil, cerr.ConvertK8sError(err)
		}
		for _, item := range resources.Items {
			if operation == "" {
				results = append(results, item)
				continue
			}

			for _, op := range item.Spec.SupportedOperations {
				if strings.ToLower(operation) == strings.ToLower(op) {
					results = append(results, item)
					break
				}
			}
		}
	}

	return results, nil
}

// ListResourceTypes ...
func ListResourceTypes(ctx context.Context, tenant string, operation string) (*types.ListResponse, error) {
	namespaces := []string{svrcommon.TenantNamespace(tenant)}
	if tenant != svrcommon.DefaultTenant {
		namespaces = append(namespaces, common.GetSystemNamespace())
	}

	resources, err := getResourceTypes(namespaces, operation)
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
	resources, err := getResourceTypes(namespaces, "")
	if err != nil {
		log.Errorf("Get resource type from tenant namespaces %v error: %v", namespaces, err)
		return nil, cerr.ConvertK8sError(err)
	}

	for _, resource := range resources {
		if strings.ToLower(string(resource.Spec.Type)) == strings.ToLower(resourceType) {
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
	types, err := getResourceTypes([]string{svrcommon.TenantNamespace(tenant)}, "")
	if err != nil {
		return nil, err
	}
	for _, t := range types {
		if strings.ToLower(string(t.Spec.Type)) != strings.ToLower(resourceType) {
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
	types, err := getResourceTypes([]string{svrcommon.TenantNamespace(tenant)}, "")
	if err != nil {
		return err
	}

	for _, t := range types {
		if strings.ToLower(string(t.Spec.Type)) != strings.ToLower(resourceType) {
			continue
		}

		err := handler.K8sClient.CycloneV1alpha1().Resources(svrcommon.TenantNamespace(tenant)).Delete(t.Name, &metav1.DeleteOptions{})
		return cerr.ConvertK8sError(err)
	}

	return cerr.ErrorContentNotFound.Error(fmt.Sprintf("resource type '%s'", resourceType))
}
