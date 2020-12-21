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

// TenantPVC returns pvc name related to the tenant
func TenantPVC(tenant string) string {
	return TenantPVCPrefix + tenant
}

// ResourceQuotaName returns resource quota name according to the namespace.
// Currently the name of resource quota is the same as namespace.
func ResourceQuotaName(namespace string) string {
	return namespace
}
