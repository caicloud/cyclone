package common

import "fmt"

// TenantNamespace gets namespace from given tenant name.
func TenantNamespace(tenant string) string {
	return fmt.Sprintf("cyclone--%s", tenant)
}
