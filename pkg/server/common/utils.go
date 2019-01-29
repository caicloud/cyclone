package common

import (
	"fmt"
	"strings"
)

// TenantNamespace gets namespace from given tenant name.
func TenantNamespace(tenant string) string {
	return fmt.Sprintf("%s%s", TenantNamespacePrefix, tenant)
}

// NamespaceTenant retrieves tenant from given namespace name.
func NamespaceTenant(n string) string {
	return strings.TrimPrefix(n, TenantNamespacePrefix)
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

// SecretIntegration returns integration name related to the secret
func SecretIntegration(s string) string {
	return s
}

// WorkerClustersSelector is a selector for clusters which are use to perform workload
func WorkerClustersSelector() string {
	return LabelClusterOn + "=" + LabelClusterOnValue
}

// ProjectSelector is a selector for cyclone CRD resources which have corresponding project label
func ProjectSelector(project string) string {
	return LabelProject + "=" + project
}
