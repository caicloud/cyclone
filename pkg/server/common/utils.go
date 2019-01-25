package common

import (
	"fmt"
	"strings"
)

// TenantNamespace gets namespace from given tenant name.
func TenantNamespace(tenant string) string {
	return fmt.Sprintf("cyclone--%s", tenant)
}

// TenantPVC returns pvc name related the tenant
func TenantPVC(tenant string) string {
	return TenantPVCPrefix + tenant
}

// TenantResourceQuota returns resource quota name related the tenant
func TenantResourceQuota(tenant string) string {
	return tenant
}

// LabelOwnerCyclone returns a label string describes resource belongs to cyclone
func LabelOwnerCyclone() string {
	return LabelOwner + "=" + OwnerCyclone
}

// IntegrationSecret returns secret name related to the integration
func IntegrationSecret(i string) string {
	return i
}

// WorkerClustersSelector is a selector for clusters which are use to perform workload
func WorkerClustersSelector() string {
	return LabelClusterOn + "=" + LabelClusterOnValue
}

// ProjectSelector is a selector for cyclone CRD resources which have corresponding project label
func ProjectSelector(project string) string {
	return LabelProject + "=" + project
}

// BuildResoucesName returns name, in k8s side, of resources under the project.
func BuildResoucesName(project string, name string) string {
	return project + "-" + name
}

// RetrieveResoucesName returns name, in user side, of resources under the project.
func RetrieveResoucesName(project string, name string) string {
	return strings.TrimPrefix(name, project+"-")
}

// AddProjectLabel adds project label for a resource metadata labels
func AddProjectLabel(rl map[string]string, project string) map[string]string {
	labels := rl
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[LabelProject] = project

	return labels
}
