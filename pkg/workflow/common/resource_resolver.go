package common

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caicloud/cyclone/pkg/apis/cyclone/v1alpha1"
	"github.com/caicloud/cyclone/pkg/common"
	"github.com/caicloud/cyclone/pkg/k8s/clientset"
	"github.com/caicloud/cyclone/pkg/meta"
	"github.com/caicloud/cyclone/pkg/workflow/controller"
)

func GetResourceTypes(client clientset.Interface, namespaces []string, operation string) ([]v1alpha1.Resource, error) {
	var results []v1alpha1.Resource
	for _, ns := range namespaces {
		resources, err := client.CycloneV1alpha1().Resources(ns).List(metav1.ListOptions{
			LabelSelector: meta.ResourceTypeSelector(),
		})
		if err != nil {
			log.Errorf("Get resource type from namespace %s error: %v", ns, err)
			return nil, err
		}
		for _, item := range resources.Items {
			if operation == "" {
				results = append(results, item)
				continue
			}

			for _, op := range item.Spec.SupportedOperations {
				if strings.EqualFold(operation, op) {
					results = append(results, item)
					break
				}
			}
		}
	}

	return results, nil
}

// GetResourceResolver gets resource resolver for a given resource type
func GetResourceResolver(client clientset.Interface, resource *v1alpha1.Resource) (string, error) {
	if len(resource.Spec.Resolver) > 0 {
		return resource.Spec.Resolver, nil
	}

	// If resolver is set in config file, use it.
	resolverConfigKey := fmt.Sprintf("%s-resolver", strings.ToLower(resource.Spec.Type))
	if resolver, ok := controller.Config.Images[resolverConfigKey]; ok {
		return resolver, nil
	}

	namespaces := []string{resource.Namespace}
	systemNamespace := common.GetSystemNamespace()
	if resource.Namespace != systemNamespace {
		namespaces = append(namespaces, systemNamespace)
	}

	resources, err := GetResourceTypes(client, namespaces, "")
	if err != nil {
		return "", err
	}

	for _, r := range resources {
		if strings.EqualFold(r.Spec.Type, resource.Spec.Type) {
			return r.Spec.Resolver, nil
		}
	}

	return "", fmt.Errorf("resolver not found for '%s'", resource.Spec.Type)
}
