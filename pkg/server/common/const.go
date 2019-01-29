package common

const (
	// AdminTenant is name of the system admin tenant, it's a default tenant created when Cyclone
	// start, and resources shared among all tenants would be placed in this tenant, such as stage
	// templates.
	AdminTenant = "admin"

	// TenantPVCPrefix is the prefix of pvc which related to a specific tenant
	TenantPVCPrefix = "cyclone-pvc-"

	// DefaultPVCSize is the default size of pvc
	DefaultPVCSize = "5Gi"

	// AnnotationTenant is the annotation key used for namespace to relate tenant information
	AnnotationTenant = "cyclone.io/tenant-info"

	// LabelProject is the label key used to indicate the project which the resources belongs to
	LabelProject = "cyclone.io/project"

	// LabelIntegrationType is the label key used to indicate type of integration
	LabelIntegrationType = "cyclone.io/integration-type"

	// LabelClusterOn is the label key used to indicate the cluster is a worker for the tenant
	LabelClusterOn = "cyclone.io/cluster-worker"

	// LabelClusterOnValue is the label value used to indicate the cluster is a worker for the tenant
	LabelClusterOnValue = "true"

	// LabelOwner is the label key used to indicate namespaces created by cyclone
	LabelOwner = "cyclone.io/owner"

	// OwnerCyclone is the label value used to indicate namespaces created by cyclone
	OwnerCyclone = "cyclone"

	// AnnotationAlias is the annotation key used to indicate the alias of resources
	AnnotationAlias = "cyclone.io/alias"

	// AnnotationDescription is the annotation key used to describe resources
	AnnotationDescription = "cyclone.io/description"

	// QuotaCPULimit represents default value of 'limits.cpu'
	QuotaCPULimit = "2"
	// QuotaCPURequest represents default value of 'requests.cpu'
	QuotaCPURequest = "1"
	// QuotaMemoryLimit represents default value of 'limits.memory'
	QuotaMemoryLimit = "4Gi"
	// QuotaMemoryRequest represents default value of 'requests.memory'
	QuotaMemoryRequest = "2Gi"

	// SecretKeyIntegration is the key of the secret dada to indicate its value is about integration information.
	SecretKeyIntegration = "integration"

	// ControlClusterName is the name of control cluster
	ControlClusterName = "control-cluster"
)
