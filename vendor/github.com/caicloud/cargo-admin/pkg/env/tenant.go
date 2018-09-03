package env

var SystemTenant = "system-tenant"

func IsSystemTenant(tid string) bool {
	return tid == SystemTenant
}
