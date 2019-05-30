package v1alpha1

import (
	"context"
	"fmt"
	"strings"

	"github.com/caicloud/nirvana/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
			for _, op := range item.Spec.SupportedOperations {
				if operation == "" || strings.ToLower(operation) == op {
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
