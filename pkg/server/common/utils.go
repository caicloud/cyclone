package common

import "fmt"

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
