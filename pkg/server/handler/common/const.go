package common

const (
	// TeanntNamespacePrefix is the prefix of namespace which related to a specific tenant
	TeanntNamespacePrefix = "cyclone--"

	// DefaultTenant is a default tenant created while cyclone initialize
	DefaultTenant = "admin"

	// DefaultTenantNamespace is the namespace of the default tenant
	DefaultTenantNamespace = TeanntNamespacePrefix + DefaultTenant

	// TenantPVCPrefix is the prefix of pvc which related to a specific tenant
	TenantPVCPrefix = "cyclone-pvc-"

	//// DefaultTenantPVC is the pvc name of the default tenant
	//DefaultTenantPVC = TenantPVCPrefix + DefaultTenant

	// DefaultPVCSize is the default size of pvc
	DefaultPVCSize = "5Gi"

	// AnnotationTenant is the annotation key used for namespace to relate tenant information
	AnnotationTenant = "cyclone.io/tenant-info"

	// LabelOwner is the label key used to indicate namespaces created by cyclone
	LabelOwner = "cyclone.io/owner"

	// OwnerCyclone is the label value used to indicate namespaces created by cyclone
	OwnerCyclone = "cyclone"

	// QuotaCPULimit represents default value of 'limits.cpu'
	QuotaCPULimit = "2"
	// QuotaCPURequest represents default value of 'requests.cpu'
	QuotaCPURequest = "1"
	// QuotaMemoryLimit represents default value of 'limits.memory'
	QuotaMemoryLimit = "4Gi"
	// QuotaMemoryRequest represents default value of 'requests.memory'
	QuotaMemoryRequest = "2Gi"
)
